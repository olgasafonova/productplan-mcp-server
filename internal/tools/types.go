// Package tools provides typed argument structs for ProductPlan MCP tool handlers.
package tools

import (
	"encoding/json"
	"fmt"
)

// ParseArgs unmarshals map[string]any into a typed struct.
func ParseArgs[T any](args map[string]any) (T, error) {
	var result T
	data, err := json.Marshal(args)
	if err != nil {
		return result, fmt.Errorf("failed to marshal arguments: %w", err)
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("failed to parse arguments: %w", err)
	}
	return result, nil
}

// --- Roadmap Args ---

// GetRoadmapArgs holds arguments for roadmap get operations.
type GetRoadmapArgs struct {
	RoadmapID string `json:"roadmap_id"`
}

// Validate checks required fields.
func (a GetRoadmapArgs) Validate() error {
	if a.RoadmapID == "" {
		return fmt.Errorf("required parameter missing: roadmap_id")
	}
	return nil
}

// ManageLaneArgs holds arguments for lane management operations.
type ManageLaneArgs struct {
	Action    string `json:"action"`
	RoadmapID string `json:"roadmap_id"`
	LaneID    string `json:"lane_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Color     string `json:"color,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageLaneArgs) Validate() error {
	if a.Action == "" {
		return fmt.Errorf("required parameter missing: action")
	}
	if a.RoadmapID == "" {
		return fmt.Errorf("required parameter missing: roadmap_id")
	}
	switch a.Action {
	case "update", "delete":
		if a.LaneID == "" {
			return fmt.Errorf("required parameter missing: lane_id (required for %s)", a.Action)
		}
	}
	return nil
}

// ManageMilestoneArgs holds arguments for milestone management operations.
type ManageMilestoneArgs struct {
	Action      string `json:"action"`
	RoadmapID   string `json:"roadmap_id"`
	MilestoneID string `json:"milestone_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Date        string `json:"date,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageMilestoneArgs) Validate() error {
	if a.Action == "" {
		return fmt.Errorf("required parameter missing: action")
	}
	if a.RoadmapID == "" {
		return fmt.Errorf("required parameter missing: roadmap_id")
	}
	switch a.Action {
	case "update", "delete":
		if a.MilestoneID == "" {
			return fmt.Errorf("required parameter missing: milestone_id (required for %s)", a.Action)
		}
	}
	return nil
}

// --- Bar Args ---

// GetBarArgs holds arguments for bar get operations.
type GetBarArgs struct {
	BarID string `json:"bar_id"`
}

// Validate checks required fields.
func (a GetBarArgs) Validate() error {
	if a.BarID == "" {
		return fmt.Errorf("required parameter missing: bar_id")
	}
	return nil
}

// ManageBarArgs holds arguments for bar management operations.
type ManageBarArgs struct {
	Action         string `json:"action"`
	BarID          string `json:"bar_id,omitempty"`
	RoadmapID      string `json:"roadmap_id,omitempty"`
	LaneID         string `json:"lane_id,omitempty"`
	Name           string `json:"name,omitempty"`
	StartDate      string `json:"start_date,omitempty"`
	EndDate        string `json:"end_date,omitempty"`
	Description    string `json:"description,omitempty"`
	LegendID       string `json:"legend_id,omitempty"`
	PercentDone    *int   `json:"percent_done,omitempty"`
	Container      *bool  `json:"container,omitempty"`
	Parked         *bool  `json:"parked,omitempty"`
	ParentID       string `json:"parent_id,omitempty"`
	StrategicValue string `json:"strategic_value,omitempty"`
	Notes          string `json:"notes,omitempty"`
	Effort         *int   `json:"effort,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageBarArgs) Validate() error {
	if a.Action == "" {
		return fmt.Errorf("required parameter missing: action")
	}
	switch a.Action {
	case "create":
		if a.RoadmapID == "" {
			return fmt.Errorf("required parameter missing: roadmap_id (required for create)")
		}
		if a.LaneID == "" {
			return fmt.Errorf("required parameter missing: lane_id (required for create)")
		}
		if a.Name == "" {
			return fmt.Errorf("required parameter missing: name (required for create)")
		}
	case "update", "delete":
		if a.BarID == "" {
			return fmt.Errorf("required parameter missing: bar_id (required for %s)", a.Action)
		}
	}
	return nil
}

// ManageBarCommentArgs holds arguments for bar comment operations.
type ManageBarCommentArgs struct {
	BarID string `json:"bar_id"`
	Body  string `json:"body"`
}

// Validate checks required fields.
func (a ManageBarCommentArgs) Validate() error {
	if a.BarID == "" {
		return fmt.Errorf("required parameter missing: bar_id")
	}
	if a.Body == "" {
		return fmt.Errorf("required parameter missing: body")
	}
	return nil
}

// ManageBarConnectionArgs holds arguments for bar connection operations.
type ManageBarConnectionArgs struct {
	Action       string `json:"action"`
	BarID        string `json:"bar_id"`
	TargetBarID  string `json:"target_bar_id,omitempty"`
	ConnectionID string `json:"connection_id,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageBarConnectionArgs) Validate() error {
	if a.Action == "" {
		return fmt.Errorf("required parameter missing: action")
	}
	if a.BarID == "" {
		return fmt.Errorf("required parameter missing: bar_id")
	}
	switch a.Action {
	case "create":
		if a.TargetBarID == "" {
			return fmt.Errorf("required parameter missing: target_bar_id (required for create)")
		}
	case "delete":
		if a.ConnectionID == "" {
			return fmt.Errorf("required parameter missing: connection_id (required for delete)")
		}
	}
	return nil
}

