---
name: productplan-leadership
description: Generates strategic overviews and cross-roadmap insights from ProductPlan for product leaders and executives. Use when reviewing portfolio health, tracking OKR progress across teams, or preparing leadership updates.
version: "1.0.0"
license: MIT
compatibility: Requires productplan-mcp-server installed and PRODUCTPLAN_API_TOKEN configured
metadata:
  author: olgasafonova
  mcp-server: productplan-mcp-server
  mcp-protocol: "2024-11-05"
  mcp-transport: stdio
  homepage: https://github.com/olgasafonova/productplan-mcp-server
  persona: product-leadership
---

# ProductPlan for Product Leadership

Strategic views and portfolio insights for product leaders.

## Portfolio Overview

### See all roadmaps

Call `list_roadmaps` to get a snapshot of all product roadmaps with their status.

### Compare roadmap progress

For each roadmap of interest:
1. Call `get_roadmap_complete` with roadmap_id
2. Compare bar counts, milestone dates, and lane distributions

### Cross-roadmap milestone view

1. Call `list_roadmaps` to get all roadmap IDs
2. For each, call `get_roadmap_milestones`
3. Compile timeline of key dates across products

## OKR Tracking

### Organization-wide OKR health

1. Call `list_objectives` to see all objectives
2. Review progress percentages across objectives
3. Identify objectives below target (< 70% progress if mid-quarter)

### Drill into struggling objectives

1. Call `get_objective` with objective_id
2. Review individual key result progress
3. Identify which key results are lagging

### Quarterly OKR summary

Request pattern:
- "Show me all Q1 2025 objectives and their progress"
- "Which objectives are at risk of missing targets?"
- "Summarize OKR health across all teams"

## Strategic Insights

### Feature pipeline health

1. Call `get_roadmap_complete` for each product
2. Count features by lane (e.g., "In Progress", "Planned", "Shipped")
3. Identify bottlenecks or empty pipelines

### Launch calendar

1. Call `list_launches` to see all upcoming launches
2. Review dates and readiness status
3. Identify launches in the next 30/60/90 days

### Idea backlog insights

1. Call `list_ideas` to see pending ideas
2. Sort by vote count to see customer demand
3. Identify high-value ideas not yet scheduled

## Common Leadership Questions

| Question | Workflow |
|----------|----------|
| "What's shipping this quarter?" | `list_launches` filtered by date |
| "How are our OKRs tracking?" | `list_objectives` → summarize progress |
| "Show me the product portfolio" | `list_roadmaps` → brief for each |
| "What are customers asking for?" | `list_ideas` sorted by votes |
| "Which teams are on track?" | Cross-reference roadmaps with OKRs |

## Reading Tool Reference

Focus on read operations for strategic views:

### Portfolio View
- `list_roadmaps` - all roadmaps at a glance
- `get_roadmap_complete` - full roadmap data in one call

### OKR Health
- `list_objectives` - all objectives with progress
- `get_objective` - key result details for one objective

### Pipeline Insights
- `get_roadmap_bars` - features on a roadmap
- `get_roadmap_milestones` - key dates

### Customer Voice
- `list_ideas` - idea backlog with vote counts
- `list_opportunities` - validated problems

### Launch Readiness
- `list_launches` - upcoming launches
- `get_launch` - launch details and checklist

### System Health
- `check_status` - verify API access
- `health_check` - diagnose connectivity

## Error Handling

### Common errors

| Error | Cause | Solution |
|-------|-------|----------|
| "Invalid API token" | Token expired | Contact admin for new token |
| "Not found" | Data changed | Refresh with list tool |
| "Rate limited" | Quota exceeded | Wait 60 seconds |

### Timeout guidance

- List operations: 1-3 seconds
- `get_roadmap_complete`: 3-5 seconds for large roadmaps
- Cross-roadmap queries: aggregate individual calls

## Style Guidelines

### DO

- Lead with executive summary before details
- Show OKR progress as percentages with status indicators
- Use tables for cross-roadmap comparisons
- Highlight items needing attention (at-risk, delayed)
- Format dates as human-readable (Q1 2025, March 2025)
- Limit initial output to top 5-10 items; offer to expand

### DON'T

- Output raw JSON or technical IDs
- Show granular feature details unless requested
- List every item without summarizing
- Include implementation details (bar_ids, lane configurations)
- Present data without context or recommendations
