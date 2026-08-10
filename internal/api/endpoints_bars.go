package api

import (
	"context"
	"encoding/json"
)

// ============================================================================
// Bars
// ============================================================================

// GetBar returns a single bar by ID.
func (c *Client) GetBar(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("bar_id", id)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/bars/"+seg)
}

// CreateBar creates a new bar.
func (c *Client) CreateBar(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/bars", data)
}

// UpdateBar updates an existing bar.
func (c *Client) UpdateBar(ctx context.Context, id string, data map[string]any) (json.RawMessage, error) {
	seg, err := safeSeg("bar_id", id)
	if err != nil {
		return nil, err
	}
	return c.Patch(ctx, "/bars/"+seg, data)
}

// DeleteBar deletes a bar.
func (c *Client) DeleteBar(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("bar_id", id)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, "/bars/"+seg)
}

// GetBarChildren returns child bars for a bar.
func (c *Client) GetBarChildren(ctx context.Context, barID string) (json.RawMessage, error) {
	seg, err := safeSeg("bar_id", barID)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/bars/"+seg+"/child_bars")
}

// ============================================================================
// Bar Comments
// ============================================================================

// GetBarComments returns comments for a bar.
func (c *Client) GetBarComments(ctx context.Context, barID string) (json.RawMessage, error) {
	seg, err := safeSeg("bar_id", barID)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/bars/"+seg+"/comments")
}

// ============================================================================
// Bar Connections (dependencies)
// ============================================================================

// GetBarConnections returns connections for a bar.
func (c *Client) GetBarConnections(ctx context.Context, barID string) (json.RawMessage, error) {
	seg, err := safeSeg("bar_id", barID)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/bars/"+seg+"/connections")
}

// CreateBarConnection creates a connection from a bar.
func (c *Client) CreateBarConnection(ctx context.Context, barID string, data map[string]any) (json.RawMessage, error) {
	seg, err := safeSeg("bar_id", barID)
	if err != nil {
		return nil, err
	}
	return c.Post(ctx, "/bars/"+seg+"/connections", data)
}

// DeleteBarConnection deletes a connection.
func (c *Client) DeleteBarConnection(ctx context.Context, barID, connectionID string) (json.RawMessage, error) {
	barSeg, connSeg, err := safeSegPair("bar_id", barID, "connection_id", connectionID)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, "/bars/"+barSeg+"/connections/"+connSeg)
}

// ============================================================================
// Bar Links (external URLs)
// ============================================================================

// GetBarLinks returns links for a bar.
func (c *Client) GetBarLinks(ctx context.Context, barID string) (json.RawMessage, error) {
	seg, err := safeSeg("bar_id", barID)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/bars/"+seg+"/links")
}

// CreateBarLink creates a link on a bar.
func (c *Client) CreateBarLink(ctx context.Context, barID string, data map[string]any) (json.RawMessage, error) {
	seg, err := safeSeg("bar_id", barID)
	if err != nil {
		return nil, err
	}
	return c.Post(ctx, "/bars/"+seg+"/links", data)
}

// DeleteBarLink deletes a link.
func (c *Client) DeleteBarLink(ctx context.Context, barID, linkID string) (json.RawMessage, error) {
	barSeg, linkSeg, err := safeSegPair("bar_id", barID, "link_id", linkID)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, "/bars/"+barSeg+"/links/"+linkSeg)
}
