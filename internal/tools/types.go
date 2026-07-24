// Package tools provides typed argument structs for ProductPlan MCP tool handlers.
package tools

import (
	"encoding/json"
	"fmt"
	"slices"
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

// fieldCheck pairs a struct field value with its JSON name. It is the
// domain unit of validation: every required-parameter rule is expressed as
// a fieldCheck plus one of the require methods below.
type fieldCheck struct{ value, name string }

// require returns a "required parameter missing" error when the value is empty.
func (c fieldCheck) require() error {
	if c.value == "" {
		return fmt.Errorf("required parameter missing: %s", c.name)
	}
	return nil
}

// requireFor returns a "required parameter missing" error scoped to a
// specific action when the value is empty.
func (c fieldCheck) requireFor(action string) error {
	if c.value == "" {
		return fmt.Errorf("required parameter missing: %s (required for %s)", c.name, action)
	}
	return nil
}

// requireAll runs require against each check in order, returning the first
// error encountered. Centralises the "validate N mandatory fields" pattern.
func requireAll(checks ...fieldCheck) error {
	for _, c := range checks {
		if err := c.require(); err != nil {
			return err
		}
	}
	return nil
}

// requireAllForAction runs requireFor against each check in order, returning
// the first error encountered. Centralises the "validate N action-gated
// fields" pattern.
func requireAllForAction(action string, checks ...fieldCheck) error {
	for _, c := range checks {
		if err := c.requireFor(action); err != nil {
			return err
		}
	}
	return nil
}

// validateParentScoped validates a manage struct for a resource that lives
// under a parent: action and the parent ID are always required, and the
// resource's own ID is required for "update" and "delete".
func validateParentScoped(action string, parent, id fieldCheck) error {
	if err := requireAll(fieldCheck{action, "action"}, parent); err != nil {
		return err
	}
	if action == "update" || action == "delete" {
		return id.requireFor(action)
	}
	return nil
}

// validateCreateOrByID validates a top-level manage struct: action is always
// required, the create field is required for "create", and the ID field is
// required for any action listed in idActions.
func validateCreateOrByID(action string, create, id fieldCheck, idActions ...string) error {
	if err := (fieldCheck{action, "action"}).require(); err != nil {
		return err
	}
	if action == "create" {
		return create.requireFor("create")
	}
	if slices.Contains(idActions, action) {
		return id.requireFor(action)
	}
	return nil
}

// validateBarSubresource validates bar-scoped sub-resources (connections,
// links): action and bar_id are always required, the create field is required
// for "create", and the delete field is required for "delete".
func validateBarSubresource(action string, bar, create, del fieldCheck) error {
	if err := requireAll(fieldCheck{action, "action"}, bar); err != nil {
		return err
	}
	switch action {
	case "create":
		return create.requireFor("create")
	case "delete":
		return del.requireFor("delete")
	}
	return nil
}

// --- Roadmap Args ---

// GetRoadmapArgs holds arguments for roadmap get operations.
type GetRoadmapArgs struct {
	RoadmapID string `json:"roadmap_id"`
}

// Validate checks required fields.
func (a GetRoadmapArgs) Validate() error {
	return fieldCheck{a.RoadmapID, "roadmap_id"}.require()
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
	return validateParentScoped(a.Action,
		fieldCheck{a.RoadmapID, "roadmap_id"}, fieldCheck{a.LaneID, "lane_id"})
}

// ManageMilestoneArgs holds arguments for milestone management operations.
type ManageMilestoneArgs struct {
	Action      string `json:"action"`
	RoadmapID   string `json:"roadmap_id"`
	MilestoneID string `json:"milestone_id,omitempty"`
	Title       string `json:"title,omitempty"`
	Date        string `json:"date,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageMilestoneArgs) Validate() error {
	return validateParentScoped(a.Action,
		fieldCheck{a.RoadmapID, "roadmap_id"}, fieldCheck{a.MilestoneID, "milestone_id"})
}

// --- Bar Args ---

// GetBarArgs holds arguments for bar get operations.
type GetBarArgs struct {
	BarID string `json:"bar_id"`
}

// Validate checks required fields.
func (a GetBarArgs) Validate() error {
	return fieldCheck{a.BarID, "bar_id"}.require()
}

// CustomFieldValue represents a name-value pair for custom fields.
type CustomFieldValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ManageBarArgs holds arguments for bar management operations.
type ManageBarArgs struct {
	Action               string             `json:"action"`
	BarID                string             `json:"bar_id,omitempty"`
	RoadmapID            string             `json:"roadmap_id,omitempty"`
	LaneID               string             `json:"lane_id,omitempty"`
	Name                 string             `json:"name,omitempty"`
	StartsOn             string             `json:"starts_on,omitempty"`
	EndsOn               string             `json:"ends_on,omitempty"`
	Description          string             `json:"description,omitempty"`
	LegendID             string             `json:"legend_id,omitempty"`
	PercentDone          *int               `json:"percent_done,omitempty"`
	Container            *bool              `json:"container,omitempty"`
	Parked               *bool              `json:"parked,omitempty"`
	ParentID             string             `json:"parent_id,omitempty"`
	StrategicValue       string             `json:"strategic_value,omitempty"`
	Notes                string             `json:"notes,omitempty"`
	Effort               *int               `json:"effort,omitempty"`
	Tags                 []string           `json:"tags,omitempty"`
	CustomTextFields     []CustomFieldValue `json:"custom_text_fields,omitempty"`
	CustomDropdownFields []CustomFieldValue `json:"custom_dropdown_fields,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageBarArgs) Validate() error {
	if err := (fieldCheck{a.Action, "action"}).require(); err != nil {
		return err
	}
	switch a.Action {
	case "create":
		return requireAllForAction("create",
			fieldCheck{a.RoadmapID, "roadmap_id"},
			fieldCheck{a.LaneID, "lane_id"},
			fieldCheck{a.Name, "name"},
		)
	case "update", "delete":
		return fieldCheck{a.BarID, "bar_id"}.requireFor(a.Action)
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
	return validateBarSubresource(a.Action, fieldCheck{a.BarID, "bar_id"},
		fieldCheck{a.TargetBarID, "target_bar_id"}, fieldCheck{a.ConnectionID, "connection_id"})
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
	return validateBarSubresource(a.Action, fieldCheck{a.BarID, "bar_id"},
		fieldCheck{a.URL, "url"}, fieldCheck{a.LinkID, "link_id"})
}

// --- Objective Args ---

// GetObjectiveArgs holds arguments for objective get operations.
type GetObjectiveArgs struct {
	ObjectiveID string `json:"objective_id"`
}

// Validate checks required fields.
func (a GetObjectiveArgs) Validate() error {
	return fieldCheck{a.ObjectiveID, "objective_id"}.require()
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
	return validateCreateOrByID(a.Action, fieldCheck{a.Name, "name"},
		fieldCheck{a.ObjectiveID, "objective_id"}, "update", "delete")
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
	return validateParentScoped(a.Action,
		fieldCheck{a.ObjectiveID, "objective_id"}, fieldCheck{a.KeyResultID, "key_result_id"})
}

// GetKeyResultArgs holds arguments for key result get operations.
type GetKeyResultArgs struct {
	ObjectiveID string `json:"objective_id"`
	KeyResultID string `json:"key_result_id"`
}

// Validate checks required fields.
func (a GetKeyResultArgs) Validate() error {
	return requireAll(
		fieldCheck{a.ObjectiveID, "objective_id"},
		fieldCheck{a.KeyResultID, "key_result_id"},
	)
}

// --- Idea Args ---

// GetIdeaArgs holds arguments for idea get operations.
type GetIdeaArgs struct {
	IdeaID string `json:"idea_id"`
}

// Validate checks required fields.
func (a GetIdeaArgs) Validate() error {
	return fieldCheck{a.IdeaID, "idea_id"}.require()
}

// GetOpportunityArgs holds arguments for opportunity get operations.
type GetOpportunityArgs struct {
	OpportunityID string `json:"opportunity_id"`
}

// Validate checks required fields.
func (a GetOpportunityArgs) Validate() error {
	return fieldCheck{a.OpportunityID, "opportunity_id"}.require()
}

// GetIdeaFormArgs holds arguments for idea form get operations.
type GetIdeaFormArgs struct {
	FormID string `json:"form_id"`
}

// Validate checks required fields.
func (a GetIdeaFormArgs) Validate() error {
	return fieldCheck{a.FormID, "form_id"}.require()
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
	return validateCreateOrByID(a.Action, fieldCheck{a.Title, "title"},
		fieldCheck{a.IdeaID, "idea_id"}, "update")
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
	return validateCreateOrByID(a.Action, fieldCheck{a.ProblemStatement, "problem_statement"},
		fieldCheck{a.OpportunityID, "opportunity_id"}, "update")
}

// --- Launch Args ---

// GetLaunchArgs holds arguments for launch get operations.
type GetLaunchArgs struct {
	LaunchID string `json:"launch_id"`
}

// Validate checks required fields.
func (a GetLaunchArgs) Validate() error {
	return fieldCheck{a.LaunchID, "launch_id"}.require()
}

// GetLaunchSectionArgs holds arguments for getting a single launch section.
type GetLaunchSectionArgs struct {
	LaunchID  string `json:"launch_id"`
	SectionID string `json:"section_id"`
}

// Validate checks required fields.
func (a GetLaunchSectionArgs) Validate() error {
	return requireAll(
		fieldCheck{a.LaunchID, "launch_id"},
		fieldCheck{a.SectionID, "section_id"},
	)
}

// GetLaunchTaskArgs holds arguments for getting a single launch task.
type GetLaunchTaskArgs struct {
	LaunchID string `json:"launch_id"`
	TaskID   string `json:"task_id"`
}

// Validate checks required fields.
func (a GetLaunchTaskArgs) Validate() error {
	return requireAll(
		fieldCheck{a.LaunchID, "launch_id"},
		fieldCheck{a.TaskID, "task_id"},
	)
}

// ManageLaunchArgs holds arguments for launch management operations.
type ManageLaunchArgs struct {
	Action      string `json:"action"`
	LaunchID    string `json:"launch_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Date        string `json:"date,omitempty"`
	Description string `json:"description,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageLaunchArgs) Validate() error {
	return validateCreateOrByID(a.Action, fieldCheck{a.Name, "name"},
		fieldCheck{a.LaunchID, "launch_id"}, "update", "delete")
}

// ManageLaunchSectionArgs holds arguments for launch section management operations.
type ManageLaunchSectionArgs struct {
	Action    string `json:"action"`
	LaunchID  string `json:"launch_id"`
	SectionID string `json:"section_id,omitempty"`
	Name      string `json:"name,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageLaunchSectionArgs) Validate() error {
	return validateParentScoped(a.Action,
		fieldCheck{a.LaunchID, "launch_id"}, fieldCheck{a.SectionID, "section_id"})
}

// ManageLaunchTaskArgs holds arguments for launch task management operations.
type ManageLaunchTaskArgs struct {
	Action         string `json:"action"`
	LaunchID       string `json:"launch_id"`
	TaskID         string `json:"task_id,omitempty"`
	SectionID      string `json:"section_id,omitempty"`
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	DueDate        string `json:"due_date,omitempty"`
	AssignedUserID string `json:"assigned_user_id,omitempty"`
	Status         string `json:"status,omitempty"`
}

// Validate checks required fields based on action.
func (a ManageLaunchTaskArgs) Validate() error {
	if err := requireAll(
		fieldCheck{a.Action, "action"},
		fieldCheck{a.LaunchID, "launch_id"},
	); err != nil {
		return err
	}
	switch a.Action {
	case "create":
		return requireAllForAction("create",
			fieldCheck{a.SectionID, "section_id"},
			fieldCheck{a.Name, "name"},
		)
	case "update", "delete":
		return fieldCheck{a.TaskID, "task_id"}.requireFor(a.Action)
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