// ManageBarLinkArgs holds arguments for bar link operations.
type ManageBarLinkArgs struct {
	Action string `json:"action"`
	BarID  string `json:"bar_id"`
	LinkID string `json:"link_id,omitempty"`
	URL    string `json:"url,omitempty"`
	Name   string `json:"name,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageBarLinkArgs) Validate() error {
	if a.Action == "" {
		return fmt.Errorf("required parameter missing: action")
	}
	if a.BarID == "" {
		return fmt.Errorf("required parameter missing: bar_id")
	}
	switch a.Action {
	case "create":
		if a.URL == "" {
			return fmt.Errorf("required parameter missing: url (required for create)")
		}
	case "update", "delete":
		if a.LinkID == "" {
			return fmt.Errorf("required parameter missing: link_id (required for %s)", a.Action)
		}
	}
	return nil
}

// --- Objective Args ---

// GetObjectiveArgs holds arguments for objective get operations.
type GetObjectiveArgs struct {
	ObjectiveID string `json:"objective_id"`
}

// Validate checks required fields.
func (a GetObjectiveArgs) Validate() error {
	if a.ObjectiveID == "" {
		return fmt.Errorf("required parameter missing: objective_id")
	}
	return nil
}

// ManageObjectiveArgs holds arguments for objective management operations.
type ManageObjectiveArgs struct {
	Action      string `json:"action"`
	ObjectiveID string `json:"objective_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	TimeFrame   string `json:"time_frame,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageObjectiveArgs) Validate() error {
	if a.Action == "" {
		return fmt.Errorf("required parameter missing: action")
	}
	switch a.Action {
	case "create":
		if a.Name == "" {
			return fmt.Errorf("required parameter missing: name (required for create)")
		}
	case "update", "delete":
		if a.ObjectiveID == "" {
			return fmt.Errorf("required parameter missing: objective_id (required for %s)", a.Action)
		}
	}
	return nil
}

