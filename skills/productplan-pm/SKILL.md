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

### Add comments to features

1. Get bar_id from `get_roadmap_bars`
2. Call `manage_bar_comment` with action="create", bar_id, content

### Link external resources

Call `manage_bar_link` with action="create", bar_id, url, title

### Manage lanes

- Create: `manage_lane` with action="create", roadmap_id, name, color
- Reorder: `manage_lane` with action="update", lane_id, position
- Delete: `manage_lane` with action="delete", lane_id

### Add milestones

Call `manage_milestone` with action="create", roadmap_id, name, date

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
3. Call `get_idea_customers` to see who requested it

### Capture new idea

Call `manage_idea` with action="create", title, description

### Tag ideas

1. Get idea_id from `list_ideas`
2. Call `manage_idea_tag` with action="add", idea_id, name (creates tag if new)

### Link customer to idea

Call `manage_idea_customer` with action="add", idea_id, customer_name, customer_email

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
2. Call `manage_bar` with action="create" and parent_id set to parent bar_id

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

### Roadmap Tools
- list_roadmaps, get_roadmap, get_roadmap_complete
- get_roadmap_bars, get_roadmap_lanes, get_roadmap_milestones
- manage_bar, manage_lane, manage_milestone

### Bar Tools
- get_bar, get_bar_children, get_bar_comments, get_bar_connections, get_bar_links
- manage_bar_comment, manage_bar_connection, manage_bar_link

### OKR Tools
- list_objectives, get_objective, list_key_results
- manage_objective, manage_key_result

### Idea Tools
- list_ideas, get_idea, get_idea_customers, get_idea_tags
- list_opportunities, get_opportunity
- list_idea_forms, get_idea_form
- manage_idea, manage_idea_customer, manage_idea_tag, manage_opportunity

### Launch Tools
- list_launches, get_launch

### Utility Tools
- check_status, health_check
