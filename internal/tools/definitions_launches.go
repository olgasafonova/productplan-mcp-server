package tools

import (
	"github.com/olgasafonova/productplan-mcp-server/internal/mcp"
)

// launchTools returns launch-related tool definitions.
func launchTools() []mcp.Tool {
	tools := launchCoreTools()
	tools = append(tools, launchSectionTools()...)
	return append(tools, launchTaskTools()...)
}

// launchCoreTools returns core launch tools (list, get, manage launch itself).
func launchCoreTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "list_launches",
			Description: `List all launches. START HERE for launches.

USE WHEN: "Show launches", "Release schedule"
Returns array of launches with ID, name, date, and description.
FAILS WHEN: API token invalid. Returns empty list if no launches exist.`,
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		{
			Name: "get_launch",
			Description: `Get launch details with checklist.

USE WHEN: "Tell me about this launch", "Launch readiness"
FAILS WHEN: launch_id not found (get valid IDs from list_launches).`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"launch_id": {Type: "string", Description: "Launch ID from list_launches"},
				},
				Required: []string{"launch_id"},
			},
		},
		{
			Name: "manage_launch",
			Description: `Create, update, or delete a launch.

USE WHEN: "Create launch", "Update date", "Delete launch"
Actions: create (name+date), update (launch_id), delete (launch_id)
Returns the created/updated launch object, or confirmation on delete.
FAILS WHEN: create without name or date, update/delete without launch_id, date not in YYYY-MM-DD format. WARNING: delete removes the launch and all its sections and tasks.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":      {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"launch_id":   {Type: "string", Description: "Launch ID (for update/delete)"},
					"name":        {Type: "string", Description: "Launch name"},
					"date":        {Type: "string", Description: "YYYY-MM-DD", Pattern: `^\d{4}-\d{2}-\d{2}$`, Examples: []any{"2025-04-01"}},
					"description": {Type: "string", Description: "Description"},
				},
				Required: []string{"action"},
			},
		},
	}
}

// launchSectionTools returns launch checklist section tool definitions.
func launchSectionTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "get_launch_sections",
			Description: `Get checklist sections for a launch.

USE WHEN: "Show sections", "Checklist categories"
For one specific section, use get_launch_section.
FAILS WHEN: launch_id not found.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"launch_id": {Type: "string", Description: "Launch ID"},
				},
				Required: []string{"launch_id"},
			},
		},
		{
			Name: "get_launch_section",
			Description: `Get a specific checklist section by ID.

USE WHEN: "Show one specific checklist section by ID"
For all sections, use get_launch_sections.
FAILS WHEN: launch_id or section_id not found.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"launch_id":  {Type: "string", Description: "Launch ID"},
					"section_id": {Type: "string", Description: "Section ID"},
				},
				Required: []string{"launch_id", "section_id"},
			},
		},
		{
			Name: "manage_launch_section",
			Description: `Create, update, or delete a checklist section.

USE WHEN: "Add Marketing section", "Rename section", "Delete section"
Actions: create (name), update (section_id), delete (section_id)
Returns the created/updated section object, or confirmation on delete.
FAILS WHEN: create without name, update/delete without section_id (get IDs from get_launch_sections). WARNING: delete removes the section and all tasks in it.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":     {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"launch_id":  {Type: "string", Description: "Launch ID"},
					"section_id": {Type: "string", Description: "Section ID (for update/delete)"},
					"name":       {Type: "string", Description: "Section name"},
				},
				Required: []string{"action", "launch_id"},
			},
		},
	}
}

// launchTaskTools returns launch task tool definitions.
func launchTaskTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name: "get_launch_tasks",
			Description: `Get all tasks for a launch.

USE WHEN: "Show tasks", "What needs to be done?"
For one specific task, use get_launch_task.
FAILS WHEN: launch_id not found.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"launch_id": {Type: "string", Description: "Launch ID"},
				},
				Required: []string{"launch_id"},
			},
		},
		{
			Name: "get_launch_task",
			Description: `Get a specific launch task by ID.

USE WHEN: "Show one specific task by ID", "Check task assignment or status"
For all tasks, use get_launch_tasks.
FAILS WHEN: launch_id or task_id not found.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"launch_id": {Type: "string", Description: "Launch ID"},
					"task_id":   {Type: "string", Description: "Task ID"},
				},
				Required: []string{"launch_id", "task_id"},
			},
		},
		{
			Name: "manage_launch_task",
			Description: `Create, update, or delete a launch task.

USE WHEN: "Add task", "Mark complete", "Assign task", "Delete task"
Actions: create (name+section_id), update (task_id), delete (task_id)
Returns the created/updated task object, or confirmation on delete.
FAILS WHEN: create without name or section_id, update/delete without task_id (get IDs from get_launch_tasks). Use list_users to get valid assigned_user_id values.`,
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"action":           {Type: "string", Description: "create, update, or delete", Enum: []string{"create", "update", "delete"}},
					"launch_id":        {Type: "string", Description: "Launch ID"},
					"task_id":          {Type: "string", Description: "Task ID (for update/delete)"},
					"section_id":       {Type: "string", Description: "Section ID (for create)"},
					"name":             {Type: "string", Description: "Task name"},
					"description":      {Type: "string", Description: "Task description"},
					"due_date":         {Type: "string", Description: "YYYY-MM-DD", Pattern: `^\d{4}-\d{2}-\d{2}$`},
					"assigned_user_id": {Type: "string", Description: "User ID to assign (get from list_users)"},
					"status":           {Type: "string", Description: "Task status", Enum: []string{"to_do", "in_progress", "completed", "blocked"}},
				},
				Required: []string{"action", "launch_id"},
			},
		},
	}
}
