package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// route is a request path template with one %s per ID segment, such as
// "/roadmaps/%s/lanes/%s". It is a distinct type so that a string literal
// converts to it implicitly while a runtime string needs an explicit
// conversion that stands out in review: caller input enters a path only as
// a pathID, through with.
type route string

// apiPath is a request path, with any query string, ready to append to the
// base URL. Every ID segment in it came through a pathID.
type apiPath string

// with validates ids in order and substitutes them into r. The first
// invalid ID fails the whole path, naming its field.
func (r route) with(ids ...pathID) (apiPath, error) {
	segs := make([]any, len(ids))
	for i, id := range ids {
		seg, err := id.segment()
		if err != nil {
			return "", err
		}
		segs[i] = seg
	}
	return apiPath(fmt.Sprintf(string(r), segs...)), nil
}

// verb is an HTTP method; callers pass the net/http Method* constants.
type verb string

// readOnly reports whether v leaves server state alone. Any other verb
// invalidates the read cache.
func (v verb) readOnly() bool {
	return v == http.MethodGet || v == http.MethodHead
}

// getAt GETs r with ids substituted, through the read cache.
func (c *Client) getAt(ctx context.Context, r route, ids ...pathID) (json.RawMessage, error) {
	p, err := r.with(ids...)
	if err != nil {
		return nil, err
	}
	return c.get(ctx, p)
}

// listAt fetches every page of the collection at r with ids substituted.
func (c *Client) listAt(ctx context.Context, r route, q Query, ids ...pathID) (json.RawMessage, error) {
	p, err := r.with(ids...)
	if err != nil {
		return nil, err
	}
	return c.getList(ctx, p, q)
}

// postAt POSTs body to r with ids substituted.
func (c *Client) postAt(ctx context.Context, r route, body any, ids ...pathID) (json.RawMessage, error) {
	p, err := r.with(ids...)
	if err != nil {
		return nil, err
	}
	return c.request(ctx, http.MethodPost, p, body)
}

// patchAt PATCHes r, with ids substituted, with body.
func (c *Client) patchAt(ctx context.Context, r route, body any, ids ...pathID) (json.RawMessage, error) {
	p, err := r.with(ids...)
	if err != nil {
		return nil, err
	}
	return c.request(ctx, http.MethodPatch, p, body)
}

// deleteAt DELETEs r with ids substituted.
func (c *Client) deleteAt(ctx context.Context, r route, ids ...pathID) (json.RawMessage, error) {
	p, err := r.with(ids...)
	if err != nil {
		return nil, err
	}
	return c.request(ctx, http.MethodDelete, p, nil)
}
