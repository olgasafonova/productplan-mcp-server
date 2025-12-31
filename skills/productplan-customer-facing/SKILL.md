---
name: productplan-customer-facing
description: Generates customer-ready roadmap views and launch timelines from ProductPlan for Sales and Customer Success teams. Use when preparing for customer calls, sharing product direction, or answering "when will feature X ship?" questions.
version: "1.0.0"
license: MIT
compatibility: Requires productplan-mcp-server installed and PRODUCTPLAN_API_TOKEN configured
metadata:
  author: olgasafonova
  mcp-server: productplan-mcp-server
  mcp-protocol: "2024-11-05"
  mcp-transport: stdio
  homepage: https://github.com/olgasafonova/productplan-mcp-server
  persona: sales-customer-success
---

# ProductPlan for Sales & Customer Success

Customer-ready roadmap views and release timelines.

## Quick Answers for Customer Calls

### "What's coming soon?"

1. Call `list_launches` to see upcoming releases
2. Filter for launches in the next 30-90 days
3. Present as a timeline customers can understand

### "When will feature X be available?"

1. Call `get_roadmap_complete` for the relevant product
2. Search bars for the feature name
3. Report the end_date as the expected availability

### "What did you recently release?"

1. Call `list_launches`
2. Filter for completed launches
3. Summarize recent releases with key highlights

### "What's on the roadmap for this year?"

1. Call `get_roadmap_complete` for the product
2. Summarize features by quarter
3. Present high-level themes, not granular details

## Roadmap Views

### Get product roadmap overview

1. Call `list_roadmaps` to find the right product
2. Call `get_roadmap_complete` with roadmap_id
3. Format as customer-friendly timeline

### View by lane/category

Lanes represent categories (e.g., "Platform", "Integrations", "UX"):
1. Call `get_roadmap_lanes` to see categories
2. Call `get_roadmap_bars` to get features
3. Group features by lane for themed presentations

### Key milestones

Call `get_roadmap_milestones` to get major dates:
- Beta releases
- GA launches
- Conference announcements

## Launch Information

### Upcoming launches

Call `list_launches` for all scheduled releases.

Key fields for customers:
- Launch name
- Target date
- Status (on track, delayed)

### Launch details

Call `get_launch` with launch_id for:
- Full description
- Included features
- Release notes draft

## Customer Feedback Loop

### Check if idea exists

1. Call `list_ideas`
2. Search for the customer's request
3. If found, share current status and vote count

### Log customer request

Call `manage_idea` with action="create", title, description to capture:
- Feature requests from calls
- Enhancement suggestions
- Pain points mentioned

### Add customer to existing idea

If the idea already exists:
1. Get idea_id from `list_ideas`
2. Call `manage_idea_customer` with action="add", idea_id, customer_name, customer_email

This increases vote count and links the customer for follow-up.

## Presentation Patterns

### Quarterly roadmap summary

Template for customer presentations:
1. Call `get_roadmap_complete`
2. Group features by quarter
3. Format as:
   - **Q1**: [Feature A, Feature B] - Theme
   - **Q2**: [Feature C, Feature D] - Theme

### Release calendar

1. Call `list_launches`
2. Format as monthly calendar:
   - **January**: Launch X, Launch Y
   - **February**: Launch Z

### Feature status check

For specific customer questions:
1. Find the feature in `get_roadmap_bars`
2. Report: Feature name, Lane, Expected date, Current status

## Tool Reference (Read-Focused)

### Roadmap Discovery
- `list_roadmaps` - find the right product roadmap
- `get_roadmap` - basic roadmap info
- `get_roadmap_complete` - full roadmap in one call

### Feature Information
- `get_roadmap_bars` - features on the roadmap
- `get_bar` - single feature details
- `get_roadmap_lanes` - feature categories

### Timeline Data
- `get_roadmap_milestones` - key dates
- `list_launches` - upcoming releases
- `get_launch` - launch details

### Customer Voice
- `list_ideas` - check for existing requests
- `get_idea` - idea details and votes
- `manage_idea` - log new request (action="create")
- `manage_idea_customer` - link customer to idea (action="add")

### System
- `check_status` - verify access before customer call

## Error Handling

### Common errors

| Error | Cause | Solution |
|-------|-------|----------|
| "Invalid API token" | Token expired | Contact admin |
| "Not found" | Roadmap changed | Refresh with list_roadmaps |
| "Rate limited" | Quota exceeded | Wait 60 seconds |

### Timeout guidance

- List operations: 1-3 seconds
- `get_roadmap_complete`: 3-5 seconds
- If slow before a call, run `health_check` first

## Style Guidelines

### DO

- Use customer-friendly language (no internal jargon)
- Present dates as quarters or months ("Q2 2025", "March")
- Lead with what's most relevant to the customer
- Group features by theme or capability
- Include "subject to change" disclaimer for future dates
- Offer to log feedback if customer mentions a need

### DON'T

- Share internal IDs, lane names, or technical details
- Promise exact dates without PM confirmation
- Expose OKRs or internal strategic details
- Show features marked as confidential or internal
- List every minor enhancement; focus on customer value
- Output raw data structures
