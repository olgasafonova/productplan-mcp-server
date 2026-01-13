package api

import (
	"context"
	"encoding/json"
	"fmt"
)

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
	return c.Get(ctx, "/roadmaps/"+id)
}

// GetRoadmapBars returns all bars for a roadmap, enriched with lane names.
func (c *Client) GetRoadmapBars(ctx context.Context, id string) (json.RawMessage, error) {
	bars, err := c.Get(ctx, "/roadmaps/"+id+"/bars")
	if err != nil {
		return nil, err
	}
	lanes, _ := c.Get(ctx, "/roadmaps/"+id+"/lanes")
	return FormatBarsWithContext(bars, lanes), nil
}

// GetRoadmapLanes returns all lanes for a roadmap.
func (c *Client) GetRoadmapLanes(ctx context.Context, id string) (json.RawMessage, error) {
	data, err := c.Get(ctx, "/roadmaps/"+id+"/lanes")
	if err != nil {
		return nil, err
	}
	return FormatLanes(data), nil
}

// GetRoadmapMilestones returns all milestones for a roadmap.
func (c *Client) GetRoadmapMilestones(ctx context.Context, id string) (json.RawMessage, error) {
	data, err := c.Get(ctx, "/roadmaps/"+id+"/milestones")
	if err != nil {
		return nil, err
	}
	return FormatMilestones(data), nil
}

// GetRoadmapLegends returns all legend entries (color codes) for a roadmap.
func (c *Client) GetRoadmapLegends(ctx context.Context, id string) (json.RawMessage, error) {
	data, err := c.Get(ctx, "/roadmaps/"+id+"/legends")
	if err != nil {
		return nil, err
	}
	return FormatLegends(data), nil
}

// GetRoadmapComments returns all comments on a roadmap.
func (c *Client) GetRoadmapComments(ctx context.Context, id string) (json.RawMessage, error) {
	return c.Get(ctx, "/roadmaps/"+id+"/comments")
}

// ============================================================================
// Bars
// ============================================================================

// GetBar returns a single bar by ID.
func (c *Client) GetBar(ctx context.Context, id string) (json.RawMessage, error) {
	return c.Get(ctx, "/bars/"+id)
}

// CreateBar creates a new bar.
func (c *Client) CreateBar(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/bars", data)
}

// UpdateBar updates an existing bar.
func (c *Client) UpdateBar(ctx context.Context, id string, data map[string]any) (json.RawMessage, error) {
	return c.Patch(ctx, "/bars/"+id, data)
}

// DeleteBar deletes a bar.
func (c *Client) DeleteBar(ctx context.Context, id string) (json.RawMessage, error) {
	return c.Delete(ctx, "/bars/"+id)
}

// GetBarChildren returns child bars for a bar.
func (c *Client) GetBarChildren(ctx context.Context, barID string) (json.RawMessage, error) {
	return c.Get(ctx, "/bars/"+barID+"/child_bars")
}

// ============================================================================
// Bar Comments
// ============================================================================

// GetBarComments returns comments for a bar.
func (c *Client) GetBarComments(ctx context.Context, barID string) (json.RawMessage, error) {
	return c.Get(ctx, "/bars/"+barID+"/comments")
}

// CreateBarComment creates a comment on a bar.
func (c *Client) CreateBarComment(ctx context.Context, barID string, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/bars/"+barID+"/comments", data)
}

// ============================================================================
// Bar Connections (dependencies)
// ============================================================================

// GetBarConnections returns connections for a bar.
func (c *Client) GetBarConnections(ctx context.Context, barID string) (json.RawMessage, error) {
	return c.Get(ctx, "/bars/"+barID+"/connections")
}

// CreateBarConnection creates a connection from a bar.
func (c *Client) CreateBarConnection(ctx context.Context, barID string, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/bars/"+barID+"/connections", data)
}

// DeleteBarConnection deletes a connection.
func (c *Client) DeleteBarConnection(ctx context.Context, barID, connectionID string) (json.RawMessage, error) {
	return c.Delete(ctx, fmt.Sprintf("/bars/%s/connections/%s", barID, connectionID))
}

// ============================================================================
// Bar Links (external URLs)
// ============================================================================

// GetBarLinks returns links for a bar.
func (c *Client) GetBarLinks(ctx context.Context, barID string) (json.RawMessage, error) {
	return c.Get(ctx, "/bars/"+barID+"/links")
}

// CreateBarLink creates a link on a bar.
func (c *Client) CreateBarLink(ctx context.Context, barID string, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/bars/"+barID+"/links", data)
}

// UpdateBarLink updates a link.
func (c *Client) UpdateBarLink(ctx context.Context, barID, linkID string, data map[string]any) (json.RawMessage, error) {
	return c.Patch(ctx, fmt.Sprintf("/bars/%s/links/%s", barID, linkID), data)
}

// DeleteBarLink deletes a link.
func (c *Client) DeleteBarLink(ctx context.Context, barID, linkID string) (json.RawMessage, error) {
	return c.Delete(ctx, fmt.Sprintf("/bars/%s/links/%s", barID, linkID))
}

