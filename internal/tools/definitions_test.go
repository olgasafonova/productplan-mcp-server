package tools

import (
	"testing"
)

func TestBuildAllTools(t *testing.T) {
	tools := BuildAllTools()

	if len(tools) == 0 {
		t.Fatal("expected tools to be registered")
	}

	// Should have 36 tools total
	if len(tools) < 30 {
		t.Errorf("expected at least 30 tools, got %d", len(tools))
	}
}

func TestBuildAllToolsNames(t *testing.T) {
	tools := BuildAllTools()

	expectedNames := []string{
		// Roadmaps
		"list_roadmaps",
		"get_roadmap",
		"get_roadmap_bars",
		"get_roadmap_lanes",
		"get_roadmap_milestones",
		"get_roadmap_complete",
		"manage_lane",
		"manage_milestone",
		// Bars
		"get_bar",
		"get_bar_children",
		"get_bar_comments",
		"get_bar_connections",
		"get_bar_links",
		"manage_bar",
		"manage_bar_comment",
		"manage_bar_connection",
		"manage_bar_link",
		// Objectives
		"list_objectives",
		"get_objective",
		"list_key_results",
		"manage_objective",
		"manage_key_result",
		// Ideas
		"list_ideas",
		"get_idea",
		"get_idea_customers",
		"get_idea_tags",
		"list_opportunities",
		"get_opportunity",
		"list_idea_forms",
		"get_idea_form",
		"manage_idea",
		"manage_idea_customer",
		"manage_idea_tag",
		"manage_opportunity",
		// Launches
		"list_launches",
		"get_launch",
		// Utility
		"check_status",
		"health_check",
	}

	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
	}

	for _, expected := range expectedNames {
		if !names[expected] {
			t.Errorf("expected tool %q not found", expected)
		}
	}
}

func TestBuildAllToolsHaveDescriptions(t *testing.T) {
	tools := BuildAllTools()

	for _, tool := range tools {
		if tool.Description == "" {
			t.Errorf("tool %q has no description", tool.Name)
		}
	}
}

func TestBuildAllToolsHaveInputSchemas(t *testing.T) {
	tools := BuildAllTools()

	for _, tool := range tools {
		if tool.InputSchema.Type != "object" {
			t.Errorf("tool %q has invalid input schema type: %q", tool.Name, tool.InputSchema.Type)
		}
	}
}

func TestRoadmapTools(t *testing.T) {
	tools := roadmapTools()

	if len(tools) != 8 {
		t.Errorf("expected 8 roadmap tools, got %d", len(tools))
	}

	// Check list_roadmaps has no required params
	for _, tool := range tools {
		if tool.Name == "list_roadmaps" {
			if len(tool.InputSchema.Required) != 0 {
				t.Errorf("list_roadmaps should not have required params")
			}
		}
		if tool.Name == "get_roadmap" {
			if len(tool.InputSchema.Required) != 1 || tool.InputSchema.Required[0] != "roadmap_id" {
				t.Errorf("get_roadmap should require roadmap_id")
			}
		}
	}
}

func TestBarTools(t *testing.T) {
	tools := barTools()

	if len(tools) != 9 {
		t.Errorf("expected 9 bar tools, got %d", len(tools))
	}
}

func TestObjectiveTools(t *testing.T) {
	tools := objectiveTools()

	if len(tools) != 5 {
		t.Errorf("expected 5 objective tools, got %d", len(tools))
	}
}

func TestIdeaTools(t *testing.T) {
	tools := ideaTools()

	if len(tools) != 12 {
		t.Errorf("expected 12 idea tools, got %d", len(tools))
	}
}

func TestLaunchTools(t *testing.T) {
	tools := launchTools()

	if len(tools) != 2 {
		t.Errorf("expected 2 launch tools, got %d", len(tools))
	}
}

func TestUtilityTools(t *testing.T) {
	tools := utilityTools()

	if len(tools) != 2 {
		t.Errorf("expected 2 utility tools, got %d", len(tools))
	}
}

func TestManageBarToolHasActionEnum(t *testing.T) {
	tools := barTools()

	for _, tool := range tools {
		if tool.Name == "manage_bar" {
			actionProp, ok := tool.InputSchema.Properties["action"]
			if !ok {
				t.Fatal("manage_bar should have action property")
			}
			if len(actionProp.Enum) != 3 {
				t.Errorf("manage_bar action should have 3 enum values, got %d", len(actionProp.Enum))
			}
		}
	}
}

func TestHealthCheckToolHasDeepParam(t *testing.T) {
	tools := utilityTools()

	for _, tool := range tools {
		if tool.Name == "health_check" {
			deepProp, ok := tool.InputSchema.Properties["deep"]
			if !ok {
				t.Fatal("health_check should have deep property")
			}
			if deepProp.Type != "boolean" {
				t.Errorf("deep property should be boolean, got %q", deepProp.Type)
			}
		}
	}
}
