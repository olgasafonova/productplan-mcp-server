package api

import (
	"context"
	"encoding/json"
)

// ============================================================================
// Bars
// ============================================================================

// GetBar returns a single bar by ID.
func (c *Client) GetBar(ctx context.Context, id BarID) (json.RawMessage, error) {
	return c.getAt(ctx, "/bars/%s", id)
}

// CreateBar creates a new bar.
func (c *Client) CreateBar(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.postAt(ctx, "/bars", data)
}

// UpdateBar updates an existing bar.
func (c *Client) UpdateBar(ctx context.Context, id BarID, data map[string]any) (json.RawMessage, error) {
	return c.patchAt(ctx, "/bars/%s", data, id)
}

// DeleteBar deletes a bar.
func (c *Client) DeleteBar(ctx context.Context, id BarID) (json.RawMessage, error) {
	return c.deleteAt(ctx, "/bars/%s", id)
}

// GetBarChildren returns child bars for a bar.
func (c *Client) GetBarChildren(ctx context.Context, barID BarID) (json.RawMessage, error) {
	return c.listAt(ctx, "/bars/%s/child_bars", Query{}, barID)
}

// ============================================================================
// Bar Comments
// ============================================================================

// GetBarComments returns comments for a bar.
func (c *Client) GetBarComments(ctx context.Context, barID BarID) (json.RawMessage, error) {
	return c.listAt(ctx, "/bars/%s/comments", Query{}, barID)
}

// ============================================================================
// Bar Connections (dependencies)
// ============================================================================

// GetBarConnections returns connections for a bar. The body is the
// {requires, required_by} shape, not a paged list, so it is a plain GET.
func (c *Client) GetBarConnections(ctx context.Context, barID BarID) (json.RawMessage, error) {
	return c.getAt(ctx, "/bars/%s/connections", barID)
}

// CreateBarConnection creates a connection from a bar.
func (c *Client) CreateBarConnection(ctx context.Context, barID BarID, data map[string]any) (json.RawMessage, error) {
	return c.postAt(ctx, "/bars/%s/connections", data, barID)
}

// DeleteBarConnection deletes a connection.
func (c *Client) DeleteBarConnection(ctx context.Context, barID BarID, connectionID ConnectionID) (json.RawMessage, error) {
	return c.deleteAt(ctx, "/bars/%s/connections/%s", barID, connectionID)
}

// ============================================================================
// Bar Links (external URLs)
// ============================================================================

// GetBarLinks returns links for a bar.
func (c *Client) GetBarLinks(ctx context.Context, barID BarID) (json.RawMessage, error) {
	return c.listAt(ctx, "/bars/%s/links", Query{}, barID)
}

// CreateBarLink creates a link on a bar.
func (c *Client) CreateBarLink(ctx context.Context, barID BarID, data map[string]any) (json.RawMessage, error) {
	return c.postAt(ctx, "/bars/%s/links", data, barID)
}

// DeleteBarLink deletes a link.
func (c *Client) DeleteBarLink(ctx context.Context, barID BarID, linkID LinkID) (json.RawMessage, error) {
	return c.deleteAt(ctx, "/bars/%s/links/%s", barID, linkID)
}
