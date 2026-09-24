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
	return c.ListIdeasWhere(ctx, Query{})
}

// ListIdeasWhere returns the ideas matching q (filters and sort).
func (c *Client) ListIdeasWhere(ctx context.Context, q Query) (json.RawMessage, error) {
	data, err := c.listAt(ctx, "/discovery/ideas", q)
	if err != nil {
		return nil, err
	}
	return FormatIdeas(data), nil
}

// GetIdea returns a single idea by ID.
func (c *Client) GetIdea(ctx context.Context, id IdeaID) (json.RawMessage, error) {
	return c.getAt(ctx, "/discovery/ideas/%s", id)
}

// CreateIdea creates a new idea.
func (c *Client) CreateIdea(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.postAt(ctx, "/discovery/ideas", data)
}

// UpdateIdea updates an existing idea.
func (c *Client) UpdateIdea(ctx context.Context, id IdeaID, data map[string]any) (json.RawMessage, error) {
	return c.patchAt(ctx, "/discovery/ideas/%s", data, id)
}

// ============================================================================
// Idea Customers
// ============================================================================

// ListAllCustomers returns all customers across all ideas.
func (c *Client) ListAllCustomers(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/discovery/ideas/customers")
}

// ============================================================================
// Idea Tags
// ============================================================================

// ListAllTags returns all tags across all ideas.
func (c *Client) ListAllTags(ctx context.Context) (json.RawMessage, error) {
	return c.get(ctx, "/discovery/ideas/tags")
}

// ============================================================================
// Opportunities
// ============================================================================

// ListOpportunities returns all opportunities.
func (c *Client) ListOpportunities(ctx context.Context) (json.RawMessage, error) {
	return c.ListOpportunitiesWhere(ctx, Query{})
}

// ListOpportunitiesWhere returns the opportunities matching q (filters and sort).
func (c *Client) ListOpportunitiesWhere(ctx context.Context, q Query) (json.RawMessage, error) {
	data, err := c.listAt(ctx, "/discovery/opportunities", q)
	if err != nil {
		return nil, err
	}
	return FormatOpportunities(data), nil
}

// GetOpportunity returns a single opportunity by ID.
func (c *Client) GetOpportunity(ctx context.Context, id OpportunityID) (json.RawMessage, error) {
	return c.getAt(ctx, "/discovery/opportunities/%s", id)
}

// CreateOpportunity creates a new opportunity.
func (c *Client) CreateOpportunity(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.postAt(ctx, "/discovery/opportunities", data)
}

// UpdateOpportunity updates an existing opportunity.
func (c *Client) UpdateOpportunity(ctx context.Context, id OpportunityID, data map[string]any) (json.RawMessage, error) {
	return c.patchAt(ctx, "/discovery/opportunities/%s", data, id)
}

// ============================================================================
// Idea Forms
// ============================================================================

// ListIdeaForms returns all idea forms.
func (c *Client) ListIdeaForms(ctx context.Context) (json.RawMessage, error) {
	return c.listAt(ctx, "/discovery/idea_forms", Query{})
}

// GetIdeaForm returns a single idea form by ID.
func (c *Client) GetIdeaForm(ctx context.Context, id IdeaFormID) (json.RawMessage, error) {
	return c.getAt(ctx, "/discovery/idea_forms/%s", id)
}
