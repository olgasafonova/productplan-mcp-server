package productplan

import (
	"strings"
	"testing"
)

func TestToolRegistry_Register(t *testing.T) {
	r := NewToolRegistry()

	def := &ToolDefinition{
		Name:        "test_tool",
		Description: "A test tool",
		Category:    CategoryRoadmaps,
		Properties: []PropertyDef{
			{Name: "id", Type: "string", Description: "The ID"},
		},
		Required: []string{"id"},
	}

	err := r.Register(def)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if r.Count() != 1 {
		t.Errorf("Expected 1 tool, got %d", r.Count())
	}
}

func TestToolRegistry_RegisterValidation(t *testing.T) {
	tests := []struct {
		name    string
		def     *ToolDefinition
		wantErr bool
	}{
		{
			name:    "empty name",
			def:     &ToolDefinition{Description: "test"},
			wantErr: true,
		},
		{
			name:    "empty description",
			def:     &ToolDefinition{Name: "test"},
			wantErr: true,
		},
		{
			name: "required not in properties",
			def: &ToolDefinition{
				Name:        "test",
				Description: "test",
				Required:    []string{"missing"},
			},
			wantErr: true,
		},
		{
			name: "valid tool",
			def: &ToolDefinition{
				Name:        "test",
				Description: "test",
				Properties: []PropertyDef{
					{Name: "id", Type: "string", Description: "ID"},
				},
				Required: []string{"id"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewToolRegistry()
			err := r.Register(tt.def)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToolRegistry_DuplicateRegistration(t *testing.T) {
	r := NewToolRegistry()

	def := &ToolDefinition{
		Name:        "test_tool",
		Description: "A test tool",
		Category:    CategoryRoadmaps,
	}

	err := r.Register(def)
	if err != nil {
		t.Errorf("First registration failed: %v", err)
	}

	err = r.Register(def)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}
}

func TestToolRegistry_Get(t *testing.T) {
	r := NewToolRegistry()

	def := &ToolDefinition{
		Name:        "test_tool",
		Description: "A test tool",
		Category:    CategoryRoadmaps,
	}
	r.MustRegister(def)

	got, ok := r.Get("test_tool")
	if !ok {
		t.Error("Expected to find tool")
	}
	if got.Name != "test_tool" {
		t.Errorf("Expected 'test_tool', got '%s'", got.Name)
	}

	_, ok = r.Get("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent tool")
	}
}

func TestToolRegistry_All(t *testing.T) {
	r := NewToolRegistry()

	tools := []*ToolDefinition{
		{Name: "tool1", Description: "First", Category: CategoryRoadmaps},
		{Name: "tool2", Description: "Second", Category: CategoryBars},
		{Name: "tool3", Description: "Third", Category: CategoryRoadmaps},
	}

	for _, def := range tools {
		r.MustRegister(def)
	}

	all := r.All()
	if len(all) != 3 {
		t.Errorf("Expected 3 tools, got %d", len(all))
	}

	// Check order is preserved
	for i, def := range all {
		if def.Name != tools[i].Name {
			t.Errorf("Order mismatch at %d: expected %s, got %s", i, tools[i].Name, def.Name)
		}
	}
}

func TestToolRegistry_ByCategory(t *testing.T) {
	r := NewToolRegistry()

	r.MustRegister(&ToolDefinition{Name: "rm1", Description: "Roadmap 1", Category: CategoryRoadmaps})
	r.MustRegister(&ToolDefinition{Name: "bar1", Description: "Bar 1", Category: CategoryBars})
	r.MustRegister(&ToolDefinition{Name: "rm2", Description: "Roadmap 2", Category: CategoryRoadmaps})

	roadmaps := r.ByCategory(CategoryRoadmaps)
	if len(roadmaps) != 2 {
		t.Errorf("Expected 2 roadmap tools, got %d", len(roadmaps))
	}

	bars := r.ByCategory(CategoryBars)
	if len(bars) != 1 {
		t.Errorf("Expected 1 bar tool, got %d", len(bars))
	}
}

func TestToolBuilder(t *testing.T) {
	def := NewTool("manage_bar").
		Description("Create, update, or delete a bar").
		Category(CategoryBars).
		Handler("handleManageBar").
		Prop("bar_id", "string", "The bar ID").
		PropEnum("action", "Action to perform", "create", "update", "delete").
		Required("action").
		Build()

	if def.Name != "manage_bar" {
		t.Errorf("Expected name 'manage_bar', got '%s'", def.Name)
	}
	if def.Category != CategoryBars {
		t.Errorf("Expected category bars, got %s", def.Category)
	}
	if len(def.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(def.Properties))
	}
	if len(def.Required) != 1 {
		t.Errorf("Expected 1 required, got %d", len(def.Required))
	}

	// Check enum property
	var enumProp *PropertyDef
	for i := range def.Properties {
		if def.Properties[i].Name == "action" {
			enumProp = &def.Properties[i]
			break
		}
	}
	if enumProp == nil {
		t.Fatal("Expected to find action property")
	}
	if len(enumProp.Enum) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(enumProp.Enum))
	}
}