// ManageKeyResultArgs holds arguments for key result management operations.
type ManageKeyResultArgs struct {
	Action       string `json:"action"`
	ObjectiveID  string `json:"objective_id"`
	KeyResultID  string `json:"key_result_id,omitempty"`
	Name         string `json:"name,omitempty"`
	TargetValue  string `json:"target_value,omitempty"`
	CurrentValue string `json:"current_value,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageKeyResultArgs) Validate() error {
	if a.Action == "" {
		return fmt.Errorf("required parameter missing: action")
	}
	if a.ObjectiveID == "" {
		return fmt.Errorf("required parameter missing: objective_id")
	}
	switch a.Action {
	case "update", "delete":
		if a.KeyResultID == "" {
			return fmt.Errorf("required parameter missing: key_result_id (required for %s)", a.Action)
		}
	}
	return nil
}

// --- Idea Args ---

// GetIdeaArgs holds arguments for idea get operations.
type GetIdeaArgs struct {
	IdeaID string `json:"idea_id"`
}

// Validate checks required fields.
func (a GetIdeaArgs) Validate() error {
	if a.IdeaID == "" {
		return fmt.Errorf("required parameter missing: idea_id")
	}
	return nil
}

// GetOpportunityArgs holds arguments for opportunity get operations.
type GetOpportunityArgs struct {
	OpportunityID string `json:"opportunity_id"`
}

// Validate checks required fields.
func (a GetOpportunityArgs) Validate() error {
	if a.OpportunityID == "" {
		return fmt.Errorf("required parameter missing: opportunity_id")
	}
	return nil
}

// GetIdeaFormArgs holds arguments for idea form get operations.
type GetIdeaFormArgs struct {
	FormID string `json:"form_id"`
}

// Validate checks required fields.
func (a GetIdeaFormArgs) Validate() error {
	if a.FormID == "" {
		return fmt.Errorf("required parameter missing: form_id")
	}
	return nil
}

// ManageIdeaArgs holds arguments for idea management operations.
type ManageIdeaArgs struct {
	Action      string `json:"action"`
	IdeaID      string `json:"idea_id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageIdeaArgs) Validate() error {
	if a.Action == "" {
		return fmt.Errorf("required parameter missing: action")
	}
	switch a.Action {
	case "create":
		if a.Title == "" {
			return fmt.Errorf("required parameter missing: title (required for create)")
		}
	case "update":
		if a.IdeaID == "" {
			return fmt.Errorf("required parameter missing: idea_id (required for update)")
		}
	}
	return nil
}

// ManageIdeaCustomerArgs holds arguments for idea customer operations.
type ManageIdeaCustomerArgs struct {
	Action     string `json:"action"`
	IdeaID     string `json:"idea_id"`
	CustomerID string `json:"customer_id,omitempty"`
	Name       string `json:"name,omitempty"`
	Email      string `json:"email,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageIdeaCustomerArgs) Validate() error {
	if a.Action == "" {
		return fmt.Errorf("required parameter missing: action")
	}
	if a.IdeaID == "" {
		return fmt.Errorf("required parameter missing: idea_id")
	}
	switch a.Action {
	case "remove":
		if a.CustomerID == "" {
			return fmt.Errorf("required parameter missing: customer_id (required for remove)")
		}
	}
	return nil
}

// ManageIdeaTagArgs holds arguments for idea tag operations.
type ManageIdeaTagArgs struct {
	Action string `json:"action"`
	IdeaID string `json:"idea_id"`
	TagID  string `json:"tag_id,omitempty"`
	Name   string `json:"name,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageIdeaTagArgs) Validate() error {
	if a.Action == "" {
		return fmt.Errorf("required parameter missing: action")
	}
	if a.IdeaID == "" {
		return fmt.Errorf("required parameter missing: idea_id")
	}
	switch a.Action {
	case "remove":
		if a.TagID == "" {
			return fmt.Errorf("required parameter missing: tag_id (required for remove)")
		}
	}
	return nil
}

// ManageOpportunityArgs holds arguments for opportunity management operations.
type ManageOpportunityArgs struct {
	Action           string `json:"action"`
	OpportunityID    string `json:"opportunity_id,omitempty"`
	ProblemStatement string `json:"problem_statement,omitempty"`
	Description      string `json:"description,omitempty"`
	WorkflowStatus   string `json:"workflow_status,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageOpportunityArgs) Validate() error {
	if a.Action == "" {
		return fmt.Errorf("required parameter missing: action")
	}
	switch a.Action {
	case "create":
		if a.ProblemStatement == "" {
			return fmt.Errorf("required parameter missing: problem_statement (required for create)")
		}
	case "update", "delete":
		if a.OpportunityID == "" {
			return fmt.Errorf("required parameter missing: opportunity_id (required for %s)", a.Action)
		}
	}
	return nil
}

// --- Launch Args ---

// GetLaunchArgs holds arguments for launch get operations.
type GetLaunchArgs struct {
	LaunchID string `json:"launch_id"`
}

// Validate checks required fields.
func (a GetLaunchArgs) Validate() error {
	if a.LaunchID == "" {
		return fmt.Errorf("required parameter missing: launch_id")
	}
	return nil
}

// --- Utility Args ---

// HealthCheckArgs holds arguments for health check operations.
type HealthCheckArgs struct {
	Deep bool `json:"deep,omitempty"`
}

// Validate always returns nil (no required fields).
func (a HealthCheckArgs) Validate() error {
	return nil
}