// ============================================================================
// Lanes
// ============================================================================

// CreateLane creates a new lane.
func (c *Client) CreateLane(ctx context.Context, roadmapID string, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/roadmaps/"+roadmapID+"/lanes", data)
}

// UpdateLane updates an existing lane.
func (c *Client) UpdateLane(ctx context.Context, roadmapID, laneID string, data map[string]any) (json.RawMessage, error) {
	return c.Patch(ctx, fmt.Sprintf("/roadmaps/%s/lanes/%s", roadmapID, laneID), data)
}

// DeleteLane deletes a lane.
func (c *Client) DeleteLane(ctx context.Context, roadmapID, laneID string) (json.RawMessage, error) {
	return c.Delete(ctx, fmt.Sprintf("/roadmaps/%s/lanes/%s", roadmapID, laneID))
}

// ============================================================================
// Milestones
// ============================================================================

// CreateMilestone creates a new milestone.
func (c *Client) CreateMilestone(ctx context.Context, roadmapID string, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/roadmaps/"+roadmapID+"/milestones", data)
}

// UpdateMilestone updates an existing milestone.
func (c *Client) UpdateMilestone(ctx context.Context, roadmapID, milestoneID string, data map[string]any) (json.RawMessage, error) {
	return c.Patch(ctx, fmt.Sprintf("/roadmaps/%s/milestones/%s", roadmapID, milestoneID), data)
}

// DeleteMilestone deletes a milestone.
func (c *Client) DeleteMilestone(ctx context.Context, roadmapID, milestoneID string) (json.RawMessage, error) {
	return c.Delete(ctx, fmt.Sprintf("/roadmaps/%s/milestones/%s", roadmapID, milestoneID))
}

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
	return c.Get(ctx, "/strategy/objectives/"+id)
}

// CreateObjective creates a new objective.
func (c *Client) CreateObjective(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/strategy/objectives", data)
}

// UpdateObjective updates an existing objective.
func (c *Client) UpdateObjective(ctx context.Context, id string, data map[string]any) (json.RawMessage, error) {
	return c.Patch(ctx, "/strategy/objectives/"+id, data)
}

// DeleteObjective deletes an objective.
func (c *Client) DeleteObjective(ctx context.Context, id string) (json.RawMessage, error) {
	return c.Delete(ctx, "/strategy/objectives/"+id)
}

// ============================================================================
// Key Results
// ============================================================================

// ListKeyResults returns key results for an objective.
func (c *Client) ListKeyResults(ctx context.Context, objectiveID string) (json.RawMessage, error) {
	return c.Get(ctx, "/strategy/objectives/"+objectiveID+"/key_results")
}

// GetKeyResult returns a single key result by ID.
func (c *Client) GetKeyResult(ctx context.Context, objectiveID, keyResultID string) (json.RawMessage, error) {
	return c.Get(ctx, fmt.Sprintf("/strategy/objectives/%s/key_results/%s", objectiveID, keyResultID))
}

// CreateKeyResult creates a new key result.
func (c *Client) CreateKeyResult(ctx context.Context, objectiveID string, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/strategy/objectives/"+objectiveID+"/key_results", data)
}

// UpdateKeyResult updates an existing key result.
func (c *Client) UpdateKeyResult(ctx context.Context, objectiveID, keyResultID string, data map[string]any) (json.RawMessage, error) {
	return c.Patch(ctx, fmt.Sprintf("/strategy/objectives/%s/key_results/%s", objectiveID, keyResultID), data)
}

// DeleteKeyResult deletes a key result.
func (c *Client) DeleteKeyResult(ctx context.Context, objectiveID, keyResultID string) (json.RawMessage, error) {
	return c.Delete(ctx, fmt.Sprintf("/strategy/objectives/%s/key_results/%s", objectiveID, keyResultID))
}

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
	return c.Get(ctx, "/discovery/ideas/"+id)
}

// CreateIdea creates a new idea.
func (c *Client) CreateIdea(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/discovery/ideas", data)
}

// UpdateIdea updates an existing idea.
func (c *Client) UpdateIdea(ctx context.Context, id string, data map[string]any) (json.RawMessage, error) {
	return c.Patch(ctx, "/discovery/ideas/"+id, data)
}

// ============================================================================
// Idea Customers
// ============================================================================

// ListAllCustomers returns all customers across all ideas.
func (c *Client) ListAllCustomers(ctx context.Context) (json.RawMessage, error) {
	return c.Get(ctx, "/discovery/ideas/customers")
}

// GetIdeaCustomers returns customers for an idea.
func (c *Client) GetIdeaCustomers(ctx context.Context, ideaID string) (json.RawMessage, error) {
	return c.Get(ctx, "/discovery/ideas/"+ideaID+"/customers")
}

// AddIdeaCustomer adds a customer to an idea.
func (c *Client) AddIdeaCustomer(ctx context.Context, ideaID string, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/discovery/ideas/"+ideaID+"/customers", data)
}

