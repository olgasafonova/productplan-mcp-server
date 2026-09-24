package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
	"github.com/olgasafonova/productplan-mcp-server/pkg/productplan"
)

// bulkConcurrency bounds in-flight API calls per bulk tool call. Every call
// also passes through the client's adaptive rate limiter, which slows all
// workers down as the rate window drains.
const bulkConcurrency = 4

// bulkItemResult is the per-item outcome of a bulk call.
type bulkItemResult struct {
	Index   int            `json:"index"`
	BarID   string         `json:"bar_id,omitempty"`
	Name    string         `json:"name,omitempty"`
	OK      bool           `json:"ok"`
	Error   string         `json:"error,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

// runBulk runs fn for each index with at most bulkConcurrency in flight,
// reusing productplan.Execute, and returns results in index order. fn
// reports failure inside its result, so batch errors only arise when the
// context is cancelled before an item starts.
func runBulk(ctx context.Context, n int, fn func(ctx context.Context, i int) bulkItemResult) []bulkItemResult {
	fns := make([]func(context.Context) (bulkItemResult, error), n)
	for i := range n {
		fns[i] = func(ctx context.Context) (bulkItemResult, error) {
			r := fn(ctx, i)
			r.Index = i
			return r, nil
		}
	}
	batch := productplan.Execute(ctx, productplan.BatchConfig{Concurrency: bulkConcurrency}, fns)
	out := make([]bulkItemResult, n)
	for _, r := range batch.Results {
		out[r.Index] = r
	}
	for _, e := range batch.Errors {
		out[e.Index] = bulkItemResult{Index: e.Index, Error: "not attempted: " + e.Err.Error()}
	}
	return out
}

// prepareAll validates every op before anything is written. Any invalid
// item fails the whole call with every problem listed.
func prepareAll(ctx context.Context, pl *barPlanner, ops []barOp) ([]bulkItemResult, error) {
	results := runBulk(ctx, len(ops), func(ctx context.Context, i int) bulkItemResult {
		r := bulkItemResult{BarID: ops[i].barID, Name: ops[i].fields.Name}
		p, err := pl.prepare(ctx, ops[i])
		if err != nil {
			r.Error = err.Error()
			return r
		}
		r.OK, r.Payload = true, p
		return r
	})
	var problems []string
	for _, r := range results {
		if !r.OK {
			problems = append(problems, fmt.Sprintf("items[%d]%s: %s", r.Index, barLabel(r.BarID), r.Error))
		}
	}
	if len(problems) > 0 {
		return nil, fmt.Errorf("validation failed for %d of %d items; nothing was written. Fix these and resend the call:\n%s",
			len(problems), len(ops), strings.Join(problems, "\n"))
	}
	return results, nil
}

func barLabel(id string) string {
	if id == "" {
		return ""
	}
	return " (bar " + id + ")"
}

// bulkResponse shapes the outcome. A call where nothing succeeded is an
// error result (nothing changed, so a corrected retry is safe); a partial
// failure is a normal result whose summary names the failures, because
// some writes landed and a blind retry of a create would duplicate them.
func bulkResponse(results []bulkItemResult, verb, noun string, dryRun bool) (json.RawMessage, error) {
	succeeded := 0
	for _, r := range results {
		if r.OK {
			succeeded++
		}
	}
	n, failed := len(results), len(results)-succeeded
	summary := fmt.Sprintf("%s %d of %d %s", verb, succeeded, n, ItemType(noun).plural(n))
	if dryRun {
		summary = fmt.Sprintf("Dry run: %d of %d %s valid; nothing was sent. Each payload is exactly what would be sent", succeeded, n, ItemType(noun).plural(n))
	}
	if failed > 0 {
		summary += fmt.Sprintf("; %d failed (see results[].error and retry only those items)", failed)
	}
	data, err := json.Marshal(map[string]any{"dry_run": dryRun, "succeeded": succeeded, "failed": failed, "results": results})
	if err != nil {
		return nil, err
	}
	if succeeded == 0 && !dryRun {
		return nil, fmt.Errorf("%s. Nothing was changed. %s", summary, data)
	}
	return json.Marshal(FormattedResponse{Summary: summary, Data: data})
}

func bulkUpdateBarsHandler(client *api.Client) mcp.Handler {
	return typedHandler[BulkUpdateBarsArgs](func(ctx context.Context, a BulkUpdateBarsArgs) (json.RawMessage, error) {
		ops := make([]barOp, len(a.Items))
		for i, it := range a.Items {
			ops[i] = barOp{barID: it.BarID, roadmapID: firstString(it.RoadmapID, a.RoadmapID), fields: mergeBarFields(a.Set, it.BarFields)}
		}
		prepared, err := prepareAll(ctx, newBarPlanner(client), ops)
		if err != nil || a.DryRun {
			return dryRunOrError(prepared, err, "bar")
		}
		results := runBulk(ctx, len(prepared), func(ctx context.Context, i int) bulkItemResult {
			pr := prepared[i]
			r := bulkItemResult{BarID: pr.BarID, OK: true}
			if _, werr := client.UpdateBar(ctx, pr.BarID, pr.Payload); werr != nil {
				r.OK, r.Error = false, explainWriteError(werr, pr.Payload).Error()
			}
			return r
		})
		return bulkResponse(results, "Updated", "bar", false)
	})
}

func bulkCreateBarsHandler(client *api.Client) mcp.Handler {
	return typedHandler[BulkCreateBarsArgs](func(ctx context.Context, a BulkCreateBarsArgs) (json.RawMessage, error) {
		ops := make([]barOp, len(a.Items))
		for i, it := range a.Items {
			ops[i] = barOp{create: true, roadmapID: a.RoadmapID, fields: mergeBarFields(a.Set, it)}
		}
		prepared, err := prepareAll(ctx, newBarPlanner(client), ops)
		if err != nil || a.DryRun {
			return dryRunOrError(prepared, err, "bar")
		}
		results := runBulk(ctx, len(prepared), func(ctx context.Context, i int) bulkItemResult {
			return createOne(ctx, client, prepared[i])
		})
		return bulkResponse(results, "Created", "bar", false)
	})
}

// createOne posts one prepared bar. A create whose ID cannot be parsed
// still counts as written: the bar exists, and retrying would duplicate it.
func createOne(ctx context.Context, client *api.Client, pr bulkItemResult) bulkItemResult {
	r := bulkItemResult{Name: pr.Name}
	data, err := client.CreateBar(ctx, pr.Payload)
	if err != nil {
		r.Error = explainWriteError(err, pr.Payload).Error()
		return r
	}
	r.OK = true
	if r.BarID, err = api.ParseCreatedID(data); err != nil {
		r.Error = fmt.Sprintf("created, but the new ID could not be read (%v); find it with get_roadmap_bars and do not retry this item", err)
	}
	return r
}

// dryRunOrError returns the validation error, or the validated payloads
// for a dry run.
func dryRunOrError(prepared []bulkItemResult, err error, noun string) (json.RawMessage, error) {
	if err != nil {
		return nil, err
	}
	return bulkResponse(prepared, "", noun, true)
}

func bulkDeleteBarsHandler(client *api.Client) mcp.Handler {
	return typedHandler[BulkDeleteBarsArgs](func(ctx context.Context, a BulkDeleteBarsArgs) (json.RawMessage, error) {
		if a.DryRun {
			results := runBulk(ctx, len(a.BarIDs), func(ctx context.Context, i int) bulkItemResult {
				r := bulkItemResult{BarID: a.BarIDs[i]}
				bar, err := client.GetBarSummary(ctx, a.BarIDs[i])
				if err != nil {
					r.Error = err.Error()
					return r
				}
				r.OK, r.Name = true, bar.Name
				return r
			})
			return bulkResponse(results, "", "bar", true)
		}
		results := runBulk(ctx, len(a.BarIDs), func(ctx context.Context, i int) bulkItemResult {
			return deleteAndVerify(ctx, client, a.BarIDs[i])
		})
		return bulkResponse(results, "Deleted", "bar", false)
	})
}

// deleteAndVerify deletes a bar and confirms with a read-back that it is
// gone (GET answers 404). ProductPlan delete receipts have claimed success
// for deletes that did not happen, so the 204 alone is not trusted.
func deleteAndVerify(ctx context.Context, client *api.Client, barID string) bulkItemResult {
	r := bulkItemResult{BarID: barID}
	if _, err := client.DeleteBar(ctx, barID); err != nil {
		if api.IsNotFound(err) {
			r.Error = "bar not found (already deleted, or the ID is wrong)"
		} else {
			r.Error = err.Error()
		}
		return r
	}
	_, err := client.GetBar(ctx, barID)
	switch {
	case api.IsNotFound(err):
		r.OK = true
	case err == nil:
		r.Error = "DELETE answered success but the bar still exists on read-back; it was NOT deleted"
	default:
		r.Error = "DELETE answered success but the read-back failed, so the deletion is unverified: " + err.Error()
	}
	return r
}
