package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// Every endpoint method that interpolates a user-supplied ID into a URL path
// MUST run the ID through safeSeg first. safeSeg validates against the safe
// ID pattern and url.PathEscape-s the result, so an adversarial caller cannot
// pivot the request via "../" traversal, "?" query injection, or "/"
// sub-resource extension. See internal/api/client.go safeSeg.

// ============================================================================
// Roadmaps
// ============================================================================

// ListRoadmaps returns all roadmaps.
func (c *Client) ListRoadmaps(ctx context.Context) (json.RawMessage, error) {
	data, err := c.Get(ctx, "/roadmaps")
	if err != nil {
		return nil, err
	}
	return FormatRoadmapList(data), nil
}

// GetRoadmap returns a single roadmap by ID.
func (c *Client) GetRoadmap(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("roadmap_id", id)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/roadmaps/"+seg)
}

// GetRoadmapBars returns all bars for a roadmap, enriched with lane names.
func (c *Client) GetRoadmapBars(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("roadmap_id", id)
	if err != nil {
		return nil, err
	}
	bars, err := c.Get(ctx, "/roadmaps/"+seg+"/bars")
	if err != nil {
		return nil, err
	}
	lanes, _ := c.Get(ctx, "/roadmaps/"+seg+"/lanes")
	return FormatBarsWithContext(bars, lanes), nil
}

// GetRoadmapLanes returns all lanes for a roadmap.
func (c *Client) GetRoadmapLanes(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("roadmap_id", id)
	if err != nil {
		return nil, err
	}
	data, err := c.Get(ctx, "/roadmaps/"+seg+"/lanes")
	if err != nil {
		return nil, err
	}
	return FormatLanes(data), nil
}

// GetRoadmapMilestones returns all milestones for a roadmap.
func (c *Client) GetRoadmapMilestones(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("roadmap_id", id)
	if err != nil {
		return nil, err
	}
	data, err := c.Get(ctx, "/roadmaps/"+seg+"/milestones")
	if err != nil {
		return nil, err
	}
	return FormatMilestones(data), nil
}

// GetRoadmapLegends returns all legend entries (color codes) for a roadmap.
// Legends are embedded in the roadmap response; there is no separate /legends endpoint.
func (c *Client) GetRoadmapLegends(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("roadmap_id", id)
	if err != nil {
		return nil, err
	}
	data, err := c.Get(ctx, "/roadmaps/"+seg)
	if err != nil {
		return nil, err
	}
	var roadmap map[string]json.RawMessage
	if err := json.Unmarshal(data, &roadmap); err != nil {
		return nil, fmt.Errorf("failed to parse roadmap response: %w", err)
	}
	legends, ok := roadmap["legends"]
	if !ok {
		return FormatLegends(json.RawMessage("[]")), nil
	}
	return FormatLegends(legends), nil
}

// GetRoadmapComments returns all comments on a roadmap.
func (c *Client) GetRoadmapComments(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("roadmap_id", id)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/roadmaps/"+seg+"/comments")
}

// ============================================================================
// Lanes
// ============================================================================

// CreateLane creates a new lane.
func (c *Client) CreateLane(ctx context.Context, roadmapID string, data map[string]any) (json.RawMessage, error) {
	seg, err := safeSeg("roadmap_id", roadmapID)
	if err != nil {
		return nil, err
	}
	return c.Post(ctx, "/roadmaps/"+seg+"/lanes", data)
}

// UpdateLane updates an existing lane.
func (c *Client) UpdateLane(ctx context.Context, roadmapID, laneID string, data map[string]any) (json.RawMessage, error) {
	rSeg, lSeg, err := safeSegPair("roadmap_id", roadmapID, "lane_id", laneID)
	if err != nil {
		return nil, err
	}
	return c.Patch(ctx, "/roadmaps/"+rSeg+"/lanes/"+lSeg, data)
}

// DeleteLane deletes a lane.
func (c *Client) DeleteLane(ctx context.Context, roadmapID, laneID string) (json.RawMessage, error) {
	rSeg, lSeg, err := safeSegPair("roadmap_id", roadmapID, "lane_id", laneID)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, "/roadmaps/"+rSeg+"/lanes/"+lSeg)
}

// ============================================================================
// Milestones
// ============================================================================

// CreateMilestone creates a new milestone.
func (c *Client) CreateMilestone(ctx context.Context, roadmapID string, data map[string]any) (json.RawMessage, error) {
	seg, err := safeSeg("roadmap_id", roadmapID)
	if err != nil {
		return nil, err
	}
	return c.Post(ctx, "/roadmaps/"+seg+"/milestones", data)
}

// UpdateMilestone updates an existing milestone.
func (c *Client) UpdateMilestone(ctx context.Context, roadmapID, milestoneID string, data map[string]any) (json.RawMessage, error) {
	rSeg, mSeg, err := safeSegPair("roadmap_id", roadmapID, "milestone_id", milestoneID)
	if err != nil {
		return nil, err
	}
	return c.Patch(ctx, "/roadmaps/"+rSeg+"/milestones/"+mSeg, data)
}

// DeleteMilestone deletes a milestone.
func (c *Client) DeleteMilestone(ctx context.Context, roadmapID, milestoneID string) (json.RawMessage, error) {
	rSeg, mSeg, err := safeSegPair("roadmap_id", roadmapID, "milestone_id", milestoneID)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, "/roadmaps/"+rSeg+"/milestones/"+mSeg)
}

// ============================================================================
// Admin
// ============================================================================

// ListUsers returns all users.
func (c *Client) ListUsers(ctx context.Context) (json.RawMessage, error) {
	return c.Get(ctx, "/users")
}

// ListTeams returns all teams.
func (c *Client) ListTeams(ctx context.Context) (json.RawMessage, error) {
	return c.Get(ctx, "/teams")
}

// CheckStatus checks the API status.
func (c *Client) CheckStatus(ctx context.Context) (json.RawMessage, error) {
	return c.Get(ctx, "/status")
}
