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
	data, err := c.Get(ctx, "/launches")
	if err != nil {
		return nil, err
	}
	return FormatLaunches(data), nil
}

// GetLaunch returns a single launch by ID.
func (c *Client) GetLaunch(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("launch_id", id)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/launches/"+seg)
}

// CreateLaunch creates a new launch.
func (c *Client) CreateLaunch(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/launches", data)
}

// UpdateLaunch updates an existing launch.
func (c *Client) UpdateLaunch(ctx context.Context, id string, data map[string]any) (json.RawMessage, error) {
	seg, err := safeSeg("launch_id", id)
	if err != nil {
		return nil, err
	}
	return c.Patch(ctx, "/launches/"+seg, data)
}

// DeleteLaunch deletes a launch.
func (c *Client) DeleteLaunch(ctx context.Context, id string) (json.RawMessage, error) {
	seg, err := safeSeg("launch_id", id)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, "/launches/"+seg)
}

// ============================================================================
// Launch Checklist Sections
// ============================================================================

// GetLaunchSections returns all checklist sections for a launch.
func (c *Client) GetLaunchSections(ctx context.Context, launchID string) (json.RawMessage, error) {
	seg, err := safeSeg("launch_id", launchID)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/launches/"+seg+"/checklist_sections")
}

// GetLaunchSection returns a single checklist section by ID.
func (c *Client) GetLaunchSection(ctx context.Context, launchID, sectionID string) (json.RawMessage, error) {
	lSeg, sSeg, err := safeSegPair("launch_id", launchID, "section_id", sectionID)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/launches/"+lSeg+"/checklist_sections/"+sSeg)
}

// CreateLaunchSection creates a new checklist section.
func (c *Client) CreateLaunchSection(ctx context.Context, launchID string, data map[string]any) (json.RawMessage, error) {
	seg, err := safeSeg("launch_id", launchID)
	if err != nil {
		return nil, err
	}
	return c.Post(ctx, "/launches/"+seg+"/checklist_sections", data)
}

// UpdateLaunchSection updates an existing checklist section.
func (c *Client) UpdateLaunchSection(ctx context.Context, launchID, sectionID string, data map[string]any) (json.RawMessage, error) {
	lSeg, sSeg, err := safeSegPair("launch_id", launchID, "section_id", sectionID)
	if err != nil {
		return nil, err
	}
	return c.Patch(ctx, "/launches/"+lSeg+"/checklist_sections/"+sSeg, data)
}

// DeleteLaunchSection deletes a checklist section.
func (c *Client) DeleteLaunchSection(ctx context.Context, launchID, sectionID string) (json.RawMessage, error) {
	lSeg, sSeg, err := safeSegPair("launch_id", launchID, "section_id", sectionID)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, "/launches/"+lSeg+"/checklist_sections/"+sSeg)
}

// ============================================================================
// Launch Tasks
// ============================================================================

// GetLaunchTasks returns all tasks for a launch.
func (c *Client) GetLaunchTasks(ctx context.Context, launchID string) (json.RawMessage, error) {
	seg, err := safeSeg("launch_id", launchID)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/launches/"+seg+"/tasks")
}

// GetLaunchTask returns a single task by ID.
func (c *Client) GetLaunchTask(ctx context.Context, launchID, taskID string) (json.RawMessage, error) {
	lSeg, tSeg, err := safeSegPair("launch_id", launchID, "task_id", taskID)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, "/launches/"+lSeg+"/tasks/"+tSeg)
}

// CreateLaunchTask creates a new task in a launch.
func (c *Client) CreateLaunchTask(ctx context.Context, launchID string, data map[string]any) (json.RawMessage, error) {
	seg, err := safeSeg("launch_id", launchID)
	if err != nil {
		return nil, err
	}
	return c.Post(ctx, "/launches/"+seg+"/tasks", data)
}

// UpdateLaunchTask updates an existing task.
func (c *Client) UpdateLaunchTask(ctx context.Context, launchID, taskID string, data map[string]any) (json.RawMessage, error) {
	lSeg, tSeg, err := safeSegPair("launch_id", launchID, "task_id", taskID)
	if err != nil {
		return nil, err
	}
	return c.Patch(ctx, "/launches/"+lSeg+"/tasks/"+tSeg, data)
}

// DeleteLaunchTask deletes a task.
func (c *Client) DeleteLaunchTask(ctx context.Context, launchID, taskID string) (json.RawMessage, error) {
	lSeg, tSeg, err := safeSegPair("launch_id", launchID, "task_id", taskID)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, "/launches/"+lSeg+"/tasks/"+tSeg)
}
