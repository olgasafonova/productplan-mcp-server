---
name: productplan-pm
description: Generates comprehensive workflows for product managers to manage roadmaps, OKRs, ideas, and launches in ProductPlan. Use when doing day-to-day product management work, updating feature timelines, tracking objectives, or triaging customer feedback.
version: "1.0.0"
license: MIT
compatibility: Requires productplan-mcp-server installed and PRODUCTPLAN_API_TOKEN configured
metadata:
  author: olgasafonova
  mcp-server: productplan-mcp-server
  mcp-protocol: "2024-11-05"
  mcp-transport: stdio
  homepage: https://github.com/olgasafonova/productplan-mcp-server
  persona: product-manager
---

# ProductPlan for Product Managers

Complete toolkit for day-to-day product management in ProductPlan.

## The ID Chain Pattern

ProductPlan tools require IDs. Always list first, then drill down:

```
list_roadmaps → roadmap_id → get_roadmap_bars → bar_id → get_bar
list_objectives → objective_id → get_objective → key_result data
list_ideas → idea_id → get_idea → customer/tag data
```

## Roadmap Management

### View your roadmap

1. Call `list_roadmaps` to get roadmap IDs
2. Call `get_roadmap_complete` with roadmap_id for all data in one call
   - Returns bars, lanes, and milestones together
   - Faster than separate calls

### Add a new feature

1. Get roadmap_id from `list_roadmaps`
2. Get lane_id from `get_roadmap_lanes` (pick the category)
3. Call `manage_bar` with:
   - action="create"
   - roadmap_id, lane_id
   - name, start_date, end_date (YYYY-MM-DD)

### Reschedule a feature

1. Get bar_id from `get_roadmap_bars`
2. Call `manage_bar` with action="update", bar_id, new start_date and end_date

### Move feature to different lane

1. Get bar_id from `get_roadmap_bars`
2. Get target lane_id from `get_roadmap_lanes`
3. Call `manage_bar` with action="update", bar_id, lane_id

### Create feature dependencies

1. Get both bar IDs from `get_roadmap_bars`
2. Call `manage_bar_connection` with action="create", bar_id (source), target_bar_id

### Read comments on features

Call `get_bar_comments` with bar_id. The ProductPlan API has no endpoint for writing comments.

### Link external resources

Call `manage_bar_link` with action="create", bar_id, url, name. The link is a plain web link: an Azure DevOps or Jira integration link can only be made in the ProductPlan UI.

### Color bars

Bar color comes from the roadmap's legend, set by name.

1. Call `get_roadmap_legends` with roadmap_id to get the legend names
2. Call `manage_bar` with action="update", bar_id, legend="<legend name>" (matched case-insensitively)
3. To remove a color, pass clear_legend=true

Never pass `legend_id`: ProductPlan reads it as "clear the color", so the server refuses it.

### Edit many bars at once

Use the bulk tools for more than a couple of bars (up to 100 per call):

1. Find the bars with `get_roadmap_bars` filters (name_contains, lane, legend, tag, starts_after, ...)
2. Call `bulk_update_bars` with items=[{bar_id}, ...] and a shared `set`, e.g. set={legend:"Committed"}. Run it with dry_run=true first to see the exact payloads
3. Every item is validated before anything is written; the result lists each bar as ok or failed

`bulk_create_bars` (roadmap_id + items) and `bulk_delete_bars` (bar_ids + confirm=true) work the same way.

### Manage lanes

- Create: `manage_lane` with action="create", roadmap_id, name (optional description, position). Lanes have no settable color
- Reorder: `manage_lane` with action="update", lane_id, position
- Delete: `manage_lane` with action="delete", lane_id

### Add milestones

Call `manage_milestone` with action="create", roadmap_id, title, date

## OKR Workflows

### Review quarterly objectives

1. Call `list_objectives` to see all objectives with progress percentages
2. For details, call `get_objective` with objective_id

### Create new objective

Call `manage_objective` with action="create", name, description, time_frame (e.g., "Q1 2025")

### Add key result to objective

