package api

import (
	"context"
	"encoding/json"
)

// ============================================================================
// Objectives (OKRs)
// ============================================================================

// ListObjectives returns all objectives.
func (c *Client) ListObjectives(ctx context.Context) (json.RawMessage, error) {
	data, err := c.listAt(ctx, "/strategy/objectives", Query{})
	if err != nil {
		return nil, err
	}
	return FormatObjectives(data), nil
}

// GetObjective returns a single objective by ID.
func (c *Client) GetObjective(ctx context.Context, id ObjectiveID) (json.RawMessage, error) {
	return c.getAt(ctx, "/strategy/objectives/%s", id)
}

// CreateObjective creates a new objective.
func (c *Client) CreateObjective(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.postAt(ctx, "/strategy/objectives", data)
}

// UpdateObjective updates an existing objective.
func (c *Client) UpdateObjective(ctx context.Context, id ObjectiveID, data map[string]any) (json.RawMessage, error) {
	return c.patchAt(ctx, "/strategy/objectives/%s", data, id)
}

// DeleteObjective deletes an objective.
func (c *Client) DeleteObjective(ctx context.Context, id ObjectiveID) (json.RawMessage, error) {
	return c.deleteAt(ctx, "/strategy/objectives/%s", id)
}

// ============================================================================
// Key Results
// ============================================================================

// ListKeyResults returns key results for an objective.
func (c *Client) ListKeyResults(ctx context.Context, objectiveID ObjectiveID) (json.RawMessage, error) {
	return c.listAt(ctx, "/strategy/objectives/%s/key_results", Query{}, objectiveID)
}

// GetKeyResult returns a single key result by ID.
func (c *Client) GetKeyResult(ctx context.Context, objectiveID ObjectiveID, keyResultID KeyResultID) (json.RawMessage, error) {
	return c.getAt(ctx, "/strategy/objectives/%s/key_results/%s", objectiveID, keyResultID)
}

// CreateKeyResult creates a new key result.
func (c *Client) CreateKeyResult(ctx context.Context, objectiveID ObjectiveID, data map[string]any) (json.RawMessage, error) {
	return c.postAt(ctx, "/strategy/objectives/%s/key_results", data, objectiveID)
}

// UpdateKeyResult updates an existing key result.
func (c *Client) UpdateKeyResult(ctx context.Context, objectiveID ObjectiveID, keyResultID KeyResultID, data map[string]any) (json.RawMessage, error) {
	return c.patchAt(ctx, "/strategy/objectives/%s/key_results/%s", data, objectiveID, keyResultID)
}

// DeleteKeyResult deletes a key result.
func (c *Client) DeleteKeyResult(ctx context.Context, objectiveID ObjectiveID, keyResultID KeyResultID) (json.RawMessage, error) {
	return c.deleteAt(ctx, "/strategy/objectives/%s/key_results/%s", objectiveID, keyResultID)
}
