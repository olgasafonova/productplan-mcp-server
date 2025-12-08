# ProductPlan MCP Server - API Coverage Roadmap

**Analysis Date:** 2025-12-08
**Current Version:** 4.1.0 (26 tools)
**Target Version:** 5.0.0 (~45-50 tools)
**API Reference:** https://productplan.readme.io/

---

## Current Implementation Status

### Fully Implemented (26 tools)

| Category | Tools | Status |
|----------|-------|--------|
| Roadmaps | `list_roadmaps`, `get_roadmap`, `get_roadmap_bars`, `get_roadmap_lanes`, `get_roadmap_milestones` | ✅ Complete |
| Bars | `get_bar`, `manage_bar` (create/update/delete) | ✅ Complete |
| Bar Children | `get_bar_children` | ✅ Complete |
| Bar Comments | `get_bar_comments`, `manage_bar_comment` | ✅ Complete |
| Bar Connections | `get_bar_connections`, `manage_bar_connection` (create/delete) | ✅ Complete |
| Bar Links | `get_bar_links`, `manage_bar_link` (create/update/delete) | ✅ Complete |
| Lanes | `manage_lane` (create/update/delete) | ✅ Complete |
| Milestones | `manage_milestone` (create/update/delete) | ✅ Complete |
| Objectives | `list_objectives`, `get_objective`, `manage_objective` (CRUD) | ✅ Complete |
| Key Results | `list_key_results`, `manage_key_result` (CRUD) | ✅ Complete |
| Ideas | `list_ideas`, `get_idea` | ⚠️ Read-only |
| Launches | `list_launches`, `get_launch` | ⚠️ Read-only |
| Admin | `check_status` | ⚠️ Partial |

---

## Missing Features (Grouped by Priority)

### ~~Priority 1: Bar Relationships~~ ✅ COMPLETED in v4.1.0

These enable AI to understand roadmap structure and dependencies.

| Endpoint | Method | Path | New Tool |
|----------|--------|------|----------|
| Get child bars | GET | `/bars/{id}/child_bars` | `get_bar_children` |
| Get bar comments | GET | `/bars/{id}/comments` | `get_bar_comments` |
| Create bar comment | POST | `/bars/{id}/comments` | `manage_bar_comment` |
| Get bar connections | GET | `/bars/{id}/connections` | `get_bar_connections` |
| Create bar connection | POST | `/bars/{id}/connections` | `manage_bar_connection` |
| Delete bar connection | DELETE | `/bars/{id}/connections/{conn_id}` | (in manage_bar_connection) |
| Get bar links | GET | `/bars/{id}/links` | `get_bar_links` |
| Create bar link | POST | `/bars/{id}/links` | `manage_bar_link` |
| Update bar link | PATCH | `/bars/{id}/links/{link_id}` | (in manage_bar_link) |
| Delete bar link | DELETE | `/bars/{id}/links/{link_id}` | (in manage_bar_link) |

**Estimated new tools:** 6-8

---

### Priority 2: Discovery Module (Ideas & Opportunities)

Critical for product feedback workflows.

#### Ideas (expand existing)

| Endpoint | Method | Path | New Tool |
|----------|--------|------|----------|
| Create idea | POST | `/discovery/ideas` | `manage_idea` |
| Update idea | PATCH | `/discovery/ideas/{id}` | (in manage_idea) |
| Get idea customers | GET | `/discovery/ideas/{id}/customers` | `get_idea_customers` |
| Add idea customer | POST | `/discovery/ideas/{id}/customers` | `manage_idea_customer` |
| Remove idea customer | DELETE | `/discovery/ideas/{id}/customers/{cust_id}` | (in manage_idea_customer) |
| Get idea tags | GET | `/discovery/ideas/{id}/tags` | `get_idea_tags` |
| Add idea tag | POST | `/discovery/ideas/{id}/tags` | `manage_idea_tag` |
| Remove idea tag | DELETE | `/discovery/ideas/{id}/tags/{tag_id}` | (in manage_idea_tag) |

#### Opportunities (new)

| Endpoint | Method | Path | New Tool |
|----------|--------|------|----------|
| List opportunities | GET | `/discovery/opportunities` | `list_opportunities` |
| Get opportunity | GET | `/discovery/opportunities/{id}` | `get_opportunity` |
| Create opportunity | POST | `/discovery/opportunities` | `manage_opportunity` |
| Update opportunity | PATCH | `/discovery/opportunities/{id}` | (in manage_opportunity) |
| Delete opportunity | DELETE | `/discovery/opportunities/{id}` | (in manage_opportunity) |

#### Idea Forms (read-only)

| Endpoint | Method | Path | New Tool |
|----------|--------|------|----------|
| List idea forms | GET | `/discovery/idea_forms` | `list_idea_forms` |
| Get idea form | GET | `/discovery/idea_forms/{id}` | `get_idea_form` |

**Estimated new tools:** 10-12

---

### Priority 3: Launches Module (Full CRUD + Checklists)

Enables AI-assisted release management.

#### Launches (expand existing)

| Endpoint | Method | Path | New Tool |
|----------|--------|------|----------|
| Create launch | POST | `/launches` | `manage_launch` |
| Update launch | PATCH | `/launches/{id}` | (in manage_launch) |
| Delete launch | DELETE | `/launches/{id}` | (in manage_launch) |

