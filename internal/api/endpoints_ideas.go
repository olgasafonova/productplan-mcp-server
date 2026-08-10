package api

import (
	"context"
	"encoding/json"
)

// ============================================================================
// Ideas
// ============================================================================

// ListIdeas returns all ideas.
func (c *Client) ListIdeas(ctx context.Context) (json.RawMessage, error) {
	data, err := c.Get(ctx, "/discovery/ideas")
	if err != nil {
		return nil, err
	}
	return FormatIdeas(data), nil
}

// GetIdea returns a single idea by ID.
func (c *Client) GetIdea(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("idea_id", id)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/discovery/ideas/"+seg)
}

// CreateIdea creates a new idea.
func (c *Client) CreateIdea(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/discovery/ideas", data)
}

// UpdateIdea updates an existing idea.
func (c *Client) UpdateIdea(ctx context.Context, id string, data map[string]any) (json.RawMessage, error) {
	seg, err := safeSeg("idea_id", id)
	if err != nil {
		return nil, err
	}
	return c.Patch(ctx, "/discovery/ideas/"+seg, data)
}

// ============================================================================
// Idea Customers
// ============================================================================

// ListAllCustomers returns all customers across all ideas.
func (c *Client) ListAllCustomers(ctx context.Context) (json.RawMessage, error) {
	return c.Get(ctx, "/discovery/ideas/customers")
}

// ============================================================================
// Idea Tags
// ============================================================================

// ListAllTags returns all tags across all ideas.
func (c *Client) ListAllTags(ctx context.Context) (json.RawMessage, error) {
	return c.Get(ctx, "/discovery/ideas/tags")
}

// ============================================================================
// Opportunities
// ============================================================================

// ListOpportunities returns all opportunities.
func (c *Client) ListOpportunities(ctx context.Context) (json.RawMessage, error) {
	data, err := c.Get(ctx, "/discovery/opportunities")
	if err != nil {
		return nil, err
	}
	return FormatOpportunities(data), nil
}

// GetOpportunity returns a single opportunity by ID.
func (c *Client) GetOpportunity(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("opportunity_id", id)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/discovery/opportunities/"+seg)
}

// CreateOpportunity creates a new opportunity.
func (c *Client) CreateOpportunity(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/discovery/opportunities", data)
}

// UpdateOpportunity updates an existing opportunity.
func (c *Client) UpdateOpportunity(ctx context.Context, id string, data map[string]any) (json.RawMessage, error) {
	seg, err := safeSeg("opportunity_id", id)
	if err != nil {
		return nil, err
	}
	return c.Patch(ctx, "/discovery/opportunities/"+seg, data)
}

// ============================================================================
// Idea Forms
// ============================================================================

// ListIdeaForms returns all idea forms.
func (c *Client) ListIdeaForms(ctx context.Context) (json.RawMessage, error) {
	return c.Get(ctx, "/discovery/idea_forms")
}

// GetIdeaForm returns a single idea form by ID.
func (c *Client) GetIdeaForm(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("idea_form_id", id)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/discovery/idea_forms/"+seg)
}
