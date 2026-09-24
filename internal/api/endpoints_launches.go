package api

import (
	"context"
	"encoding/json"
)

// ============================================================================
// Launches
// ============================================================================

// ListLaunches returns all launches.
func (c *Client) ListLaunches(ctx context.Context) (json.RawMessage, error) {
	return c.ListLaunchesWhere(ctx, Query{})
}

// ListLaunchesWhere returns the launches matching q (filters and sort).
func (c *Client) ListLaunchesWhere(ctx context.Context, q Query) (json.RawMessage, error) {
	data, err := c.listAt(ctx, "/launches", q)
	if err != nil {
		return nil, err
	}
	return FormatLaunches(data), nil
}

// GetLaunch returns a single launch by ID.
func (c *Client) GetLaunch(ctx context.Context, id LaunchID) (json.RawMessage, error) {
	return c.getAt(ctx, "/launches/%s", id)
}

// CreateLaunch creates a new launch.
func (c *Client) CreateLaunch(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.postAt(ctx, "/launches", data)
}

// UpdateLaunch updates an existing launch.
func (c *Client) UpdateLaunch(ctx context.Context, id LaunchID, data map[string]any) (json.RawMessage, error) {
	return c.patchAt(ctx, "/launches/%s", data, id)
}

// DeleteLaunch deletes a launch.
func (c *Client) DeleteLaunch(ctx context.Context, id LaunchID) (json.RawMessage, error) {
	return c.deleteAt(ctx, "/launches/%s", id)
}

// ============================================================================
// Launch Checklist Sections
// ============================================================================

// GetLaunchSections returns all checklist sections for a launch.
func (c *Client) GetLaunchSections(ctx context.Context, launchID LaunchID) (json.RawMessage, error) {
	return c.listAt(ctx, "/launches/%s/checklist_sections", Query{}, launchID)
}

// GetLaunchSection returns a single checklist section by ID.
func (c *Client) GetLaunchSection(ctx context.Context, launchID LaunchID, sectionID SectionID) (json.RawMessage, error) {
	return c.getAt(ctx, "/launches/%s/checklist_sections/%s", launchID, sectionID)
}

// CreateLaunchSection creates a new checklist section.
func (c *Client) CreateLaunchSection(ctx context.Context, launchID LaunchID, data map[string]any) (json.RawMessage, error) {
	return c.postAt(ctx, "/launches/%s/checklist_sections", data, launchID)
}

// UpdateLaunchSection updates an existing checklist section.
func (c *Client) UpdateLaunchSection(ctx context.Context, launchID LaunchID, sectionID SectionID, data map[string]any) (json.RawMessage, error) {
	return c.patchAt(ctx, "/launches/%s/checklist_sections/%s", data, launchID, sectionID)
}

// DeleteLaunchSection deletes a checklist section.
func (c *Client) DeleteLaunchSection(ctx context.Context, launchID LaunchID, sectionID SectionID) (json.RawMessage, error) {
	return c.deleteAt(ctx, "/launches/%s/checklist_sections/%s", launchID, sectionID)
}

// ============================================================================
// Launch Tasks
// ============================================================================

// GetLaunchTasks returns all tasks for a launch.
func (c *Client) GetLaunchTasks(ctx context.Context, launchID LaunchID) (json.RawMessage, error) {
	return c.listAt(ctx, "/launches/%s/tasks", Query{}, launchID)
}

// GetLaunchTask returns a single task by ID.
func (c *Client) GetLaunchTask(ctx context.Context, launchID LaunchID, taskID TaskID) (json.RawMessage, error) {
	return c.getAt(ctx, "/launches/%s/tasks/%s", launchID, taskID)
}

// CreateLaunchTask creates a new task in a launch.
func (c *Client) CreateLaunchTask(ctx context.Context, launchID LaunchID, data map[string]any) (json.RawMessage, error) {
	return c.postAt(ctx, "/launches/%s/tasks", data, launchID)
}

// UpdateLaunchTask updates an existing task.
func (c *Client) UpdateLaunchTask(ctx context.Context, launchID LaunchID, taskID TaskID, data map[string]any) (json.RawMessage, error) {
	return c.patchAt(ctx, "/launches/%s/tasks/%s", data, launchID, taskID)
}

// DeleteLaunchTask deletes a task.
func (c *Client) DeleteLaunchTask(ctx context.Context, launchID LaunchID, taskID TaskID) (json.RawMessage, error) {
	return c.deleteAt(ctx, "/launches/%s/tasks/%s", launchID, taskID)
}