#### Checklist Sections

| Endpoint | Method | Path | New Tool |
|----------|--------|------|----------|
| Get checklist sections | GET | `/launches/{id}/checklist_sections` | `get_launch_sections` |
| Create checklist section | POST | `/launches/{id}/checklist_sections` | `manage_launch_section` |
| Update checklist section | PATCH | `/launches/{id}/checklist_sections/{sec_id}` | (in manage_launch_section) |
| Delete checklist section | DELETE | `/launches/{id}/checklist_sections/{sec_id}` | (in manage_launch_section) |

#### Checklist Tasks

| Endpoint | Method | Path | New Tool |
|----------|--------|------|----------|
| Get section tasks | GET | `/launches/{id}/checklist_sections/{sec_id}/tasks` | `get_section_tasks` |
| Create task | POST | `/launches/{id}/checklist_sections/{sec_id}/tasks` | `manage_section_task` |
| Update task | PATCH | `/launches/{id}/checklist_sections/{sec_id}/tasks/{task_id}` | (in manage_section_task) |
| Delete task | DELETE | `/launches/{id}/checklist_sections/{sec_id}/tasks/{task_id}` | (in manage_section_task) |

**Estimated new tools:** 6-8

---

### Priority 4: Roadmap Comments & Admin

#### Roadmap Comments

| Endpoint | Method | Path | New Tool |
|----------|--------|------|----------|
| Get roadmap comments | GET | `/roadmaps/{id}/comments` | `get_roadmap_comments` |

#### Admin (expose existing code)

| Endpoint | Method | Path | New Tool |
|----------|--------|------|----------|
| List users | GET | `/users` | `list_users` |
| List teams | GET | `/teams` | `list_teams` |

**Estimated new tools:** 3

---

## Implementation Plan

### Phase 1: Bar Relationships
```
Files to modify: main.go
New API methods:
- GetBarChildren(barID)
- GetBarComments(barID)
- CreateBarComment(barID, data)
- GetBarConnections(barID)
- CreateBarConnection(barID, data)
- DeleteBarConnection(barID, connID)
- GetBarLinks(barID)
- CreateBarLink(barID, data)
- UpdateBarLink(barID, linkID, data)
- DeleteBarLink(barID, linkID)

New MCP tools:
- get_bar_children
- get_bar_comments
- manage_bar_comment (action: create)
- get_bar_connections
- manage_bar_connection (action: create, delete)
- get_bar_links
- manage_bar_link (action: create, update, delete)
```

### Phase 2: Discovery Module
```
New API methods:
- CreateIdea(data)
- UpdateIdea(id, data)
- GetIdeaCustomers(ideaID)
- AddIdeaCustomer(ideaID, data)
- RemoveIdeaCustomer(ideaID, custID)
- GetIdeaTags(ideaID)
- AddIdeaTag(ideaID, data)
- RemoveIdeaTag(ideaID, tagID)
- ListOpportunities()
- GetOpportunity(id)
- CreateOpportunity(data)
- UpdateOpportunity(id, data)
- DeleteOpportunity(id)
- ListIdeaForms()
- GetIdeaForm(id)

New MCP tools:
- manage_idea (action: create, update)
- get_idea_customers
- manage_idea_customer (action: add, remove)
- get_idea_tags
- manage_idea_tag (action: add, remove)
- list_opportunities
- get_opportunity
- manage_opportunity (action: create, update, delete)
- list_idea_forms
- get_idea_form
```

### Phase 3: Launches Module
```
New API methods:
- CreateLaunch(data)
- UpdateLaunch(id, data)
- DeleteLaunch(id)
- GetLaunchSections(launchID)
- CreateLaunchSection(launchID, data)
- UpdateLaunchSection(launchID, sectionID, data)
- DeleteLaunchSection(launchID, sectionID)
- GetSectionTasks(launchID, sectionID)
- CreateSectionTask(launchID, sectionID, data)
- UpdateSectionTask(launchID, sectionID, taskID, data)
- DeleteSectionTask(launchID, sectionID, taskID)

New MCP tools:
- manage_launch (action: create, update, delete)
- get_launch_sections
- manage_launch_section (action: create, update, delete)
- get_section_tasks
- manage_section_task (action: create, update, delete)
```

### Phase 4: Comments & Admin
```
New API methods:
- GetRoadmapComments(roadmapID)

Expose existing:
- ListUsers()
- ListTeams()

New MCP tools:
- get_roadmap_comments
- list_users
- list_teams
```

---

## Version Targets

| Version | Tools | Features Added |
|---------|-------|----------------|
| 4.0.0 (current) | 19 | Base implementation |
| 4.1.0 | ~27 | + Bar relationships |
| 4.2.0 | ~39 | + Discovery module |
| 4.3.0 | ~45 | + Launches module |
| 5.0.0 | ~48 | + Comments & Admin |

---

## Testing Checklist

For each new tool, verify:
- [ ] API method works with valid inputs
- [ ] Error handling for invalid IDs
- [ ] Response formatting is consistent
- [ ] MCP tool definition has clear description
- [ ] CLI command works (if applicable)

---

## Notes

- Keep the "action-based" pattern for write tools (create/update/delete in one tool)
- Add response formatting functions for new data types
- Update README.md and help text after each phase
- Update email draft with final tool count before sending