func TestToolRegistry_ToMCPFormat(t *testing.T) {
	r := NewToolRegistry()

	r.MustRegister(NewTool("list_items").
		Description("List all items").
		Category(CategoryRoadmaps).
		Build())

	r.MustRegister(NewTool("get_item").
		Description("Get an item by ID").
		Category(CategoryRoadmaps).
		Prop("id", "string", "Item ID").
		Required("id").
		Build())

	mcpTools := r.ToMCPFormat()

	if len(mcpTools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(mcpTools))
	}

	// Check first tool
	if mcpTools[0]["name"] != "list_items" {
		t.Errorf("Expected 'list_items', got %v", mcpTools[0]["name"])
	}

	// Check second tool has required
	schema, ok := mcpTools[1]["inputSchema"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected inputSchema to be a map")
	}
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Expected required to be []string")
	}
	if len(required) != 1 || required[0] != "id" {
		t.Errorf("Expected required=['id'], got %v", required)
	}
}

func TestToolRegistry_GenerateMarkdownDocs(t *testing.T) {
	r := NewToolRegistry()

	r.MustRegister(NewTool("list_roadmaps").
		Description("List all roadmaps").
		Category(CategoryRoadmaps).
		Build())

	r.MustRegister(NewTool("get_bar").
		Description("Get a bar by ID").
		Category(CategoryBars).
		Prop("bar_id", "string", "The bar ID").
		Required("bar_id").
		Build())

	docs := r.GenerateMarkdownDocs()

	if !strings.Contains(docs, "# Tool Reference") {
		t.Error("Expected markdown header")
	}
	if !strings.Contains(docs, "list_roadmaps") {
		t.Error("Expected list_roadmaps in docs")
	}
	if !strings.Contains(docs, "*(required)*") {
		t.Error("Expected required marker")
	}
}

func TestToolRegistry_Summary(t *testing.T) {
	r := NewToolRegistry()

	r.MustRegister(&ToolDefinition{Name: "rm1", Description: "Roadmap 1", Category: CategoryRoadmaps})
	r.MustRegister(&ToolDefinition{Name: "bar1", Description: "Bar 1", Category: CategoryBars})
	r.MustRegister(&ToolDefinition{Name: "rm2", Description: "Roadmap 2", Category: CategoryRoadmaps})

	summary := r.Summary()

	if !strings.Contains(summary, "3 tools") {
		t.Errorf("Expected '3 tools' in summary, got: %s", summary)
	}
	if !strings.Contains(summary, "roadmaps: 2") {
		t.Errorf("Expected 'roadmaps: 2' in summary, got: %s", summary)
	}
}

func TestToolRegistry_MustRegister_Panic(t *testing.T) {
	r := NewToolRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid tool")
		}
	}()

	// Should panic - no description
	r.MustRegister(&ToolDefinition{Name: "bad_tool"})
}

func TestToolRegistry_Names(t *testing.T) {
	r := NewToolRegistry()

	r.MustRegister(&ToolDefinition{Name: "alpha", Description: "A", Category: CategoryRoadmaps})
	r.MustRegister(&ToolDefinition{Name: "beta", Description: "B", Category: CategoryRoadmaps})

	names := r.Names()

	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}
	if names[0] != "alpha" || names[1] != "beta" {
		t.Errorf("Expected [alpha, beta], got %v", names)
	}

	// Modifying returned slice shouldn't affect registry
	names[0] = "modified"
	names2 := r.Names()
	if names2[0] != "alpha" {
		t.Error("Registry should not be modified")
	}
}
