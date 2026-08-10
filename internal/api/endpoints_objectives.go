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
	data, err := c.Get(ctx, "/strategy/objectives")
	if err != nil {
		return nil, err
	}
	return FormatObjectives(data), nil
}

// GetObjective returns a single objective by ID.
func (c *Client) GetObjective(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("objective_id", id)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/strategy/objectives/"+seg)
}

// CreateObjective creates a new objective.
func (c *Client) CreateObjective(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/strategy/objectives", data)
}

// UpdateObjective updates an existing objective.
func (c *Client) UpdateObjective(ctx context.Context, id string, data map[string]any) (json.RawMessage, error) {
	seg, err := safeSeg("objective_id", id)
	if err != nil {
		return nil, err
	}
	return c.Patch(ctx, "/strategy/objectives/"+seg, data)
}

// DeleteObjective deletes an objective.
func (c *Client) DeleteObjective(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("objective_id", id)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, "/strategy/objectives/"+seg)
}

// ============================================================================
// Key Results
// ============================================================================

// ListKeyResults returns key results for an objective.
func (c *Client) ListKeyResults(ctx context.Context, objectiveID string) (json.RawMessage, error) {
	seg, err := safeSeg("objective_id", objectiveID)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/strategy/objectives/"+seg+"/key_results")
}

// GetKeyResult returns a single key result by ID.
func (c *Client) GetKeyResult(ctx context.Context, objectiveID, keyResultID string) (json.RawMessage, error) {
	oSeg, krSeg, err := safeSegPair("objective_id", objectiveID, "key_result_id", keyResultID)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/strategy/objectives/"+oSeg+"/key_results/"+krSeg)
}

// CreateKeyResult creates a new key result.
func (c *Client) CreateKeyResult(ctx context.Context, objectiveID string, data map[string]any) (json.RawMessage, error) {
	seg, err := safeSeg("objective_id", objectiveID)
	if err != nil {
		return nil, err
	}
	return c.Post(ctx, "/strategy/objectives/"+seg+"/key_results", data)
}

// UpdateKeyResult updates an existing key result.
func (c *Client) UpdateKeyResult(ctx context.Context, objectiveID, keyResultID string, data map[string]any) (json.RawMessage, error) {
	oSeg, krSeg, err := safeSegPair("objective_id", objectiveID, "key_result_id", keyResultID)
	if err != nil {
		return nil, err
	}
	return c.Patch(ctx, "/strategy/objectives/"+oSeg+"/key_results/"+krSeg, data)
}

// DeleteKeyResult deletes a key result.
func (c *Client) DeleteKeyResult(ctx context.Context, objectiveID, keyResultID string) (json.RawMessage, error) {
	oSeg, krSeg, err := safeSegPair("objective_id", objectiveID, "key_result_id", keyResultID)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, "/strategy/objectives/"+oSeg+"/key_results/"+krSeg)
}
