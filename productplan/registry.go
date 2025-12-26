package productplan

import (
	"fmt"
	"sort"
	"strings"
)

// ToolCategory groups related tools for organization.
type ToolCategory string

const (
	CategoryRoadmaps   ToolCategory = "roadmaps"
	CategoryBars       ToolCategory = "bars"
	CategoryObjectives ToolCategory = "objectives"
	CategoryIdeas      ToolCategory = "ideas"
	CategoryLaunches   ToolCategory = "launches"
	CategoryUtility    ToolCategory = "utility"
)

// ToolDefinition represents a complete tool specification.
type ToolDefinition struct {
	Name        string
	Description string
	Category    ToolCategory
	Properties  []PropertyDef
	Required    []string
	Handler     string // Handler function name for documentation
}

// PropertyDef defines a tool input property.
type PropertyDef struct {
	Name        string
	Type        string
	Description string
	Enum        []string
}

// ToolRegistry manages tool definitions.
type ToolRegistry struct {
	tools    map[string]*ToolDefinition
	order    []string // Maintain insertion order
	byCategory map[ToolCategory][]*ToolDefinition
}

// NewToolRegistry creates a new tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:      make(map[string]*ToolDefinition),
		order:      make([]string, 0),
		byCategory: make(map[ToolCategory][]*ToolDefinition),
	}
}

// Register adds a tool to the registry.
func (r *ToolRegistry) Register(def *ToolDefinition) error {
	if def.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	if def.Description == "" {
		return fmt.Errorf("tool %s: description is required", def.Name)
	}
	if _, exists := r.tools[def.Name]; exists {
		return fmt.Errorf("tool %s: already registered", def.Name)
	}

	// Validate required fields exist in properties
	propNames := make(map[string]bool)
	for _, p := range def.Properties {
		propNames[p.Name] = true
	}
	for _, req := range def.Required {
		if !propNames[req] {
			return fmt.Errorf("tool %s: required field %s not in properties", def.Name, req)
		}
	}

	r.tools[def.Name] = def
	r.order = append(r.order, def.Name)
	r.byCategory[def.Category] = append(r.byCategory[def.Category], def)
	return nil
}

// MustRegister registers a tool and panics on error.
func (r *ToolRegistry) MustRegister(def *ToolDefinition) {
	if err := r.Register(def); err != nil {
		panic(err)
	}
}

// Get retrieves a tool by name.
func (r *ToolRegistry) Get(name string) (*ToolDefinition, bool) {
	def, ok := r.tools[name]
	return def, ok
}

// All returns all registered tools in order.
func (r *ToolRegistry) All() []*ToolDefinition {
	result := make([]*ToolDefinition, len(r.order))
	for i, name := range r.order {
		result[i] = r.tools[name]
	}
	return result
}

// ByCategory returns tools in a specific category.
func (r *ToolRegistry) ByCategory(cat ToolCategory) []*ToolDefinition {
	return r.byCategory[cat]
}

// Names returns all tool names.
func (r *ToolRegistry) Names() []string {
	result := make([]string, len(r.order))
	copy(result, r.order)
	return result
}

// Count returns the number of registered tools.
func (r *ToolRegistry) Count() int {
	return len(r.tools)
}

// Categories returns all categories with tools.
func (r *ToolRegistry) Categories() []ToolCategory {
	cats := make([]ToolCategory, 0, len(r.byCategory))
	for cat := range r.byCategory {
		cats = append(cats, cat)
	}
	// Sort for consistent output
	sort.Slice(cats, func(i, j int) bool {
		return string(cats[i]) < string(cats[j])
	})
	return cats
}

// ToolBuilder provides a fluent API for building tool definitions.
type ToolBuilder struct {
	def *ToolDefinition
}

// NewTool starts building a new tool definition.
func NewTool(name string) *ToolBuilder {
	return &ToolBuilder{
		def: &ToolDefinition{
			Name:       name,
			Properties: make([]PropertyDef, 0),
			Required:   make([]string, 0),
		},
	}
}