// RemoveIdeaCustomer removes a customer from an idea.
func (c *Client) RemoveIdeaCustomer(ctx context.Context, ideaID, customerID string) (json.RawMessage, error) {
	return c.Delete(ctx, fmt.Sprintf("/discovery/ideas/%s/customers/%s", ideaID, customerID))
}

// ============================================================================
// Idea Tags
// ============================================================================

// ListAllTags returns all tags across all ideas.
func (c *Client) ListAllTags(ctx context.Context) (json.RawMessage, error) {
	return c.Get(ctx, "/discovery/ideas/tags")
}

// GetIdeaTags returns tags for an idea.
func (c *Client) GetIdeaTags(ctx context.Context, ideaID string) (json.RawMessage, error) {
	return c.Get(ctx, "/discovery/ideas/"+ideaID+"/tags")
}

// AddIdeaTag adds a tag to an idea.
func (c *Client) AddIdeaTag(ctx context.Context, ideaID string, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/discovery/ideas/"+ideaID+"/tags", data)
}

// RemoveIdeaTag removes a tag from an idea.
func (c *Client) RemoveIdeaTag(ctx context.Context, ideaID, tagID string) (json.RawMessage, error) {
	return c.Delete(ctx, fmt.Sprintf("/discovery/ideas/%s/tags/%s", ideaID, tagID))
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
	return c.Get(ctx, "/discovery/opportunities/"+id)
}

// CreateOpportunity creates a new opportunity.
func (c *Client) CreateOpportunity(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/discovery/opportunities", data)
}

// UpdateOpportunity updates an existing opportunity.
func (c *Client) UpdateOpportunity(ctx context.Context, id string, data map[string]any) (json.RawMessage, error) {
	return c.Patch(ctx, "/discovery/opportunities/"+id, data)
}

// DeleteOpportunity deletes an opportunity.
func (c *Client) DeleteOpportunity(ctx context.Context, id string) (json.RawMessage, error) {
	return c.Delete(ctx, "/discovery/opportunities/"+id)
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
	return c.Get(ctx, "/discovery/idea_forms/"+id)
}

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
	return c.Get(ctx, "/launches/"+id)
}

// CreateLaunch creates a new launch.
func (c *Client) CreateLaunch(ctx context.Context, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/launches", data)
}

// UpdateLaunch updates an existing launch.
func (c *Client) UpdateLaunch(ctx context.Context, id string, data map[string]any) (json.RawMessage, error) {
	return c.Patch(ctx, "/launches/"+id, data)
}

// DeleteLaunch deletes a launch.
func (c *Client) DeleteLaunch(ctx context.Context, id string) (json.RawMessage, error) {
	return c.Delete(ctx, "/launches/"+id)
}

// ============================================================================
// Launch Checklist Sections
// ============================================================================

// GetLaunchSections returns all checklist sections for a launch.
func (c *Client) GetLaunchSections(ctx context.Context, launchID string) (json.RawMessage, error) {
	return c.Get(ctx, "/launches/"+launchID+"/checklist_sections")
}

// CreateLaunchSection creates a new checklist section.
func (c *Client) CreateLaunchSection(ctx context.Context, launchID string, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/launches/"+launchID+"/checklist_sections", data)
}

// UpdateLaunchSection updates an existing checklist section.
func (c *Client) UpdateLaunchSection(ctx context.Context, launchID, sectionID string, data map[string]any) (json.RawMessage, error) {
	return c.Patch(ctx, fmt.Sprintf("/launches/%s/checklist_sections/%s", launchID, sectionID), data)
}

// DeleteLaunchSection deletes a checklist section.
func (c *Client) DeleteLaunchSection(ctx context.Context, launchID, sectionID string) (json.RawMessage, error) {
	return c.Delete(ctx, fmt.Sprintf("/launches/%s/checklist_sections/%s", launchID, sectionID))
}

// ============================================================================
// Launch Tasks
// ============================================================================

// GetLaunchTasks returns all tasks for a launch.
func (c *Client) GetLaunchTasks(ctx context.Context, launchID string) (json.RawMessage, error) {
	return c.Get(ctx, "/launches/"+launchID+"/tasks")
}

// CreateLaunchTask creates a new task in a launch.
func (c *Client) CreateLaunchTask(ctx context.Context, launchID string, data map[string]any) (json.RawMessage, error) {
	return c.Post(ctx, "/launches/"+launchID+"/tasks", data)
}

// UpdateLaunchTask updates an existing task.
func (c *Client) UpdateLaunchTask(ctx context.Context, launchID, taskID string, data map[string]any) (json.RawMessage, error) {
	return c.Patch(ctx, fmt.Sprintf("/launches/%s/tasks/%s", launchID, taskID), data)
}

// DeleteLaunchTask deletes a task.
func (c *Client) DeleteLaunchTask(ctx context.Context, launchID, taskID string) (json.RawMessage, error) {
	return c.Delete(ctx, fmt.Sprintf("/launches/%s/tasks/%s", launchID, taskID))
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
