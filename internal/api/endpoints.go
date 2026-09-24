package api

import (
	"context"
	"encoding/json"
	"net/http"
)

// Every endpoint method that puts a caller-supplied ID into a URL takes a
// typed ID (ids.go) and builds its path with route.with, which validates and
// escapes each ID (safeSeg) before interpolation. An adversarial caller
// therefore cannot pivot the request via "../" traversal, "?" query
// injection, or "/" sub-resource extension, and a raw string cannot be
// passed where an ID is expected.

// ============================================================================
// Roadmaps
// ============================================================================

// ListRoadmaps returns all roadmaps.
func (c *Client) ListRoadmaps(ctx context.Context) (json.RawMessage, error) {
	return c.ListRoadmapsWhere(ctx, Query{})
}

// ListRoadmapsWhere returns the roadmaps matching q (filters and sort).
func (c *Client) ListRoadmapsWhere(ctx context.Context, q Query) (json.RawMessage, error) {
	data, err := c.listAt(ctx, "/roadmaps", q)
	if err != nil {
		return nil, err
	}
	return FormatRoadmapList(data), nil
}

// GetRoadmap returns a single roadmap by ID.
func (c *Client) GetRoadmap(ctx context.Context, id RoadmapID) (json.RawMessage, error) {
	return c.getAt(ctx, "/roadmaps/%s", id)
}

// GetRoadmapLanes returns all lanes for a roadmap.
func (c *Client) GetRoadmapLanes(ctx context.Context, id RoadmapID) (json.RawMessage, error) {
	return c.roadmapChildList(ctx, roadmapChild{id: id, route: "/roadmaps/%s/lanes", format: FormatLanes})
}

// GetRoadmapMilestones returns all milestones for a roadmap.
func (c *Client) GetRoadmapMilestones(ctx context.Context, id RoadmapID) (json.RawMessage, error) {
	return c.roadmapChildList(ctx, roadmapChild{id: id, route: "/roadmaps/%s/milestones", format: FormatMilestones})
}

// roadmapChild names a collection nested under a roadmap and the
// projection applied to it.
type roadmapChild struct {
	id     RoadmapID
	route  route
	format func(json.RawMessage) json.RawMessage
}

// roadmapChildList fetches every page of a roadmap's child collection and
// projects it.
func (c *Client) roadmapChildList(ctx context.Context, child roadmapChild) (json.RawMessage, error) {
	data, err := c.listAt(ctx, child.route, Query{}, child.id)
	if err != nil {
		return nil, err
	}
	return child.format(data), nil
}

// GetRoadmapLegends returns the legend names (bar colors) for a roadmap as a
// JSON array of strings. Legends are embedded in the roadmap response as bare
// names; there is no separate /legends endpoint and no legend ID or hex color
// in the API. A bar's color is set by sending one of these names as `legend`.
func (c *Client) GetRoadmapLegends(ctx context.Context, id RoadmapID) (json.RawMessage, error) {
	schema, err := c.GetBarWriteSchema(ctx, id)
	if err != nil {
		return nil, err
	}
	return json.Marshal(schema.Legends)
}

// GetRoadmapComments returns all comments on a roadmap.
func (c *Client) GetRoadmapComments(ctx context.Context, id RoadmapID) (json.RawMessage, error) {
	return c.listAt(ctx, "/roadmaps/%s/comments", Query{}, id)
}

// ============================================================================
// Lanes
// ============================================================================

// CreateLane creates a new lane.
func (c *Client) CreateLane(ctx context.Context, roadmapID RoadmapID, data map[string]any) (json.RawMessage, error) {
	return c.postAt(ctx, "/roadmaps/%s/lanes", data, roadmapID)
}

// UpdateLane updates an existing lane.
func (c *Client) UpdateLane(ctx context.Context, roadmapID RoadmapID, laneID LaneID, data map[string]any) (json.RawMessage, error) {
	return c.patchAt(ctx, "/roadmaps/%s/lanes/%s", data, roadmapID, laneID)
}

// DeleteLane deletes a lane.
func (c *Client) DeleteLane(ctx context.Context, roadmapID RoadmapID, laneID LaneID) (json.RawMessage, error) {
	return c.deleteAt(ctx, "/roadmaps/%s/lanes/%s", roadmapID, laneID)
}

// ============================================================================
// Milestones
// ============================================================================

// CreateMilestone creates a new milestone.
func (c *Client) CreateMilestone(ctx context.Context, roadmapID RoadmapID, data map[string]any) (json.RawMessage, error) {
	return c.postAt(ctx, "/roadmaps/%s/milestones", data, roadmapID)
}

// UpdateMilestone updates an existing milestone.
func (c *Client) UpdateMilestone(ctx context.Context, roadmapID RoadmapID, milestoneID MilestoneID, data map[string]any) (json.RawMessage, error) {
	return c.patchAt(ctx, "/roadmaps/%s/milestones/%s", data, roadmapID, milestoneID)
}

// DeleteMilestone deletes a milestone.
func (c *Client) DeleteMilestone(ctx context.Context, roadmapID RoadmapID, milestoneID MilestoneID) (json.RawMessage, error) {
	return c.deleteAt(ctx, "/roadmaps/%s/milestones/%s", roadmapID, milestoneID)
}

// ============================================================================
// Admin
// ============================================================================

// ListUsers returns all users.
func (c *Client) ListUsers(ctx context.Context) (json.RawMessage, error) {
	return c.listAt(ctx, "/users", Query{})
}

// ListTeams returns all teams.
func (c *Client) ListTeams(ctx context.Context) (json.RawMessage, error) {
	return c.listAt(ctx, "/teams", Query{})
}

// CheckStatus checks the API status.
func (c *Client) CheckStatus(ctx context.Context) (json.RawMessage, error) {
	// Bypass the read cache: a connectivity probe answered from memory
	// would report "up" for up to a TTL after the API went down.
	return c.request(ctx, http.MethodGet, "/status", nil)
}