Call `manage_key_result` with action="create", objective_id, name, start_value, target_value

### Update key result progress

1. Get key_result_id from `get_objective` or `list_key_results`
2. Call `manage_key_result` with action="update", objective_id, key_result_id, current_value

### Link features to objectives

Features can be linked to key results in the ProductPlan UI to show OKR alignment.

## Idea Management

### Triage incoming ideas

1. Call `list_ideas` to see all ideas with vote counts and status
2. For promising ideas, call `get_idea` for full details
3. Call `list_all_customers` to see the customers on the account (the API has no per-idea customer list)

### Capture new idea

Call `manage_idea` with action="create", title, description

### Tag ideas and link customers

Call `list_all_tags` or `list_all_customers` to see what exists. Adding a tag or a customer to an idea is not possible through the tools: the ProductPlan API has no endpoint for it, so do it in the ProductPlan UI.

### Promote idea to opportunity

1. Call `manage_opportunity` with action="create", problem_statement, description
2. Ideas can be linked to opportunities in the ProductPlan UI

### Work with idea forms

- `list_idea_forms` - see available submission forms
- `get_idea_form` - get form details and fields

## Launch Coordination

### View upcoming launches

Call `list_launches` to see all launches with dates and status

### Get launch details

Call `get_launch` with launch_id for full checklist and assignments

## Nested Features

### Create child features

1. Get parent bar_id from `get_roadmap_bars`
2. Call `manage_bar` with action="create", container_bar_id set to the parent bar_id, and starts_on/ends_on (ProductPlan refuses to nest an undated bar). The parent must have is_container=true, and the child's parked state must match the parent's

### View feature hierarchy

Call `get_bar_children` with bar_id to see nested features

## Error Handling

### Common errors

| Error | Cause | Solution |
|-------|-------|----------|
| "Invalid API token" | Token expired | Verify at ProductPlan Settings → API |
| "Not found" | ID stale | Re-run list tool for fresh IDs |
| "Rate limited" | Quota exceeded | Wait 60 seconds, retry |
| "Permission denied" | Access restricted | Check role permissions |

### Timeout guidance

- List operations: 1-3 seconds typical
- `get_roadmap_complete`: 3-5 seconds for large roadmaps
- Operations exceeding 10 seconds: check network connectivity
- Use `health_check` with deep=true to diagnose issues

### Recovery steps

1. On error, call `check_status` to verify authentication
2. If auth fails, reconfigure API token
3. For "not found" errors, re-fetch IDs from list tools

## Style Guidelines

### DO

- Show features in tables: Name, Lane, Start Date, End Date, Status
- Display OKR progress as percentages
- Format dates as human-readable (March 15, 2025)
- Summarize lists over 10 items
- Include bar_id when referencing features for follow-up actions

### DON'T

- Output raw JSON to users
- Show IDs without context
- List more than 10 items without offering to filter
- Mix YYYY-MM-DD and human-readable dates in same response

## Tool Reference

50 tools: 35 read, 15 write.

### Roadmap Tools
- list_roadmaps, get_roadmap, get_roadmap_complete, get_roadmap_comments
- get_roadmap_bars, get_roadmap_lanes, get_roadmap_milestones, get_roadmap_legends
- manage_bar, manage_lane, manage_milestone

### Bar Tools
- get_bar, get_bar_children, get_bar_comments, get_bar_connections, get_bar_links
- manage_bar_connection, manage_bar_link
- bulk_update_bars, bulk_create_bars, bulk_delete_bars

### OKR Tools
- list_objectives, get_objective, list_key_results, get_key_result
- manage_objective, manage_key_result

### Idea Tools
- list_ideas, get_idea, list_all_customers, list_all_tags
- list_opportunities, get_opportunity
- list_idea_forms, get_idea_form
- manage_idea, manage_opportunity

### Launch Tools
- list_launches, get_launch, get_launch_sections, get_launch_section, get_launch_tasks, get_launch_task
- manage_launch, manage_launch_section, manage_launch_task

### Utility Tools
- check_status, health_check, list_users, list_teams
