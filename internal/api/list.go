package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/olgasafonova/productplan-mcp-server/internal/logging"
)

const (
	// listPageSize is the page size requested from collection endpoints.
	// The API documents a default of 200 and a maximum of 500; asking for
	// the maximum keeps the number of round trips (and rate-limit spend) low.
	listPageSize = 500

	// maxListPages bounds how many pages one list call will fetch
	// (50 x 500 = 25,000 records). Anything beyond is reported as
	// incomplete rather than fetched, so a runaway collection cannot turn
	// one tool call into hundreds of requests.
	maxListPages = 50

	// pageFetchConcurrency bounds parallel page fetches after page 1.
	pageFetchConcurrency = 4
)

// Query is a Ransack-style filter for ProductPlan collection endpoints.
// Predicates maps an attribute+predicate compound (for example "name_i_cont"
// or "starts_on_gteq") to its value; Sort is "field asc" or "field desc".
// Both are encoded in bracket notation: q[name_i_cont]=Ship&q[s]=name+asc.
//
// Validation of which attributes and predicates a tool may send lives in the
// tools layer (internal/tools/filters.go); this type only encodes.
type Query struct {
	Predicates map[string]string
	Sort       string
}

// IsZero reports whether the query carries no filter and no sort.
func (q Query) IsZero() bool { return len(q.Predicates) == 0 && q.Sort == "" }

// encode adds q[...] parameters to v in a deterministic order, so the same
// logical query always yields the same URL (and therefore the same cache key).
func (q Query) encode(v url.Values) {
	keys := make([]string, 0, len(q.Predicates))
	for k := range q.Predicates {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v.Set("q["+k+"]", q.Predicates[k])
	}
	if q.Sort != "" {
		v.Set("q[s]", q.Sort)
	}
}

// listURL builds endpoint?page=N&page_size=500&q[...]. url.Values.Encode
// sorts keys, so the result is stable for a given query and page.
func listURL(endpoint apiPath, q Query, page int) apiPath {
	v := url.Values{}
	v.Set("page_size", strconv.Itoa(listPageSize))
	if page > 1 {
		v.Set("page", strconv.Itoa(page))
	}
	q.encode(v)
	return endpoint + "?" + apiPath(v.Encode())
}

// paging mirrors the API's paging block. PagesFetched is ours: it is set on
// merged responses so formatters can tell a complete merge from a capped one.
type paging struct {
	RecordCount  int `json:"record_count"`
	PageCount    int `json:"page_count"`
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size"`
	PagesFetched int `json:"pages_fetched,omitempty"`
}

// incomplete reports whether a merged response stopped before the last page.
func (p *paging) incomplete() bool {
	return p != nil && p.PagesFetched > 0 && p.PagesFetched < p.PageCount
}

type pagedEnvelope struct {
	Results []json.RawMessage `json:"results"`
	Paging  *paging           `json:"paging"`
}

// parseEnvelope decodes a {"results":[...],"paging":{...}} body. ok is false
// for bare arrays, single objects, and anything else.
func parseEnvelope(data json.RawMessage) (pagedEnvelope, bool) {
	var env pagedEnvelope
	trimmed := strings.TrimSpace(string(data))
	if !strings.HasPrefix(trimmed, "{") {
		return env, false
	}
	if err := json.Unmarshal(data, &env); err != nil || env.Results == nil {
		return env, false
	}
	return env, true
}

// getList fetches a collection endpoint, following paging.page_count so the
// caller never receives a silent partial page. Pages 2..N are fetched with
// bounded concurrency. If any page fails the whole call fails: returning a
// partial merge as if it were complete is the exact failure this exists to
// prevent. Past maxListPages the merge stops, is logged, and is marked with
// paging.pages_fetched < paging.page_count for the formatters to report.
//
// Bodies that are not a paged envelope (bare arrays, the connections
// {requires, required_by} shape) are returned unchanged.
func (c *Client) getList(ctx context.Context, endpoint apiPath, q Query) (json.RawMessage, error) {
	first, err := c.get(ctx, listURL(endpoint, q, 1))
	if err != nil {
		return nil, err
	}
	env, ok := parseEnvelope(first)
	if !ok || !env.Paging.multiPage() {
		return first, nil
	}

	lf := listFetch{client: c, endpoint: endpoint, query: q, first: env}
	pages, err := lf.pages(ctx)
	if err != nil {
		c.logger.Error("paginated list fetch failed", logging.Endpoint(string(endpoint)), logging.Error(err))
		return nil, fmt.Errorf("list %s: %w", endpoint, err)
	}
	lf.warnIfCapped()
	return json.Marshal(lf.merge(pages))
}

// multiPage reports whether the collection spans more than one page.
func (p *paging) multiPage() bool {
	return p != nil && p.PageCount > 1
}

// listFetch is one multi-page GetList in progress, seeded with page 1.
type listFetch struct {
	client   *Client
	endpoint apiPath
	query    Query
	first    pagedEnvelope
}

// fetchCount is how many pages will be fetched: all, up to maxListPages.
func (lf listFetch) fetchCount() int {
	return min(lf.first.Paging.PageCount, maxListPages)
}

// pages fetches pages 2..fetchCount with bounded concurrency and returns
// every page's results in order. Any page failing fails the whole fetch.
func (lf listFetch) pages(ctx context.Context) ([][]json.RawMessage, error) {
	pageCount := lf.first.Paging.PageCount
	pages := make([][]json.RawMessage, lf.fetchCount())
	pages[0] = lf.first.Results

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(pageFetchConcurrency)
	for p := 2; p <= len(pages); p++ {
		g.Go(func() error {
			data, err := lf.client.get(gctx, listURL(lf.endpoint, lf.query, p))
			if err != nil {
				return fmt.Errorf("page %d of %d: %w", p, pageCount, err)
			}
			pe, ok := parseEnvelope(data)
			if !ok {
				return fmt.Errorf("page %d of %d: response is not a paged results envelope", p, pageCount)
			}
			pages[p-1] = pe.Results
			return nil
		})
	}
	return pages, g.Wait()
}

// warnIfCapped logs when maxListPages stopped the fetch early.
func (lf listFetch) warnIfCapped() {
	pg := lf.first.Paging
	if lf.fetchCount() >= pg.PageCount {
		return
	}
	lf.client.logger.Warn("paginated list capped",
		logging.Endpoint(string(lf.endpoint)),
		logging.F("page_count", pg.PageCount),
		logging.F("pages_fetched", lf.fetchCount()),
		logging.F("record_count", pg.RecordCount),
	)
}

// merge concatenates the pages into one envelope whose paging records how
// many pages were fetched.
func (lf listFetch) merge(pages [][]json.RawMessage) pagedEnvelope {
	var all []json.RawMessage
	for _, page := range pages {
		all = append(all, page...)
	}
	pg := lf.first.Paging
	return pagedEnvelope{
		Results: all,
		Paging: &paging{
			RecordCount:  pg.RecordCount,
			PageCount:    pg.PageCount,
			PageSize:     pg.PageSize,
			PagesFetched: lf.fetchCount(),
		},
	}
}
