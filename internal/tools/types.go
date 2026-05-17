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

// fieldCheck pairs a struct field value with its JSON name for batch validation.
type fieldCheck struct{ value, name string }

// requireField returns a "required parameter missing" error when value is empty.
// Returns nil otherwise.
func requireField(value, name string) error {
	if value == "" {
		return fmt.Errorf("required parameter missing: %s", name)
	}
	return nil
}

// requireFieldForAction returns a "required parameter missing" error scoped
// to a specific action. Returns nil when value is non-empty.
func requireFieldForAction(value, name, action string) error {
	if value == "" {
		return fmt.Errorf("required parameter missing: %s (required for %s)", name, action)
	}
	return nil
}

// requireAll runs requireField against each check in order, returning the
// first error encountered. Centralises the "validate N mandatory fields" pattern.
func requireAll(checks ...fieldCheck) error {
	for _, c := range checks {
		if err := requireField(c.value, c.name); err != nil {
			return err
		}
	}
	return nil
}

// requireAllForAction runs requireFieldForAction against each check in order,
// returning the first error encountered. Centralises the "validate N action-gated
// fields" pattern.
func requireAllForAction(action string, checks ...fieldCheck) error {
	for _, c := range checks {
		if err := requireFieldForAction(c.value, c.name, action); err != nil {
			return err
		}
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
	return requireField(a.RoadmapID, "roadmap_id")
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
	if err := requireAll(
		fieldCheck{a.Action, "action"},
		fieldCheck{a.RoadmapID, "roadmap_id"},
	); err != nil {
		return err
	}
	if a.Action == "update" || a.Action == "delete" {
		return requireFieldForAction(a.LaneID, "lane_id", a.Action)
	}
	return nil
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
	if err := requireAll(
		fieldCheck{a.Action, "action"},
		fieldCheck{a.RoadmapID, "roadmap_id"},
	); err != nil {
		return err
	}
	if a.Action == "update" || a.Action == "delete" {
		return requireFieldForAction(a.MilestoneID, "milestone_id", a.Action)
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
	return requireField(a.BarID, "bar_id")
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
	if err := requireField(a.Action, "action"); err != nil {
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
		return requireFieldForAction(a.BarID, "bar_id", a.Action)
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
	if err := requireAll(
		fieldCheck{a.Action, "action"},
		fieldCheck{a.BarID, "bar_id"},
	); err != nil {
		return err
	}
	switch a.Action {
	case "create":
		return requireFieldForAction(a.TargetBarID, "target_bar_id", "create")
	case "delete":
		return requireFieldForAction(a.ConnectionID, "connection_id", "delete")
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
	if err := requireAll(
		fieldCheck{a.Action, "action"},
		fieldCheck{a.BarID, "bar_id"},
	); err != nil {
		return err
	}
	switch a.Action {
	case "create":
		return requireFieldForAction(a.URL, "url", "create")
	case "delete":
		return requireFieldForAction(a.LinkID, "link_id", "delete")
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
	return requireField(a.ObjectiveID, "objective_id")
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
	if err := requireField(a.Action, "action"); err != nil {
		return err
	}
	switch a.Action {
	case "create":
		return requireFieldForAction(a.Name, "name", "create")
	case "update", "delete":
		return requireFieldForAction(a.ObjectiveID, "objective_id", a.Action)
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
	if err := requireAll(
		fieldCheck{a.Action, "action"},
		fieldCheck{a.ObjectiveID, "objective_id"},
	); err != nil {
		return err
	}
	if a.Action == "update" || a.Action == "delete" {
		return requireFieldForAction(a.KeyResultID, "key_result_id", a.Action)
	}
	return nil
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
	return requireField(a.IdeaID, "idea_id")
}

// GetOpportunityArgs holds arguments for opportunity get operations.
type GetOpportunityArgs struct {
	OpportunityID string `json:"opportunity_id"`
}

// Validate checks required fields.
func (a GetOpportunityArgs) Validate() error {
	return requireField(a.OpportunityID, "opportunity_id")
}

// GetIdeaFormArgs holds arguments for idea form get operations.
type GetIdeaFormArgs struct {
	FormID string `json:"form_id"`
}

// Validate checks required fields.
func (a GetIdeaFormArgs) Validate() error {
	return requireField(a.FormID, "form_id")
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
	if err := requireField(a.Action, "action"); err != nil {
		return err
	}
	switch a.Action {
	case "create":
		return requireFieldForAction(a.Title, "title", "create")
	case "update":
		return requireFieldForAction(a.IdeaID, "idea_id", "update")
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
	if err := requireField(a.Action, "action"); err != nil {
		return err
	}
	switch a.Action {
	case "create":
		return requireFieldForAction(a.ProblemStatement, "problem_statement", "create")
	case "update":
		return requireFieldForAction(a.OpportunityID, "opportunity_id", "update")
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
	return requireField(a.LaunchID, "launch_id")
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
	if err := requireField(a.Action, "action"); err != nil {
		return err
	}
	switch a.Action {
	case "create":
		return requireFieldForAction(a.Name, "name", "create")
	case "update", "delete":
		return requireFieldForAction(a.LaunchID, "launch_id", a.Action)
	}
	return nil
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
	if err := requireAll(
		fieldCheck{a.Action, "action"},
		fieldCheck{a.LaunchID, "launch_id"},
	); err != nil {
		return err
	}
	if a.Action == "update" || a.Action == "delete" {
		return requireFieldForAction(a.SectionID, "section_id", a.Action)
	}
	return nil
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
		return requireFieldForAction(a.TaskID, "task_id", a.Action)
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
