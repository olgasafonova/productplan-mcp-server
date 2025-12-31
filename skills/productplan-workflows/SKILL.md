---
name: productplan-workflows
description: Generates step-by-step workflows for managing ProductPlan roadmaps, OKRs, ideas, and launches via MCP server tools. Use when working with ProductPlan data, planning features, tracking objectives, or coordinating product launches.
version: "1.0.0"
license: MIT
compatibility: Requires productplan-mcp-server installed and PRODUCTPLAN_API_TOKEN configured
metadata:
  author: olgasafonova
  mcp-server: productplan-mcp-server
  mcp-protocol: "2024-11-05"
  mcp-transport: stdio
  homepage: https://github.com/olgasafonova/productplan-mcp-server
---

# ProductPlan Workflows

This skill teaches you how to work effectively with ProductPlan through the MCP server.

## Core Concepts

ProductPlan organizes product work into five areas:

| Area | Purpose | Start With |
|------|---------|------------|
| Roadmaps | Visual timeline of planned features | `list_roadmaps` |
| OKRs | Strategic objectives and key results | `list_objectives` |
| Ideas | Customer feedback and feature requests | `list_ideas` |
| Opportunities | Validated problems worth solving | `list_opportunities` |
| Launches | Release coordination and checklists | `list_launches` |

## The ID Chain Pattern

ProductPlan tools follow a consistent pattern: list first, then drill down.

```
list_roadmaps → roadmap_id → get_roadmap_bars → bar_id → get_bar
list_objectives → objective_id → get_objective → key_result data
list_ideas → idea_id → get_idea → customer/tag data
```

Always start with a list tool to get IDs before calling detail tools.

## Roadmap Workflows

### View roadmap contents

1. Call `list_roadmaps` to see available roadmaps and get IDs
2. Call `get_roadmap_complete` with the roadmap_id for full data in one call
   - This is faster than calling bars/lanes/milestones separately

### Add a feature to the roadmap

1. Get roadmap_id from `list_roadmaps`
2. Get lane_id from `get_roadmap_lanes` (pick the right category)
3. Call `manage_bar` with action="create", roadmap_id, lane_id, name, start_date, end_date

### Move a feature to a different lane

1. Get bar_id from `get_roadmap_bars`
2. Get new lane_id from `get_roadmap_lanes`
3. Call `manage_bar` with action="update", bar_id, lane_id

### Create dependencies between features

1. Get both bar IDs from `get_roadmap_bars`
2. Call `manage_bar_connection` with action="create", bar_id (source), target_bar_id

### Add a milestone

Call `manage_milestone` with action="create", roadmap_id, name, date (YYYY-MM-DD format)

## OKR Workflows

### Review OKR progress

1. Call `list_objectives` to see all objectives with progress percentages
2. For details, call `get_objective` with objective_id to see key results

### Create a new objective

Call `manage_objective` with action="create", name, description, time_frame (e.g., "Q1 2025")

### Update key result progress

1. Get key_result_id from `get_objective` or `list_key_results`
2. Call `manage_key_result` with action="update", objective_id, key_result_id, current_value

## Idea Discovery Workflows

### Triage customer feedback

1. Call `list_ideas` to see all ideas with vote counts and status
2. For promising ideas, call `get_idea` for full details
3. Call `get_idea_customers` to see who requested it

### Capture a new idea

Call `manage_idea` with action="create", title, description

### Tag ideas for categorization

1. Get idea_id from `list_ideas`
2. Call `manage_idea_tag` with action="add", idea_id, name (creates tag if new)

### Link ideas to opportunities

1. Create opportunity with `manage_opportunity` action="create", problem_statement
2. Ideas can be linked to opportunities in the ProductPlan UI

## Error Handling

### Common errors and solutions

| Error | Cause | Solution |
|-------|-------|----------|
| "Invalid API token" | Token expired or incorrect | Verify token at ProductPlan Settings → API |
| "Not found" | ID no longer exists | Re-run list tool to get fresh IDs |
| "Rate limited" | Request quota exceeded | Wait 60 seconds, then retry |
| "Permission denied" | No access to this resource | Check your ProductPlan role permissions |

### Timeout guidance

- List operations: typically complete in 1-3 seconds
- `get_roadmap_complete`: may take 3-5 seconds for large roadmaps
- If operations exceed 10 seconds, check network connectivity
- Use `health_check` with deep=true to diagnose API issues

### Recovery steps

1. On any error, call `check_status` to verify authentication
2. If authentication fails, the API token needs to be reconfigured
3. For "not found" errors, IDs may have changed; re-fetch from list tools

## Common Patterns

### Efficient data fetching

- Use `get_roadmap_complete` instead of multiple calls; it fetches in parallel
- Use `health_check` with deep=true to verify API connectivity
- Use `check_status` to verify authentication before complex operations

### Action-based tools

All write operations use `manage_*` tools with an action parameter:

| Action | Purpose | Required Fields |
|--------|---------|-----------------|
| create | Add new item | Varies by type |
| update | Modify existing | item_id + fields to change |
| delete | Remove item | item_id |

### Date formats

All dates use YYYY-MM-DD format: "2025-03-15"

## Style Guidelines

### DO

- List roadmap items in tables with Name, Lane, Start Date, End Date columns
- Show OKR progress as percentages with visual indicators
- Format dates as human-readable (e.g., "March 15, 2025")
- Summarize large datasets; offer to show details on request
- Use concise, action-oriented language
- Present data in tables when comparing items
- Include the source tool name when referencing data origin

### DON'T

- Output raw JSON responses to users
- Show internal IDs without context
- List more than 10 items without summarizing
- Mix date formats in the same response

## Tool Quick Reference

### Read tools (24 total)

**Roadmaps:** list_roadmaps, get_roadmap, get_roadmap_bars, get_roadmap_lanes, get_roadmap_milestones, get_roadmap_complete

**Bars:** get_bar, get_bar_children, get_bar_comments, get_bar_connections, get_bar_links

**OKRs:** list_objectives, get_objective, list_key_results

**Ideas:** list_ideas, get_idea, get_idea_customers, get_idea_tags, list_opportunities, get_opportunity, list_idea_forms, get_idea_form

**Launches:** list_launches, get_launch

**Utility:** check_status, health_check

### Write tools (12 total)

**Roadmaps:** manage_bar, manage_lane, manage_milestone

**Bar relationships:** manage_bar_comment, manage_bar_connection, manage_bar_link

**OKRs:** manage_objective, manage_key_result

**Ideas:** manage_idea, manage_idea_customer, manage_idea_tag, manage_opportunity