// Description sets the tool description.
func (b *ToolBuilder) Description(desc string) *ToolBuilder {
	b.def.Description = desc
	return b
}

// Category sets the tool category.
func (b *ToolBuilder) Category(cat ToolCategory) *ToolBuilder {
	b.def.Category = cat
	return b
}

// Handler sets the handler function name.
func (b *ToolBuilder) Handler(name string) *ToolBuilder {
	b.def.Handler = name
	return b
}

// Prop adds a property to the tool.
func (b *ToolBuilder) Prop(name, propType, desc string) *ToolBuilder {
	b.def.Properties = append(b.def.Properties, PropertyDef{
		Name:        name,
		Type:        propType,
		Description: desc,
	})
	return b
}

// PropEnum adds an enum property.
func (b *ToolBuilder) PropEnum(name, desc string, values ...string) *ToolBuilder {
	b.def.Properties = append(b.def.Properties, PropertyDef{
		Name:        name,
		Type:        "string",
		Description: desc,
		Enum:        values,
	})
	return b
}

// Required marks properties as required.
func (b *ToolBuilder) Required(names ...string) *ToolBuilder {
	b.def.Required = append(b.def.Required, names...)
	return b
}

// Build returns the completed tool definition.
func (b *ToolBuilder) Build() *ToolDefinition {
	return b.def
}

// ToMCPFormat converts the registry to MCP tool format.
// This returns a format suitable for JSON marshaling in MCP responses.
func (r *ToolRegistry) ToMCPFormat() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(r.order))

	for _, name := range r.order {
		def := r.tools[name]

		props := make(map[string]interface{})
		for _, p := range def.Properties {
			propDef := map[string]interface{}{
				"type":        p.Type,
				"description": p.Description,
			}
			if len(p.Enum) > 0 {
				propDef["enum"] = p.Enum
			}
			props[p.Name] = propDef
		}

		tool := map[string]interface{}{
			"name":        def.Name,
			"description": def.Description,
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": props,
			},
		}

		if len(def.Required) > 0 {
			tool["inputSchema"].(map[string]interface{})["required"] = def.Required
		}

		result = append(result, tool)
	}

	return result
}

// GenerateMarkdownDocs generates documentation for all tools.
func (r *ToolRegistry) GenerateMarkdownDocs() string {
	var sb strings.Builder

	sb.WriteString("# Tool Reference\n\n")
	sb.WriteString(fmt.Sprintf("Total tools: %d\n\n", len(r.tools)))

	for _, cat := range r.Categories() {
		tools := r.byCategory[cat]
		sb.WriteString(fmt.Sprintf("## %s (%d tools)\n\n", capitalizeFirst(string(cat)), len(tools)))

		for _, def := range tools {
			sb.WriteString(fmt.Sprintf("### %s\n\n", def.Name))
			sb.WriteString(fmt.Sprintf("%s\n\n", def.Description))

			if len(def.Properties) > 0 {
				sb.WriteString("**Parameters:**\n\n")
				for _, p := range def.Properties {
					required := ""
					for _, req := range def.Required {
						if req == p.Name {
							required = " *(required)*"
							break
						}
					}
					sb.WriteString(fmt.Sprintf("- `%s` (%s)%s: %s", p.Name, p.Type, required, p.Description))
					if len(p.Enum) > 0 {
						sb.WriteString(fmt.Sprintf(" Values: %s", strings.Join(p.Enum, ", ")))
					}
					sb.WriteString("\n")
				}
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

// Summary returns a brief summary of the registry.
func (r *ToolRegistry) Summary() string {
	var parts []string
	for _, cat := range r.Categories() {
		parts = append(parts, fmt.Sprintf("%s: %d", cat, len(r.byCategory[cat])))
	}
	return fmt.Sprintf("Registry: %d tools (%s)", len(r.tools), strings.Join(parts, ", "))
}

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	// Handle ASCII only for simplicity (category names are ASCII)
	first := s[0]
	if first >= 'a' && first <= 'z' {
		return string(first-32) + s[1:]
	}
	return s
}
