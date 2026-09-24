# Changelog

All notable changes to the ProductPlan MCP Server are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- **List calls no longer stop at the first page.** Every collection endpoint returns `{results, paging}` and the API documents a default `page_size` of 200. All list tools issued one bare GET, so a roadmap with more than 200 bars (or an account with more than 200 ideas, users, ...) was silently truncated. `GetList` (`internal/api/list.go`) now requests `page_size=500`, follows `paging.page_count` with bounded concurrency, and fails the whole call if any page fails. It stops at 50 pages and says so: the payload carries `incomplete`, the upstream `record_count`, and a note that the summary repeats.
- **`get_roadmap_bars` discarded the lanes error** (`lanes, _ :=` at `internal/api/endpoints.go:47`, an Article IV violation). A lanes failure is now logged and reported as a warning in the summary while the bars still return; a bars failure fails the call.
- **Bar projection matched a shape the API does not send.** It unmarshalled a bare array where the API returns an envelope, so projection failed and the raw, uncapped bars leaked through; it also read `start_date`/`end_date`/`lane_id`. Bars now project the documented fields: `starts_on`, `ends_on`, `lane_name` (from the bar), `lane_id` (joined from the lane list, never guessed when two lanes share a name), `legend`, `tags`, `percent_done`, `is_container`, `parked`. Description and custom fields stay behind `get_bar`.
- **List summaries were missing for most list tools.** `FormatList` only understood bare arrays, so enveloped and api-projected payloads passed through with no summary, no `No <items> found` message and, for raw envelopes, no 50-item cap (Article V). Both shapes are now summarised and capped. Plurals are fixed (`opportunities`, `launches`).
- **Projections picked fields that do not exist.** Lanes: `color` replaced by `description` and `position`. Milestones: `title` (was `name`). Launches: `launch_date` (was `date`) plus `progress`. Objectives: `risk_status`, `start_date`, `end_date`, `key_results_count` (was `status`, `time_frame`). Tool descriptions for `get_roadmap_lanes`, `get_roadmap_legends`, `list_ideas`, `list_launches` and `list_objectives` now name what is returned.

### Added

- **Filters on list tools, no new tools.** One table in `internal/tools/filters.go` maps friendly args to ProductPlan `q[...]` predicates and generates the schema and sort allowlist. Server-side: `get_roadmap_bars` (`name_contains`, `starts_after`, `starts_before`, `ends_after`, `ends_before`, `is_container`, `sort`), `list_roadmaps` (`name_contains`, `sort`), `list_ideas` (`name_contains`, `channel`, `sort`), `list_opportunities` (`problem_contains`, `workflow_status`, `sort`), `list_launches` (`name_contains`, `status`, `launch_after`, `launch_before`, `sort`). Client-side on `get_roadmap_bars`, applied before the 50-item cap: `lane` (name or ID), `legend`, `tag`. Dates must be real `YYYY-MM-DD` dates; an unknown sort field errors with the allowed list; an empty filtered result reads `No <items> matched the filters`.
- **In-process read cache.** GETs are cached per path+query for 60s (`PRODUCTPLAN_CACHE_TTL`: `90s`, `2m`, or seconds; `0` disables; a malformed value is refused at startup). Concurrent identical GETs share one upstream call. Any POST, PATCH or DELETE clears the whole cache, success or not, and an in-flight read that straddles a write is never stored. `check_status` bypasses the cache. `health_check` reports `cache.{enabled, ttl_seconds, entries, hits, misses, coalesced, invalidations}`.

### Changed

- **Bars and lanes are fetched concurrently**, and `get_roadmap_complete`'s duplicate lanes fetch collapses to one call. Against a server sleeping 20ms per request, `BenchmarkGetRoadmapComplete` measures about 22ms/op for the handler against 108ms/op for the same calls made sequentially.
- **Tuned HTTP transport.** One transport per client, cloned from `http.DefaultTransport`, with HTTP/2 forced, 16 idle connections per host (stdlib default: 2), 90s idle timeout and a 10s TLS handshake timeout. Redirect refusal is unchanged.
- `golang.org/x/sync` is now a direct dependency (`errgroup`, `singleflight`); it was already in `go.sum` as indirect.

## [5.1.0] - 2026-05-03

### Security

- **`manage_*` tools now correctly annotated as Destructive.** Twelve write tools (`manage_lane`, `manage_milestone`, `manage_bar`, `manage_bar_connection`, `manage_bar_link`, `manage_objective`, `manage_key_result`, `manage_idea`, `manage_opportunity`, `manage_launch`, `manage_launch_section`, `manage_launch_task`) all support `action=delete` with documented cascade-delete behavior, but none carried `DestructiveHint=true`. Repo-wide grep for `DestructiveHint` returned a single hit — the type definition. MCP clients that gate user-confirmation on this hint will now prompt before `manage_*` deletions and updates. Closes HG-3 violation.
- **`IdempotentHint=true` no longer blanket-set on `manage_*`.** The previous annotation was a lie: `action=create` twice produces two records; `action=delete` twice 404s on the second call. Retry-aware MCP clients that trust the hint could silently duplicate writes on network blips. Annotation now omitted from `manage_*` tools so clients treat them as non-idempotent (safe default).
- **Path-injection hardening across 40+ API endpoint methods.** Every method that interpolated user-supplied IDs into URL paths (`bars`, `launches`, `objectives`, `key_results`, `ideas`, `opportunities`, `roadmaps`, `lanes`, `milestones`, `launch_sections`, `launch_tasks`, etc.) now validates the ID against `^[A-Za-z0-9_-]+$` and `url.PathEscape`-s it before concatenation. Before: a prompt-injected agent could send `bar_id="../../strategy/objectives/SECRET"` and the request silently pivoted to a different resource via the upstream proxy normalising `..` segments. Now: invalid IDs rejected at the validator with `bar_id contains invalid characters` before any HTTP call. Real ProductPlan IDs (alphanumeric tokens, optionally with `-`/`_`) all match the regex; no legitimate workflow regresses.

These three fixes close violations of patterns graduated 2026-04-25 (HG-3 destructive annotations + path-injection class). Validators in `pkg/productplan/validation.go` (`RequireBarID`, `RequireRoadmapID`, etc.) were previously dead code — now wired in via a new `safeSeg` helper at the API client boundary. Found by an autonomous-vulnerability-research sweep across the MCP portfolio.

## [5.0.1] - 2026-04-26

### Security
- **Sanitize non-JSON response bodies before reaching MCP caller** (HG-2). `pkg/productplan/errors.go` previously stuffed raw response bodies into `APIError.Details`, which then rendered into `Error()` and surfaced verbatim to the MCP caller via `internal/api/client.go:161`. Multi-line HTML error pages and stack traces could flow through unbounded. New `sanitizeBodyForCaller` truncates at first newline and caps at 200 chars; full body remains available to operators via server-side request logging. Surfaced by the [4-axis portfolio audit](https://github.com/olgasafonova/claude-code-config/tree/main/audits/2026-04-26-portfolio-4axis.md).

### Fixed
- Handler default error path, stale test count, and return-value descriptions
- Remove stale tool references, add enum constraints, resolve lint findings

### Changed
- **Internal refactor**: `typedHandler` generic eliminates parse/validate boilerplate across handlers (no behavior change)

### Infrastructure
- `make check` regression target wired into CI
- `go.sum` integrity check + `govulncheck` (advisory) on every CI run
- CODEOWNERS protects workflow files from drive-by PRs
- Auto-dispatch `mcp-registry.yml` from release workflow
- Dependabot groups Go dependency updates (less PR noise)
- Beads issue tracking initialized; `.gitattributes` declares the beads merge driver
- Deslop baseline committed for cloud-routine regression detection

### Dependencies
- `actions/github-script` 8 → 9
- `softprops/action-gh-release` 2 → 3
- Multiple Docker actions (login, setup-buildx, metadata, build-push) on latest majors

## [5.0.0] - 2026-02-21

### Breaking Changes
- **5 tools removed** that called non-existent API endpoints:
  - `manage_bar_comment` — bar comments are read-only in the API
  - `get_idea_customers` — per-idea customer endpoint doesn't exist (use `list_all_customers`)
  - `manage_idea_customer` — per-idea customer mutations don't exist
  - `get_idea_tags` — per-idea tag endpoint doesn't exist (use `list_all_tags`)
  - `manage_idea_tag` — per-idea tag mutations don't exist
- **Field names changed** to match ProductPlan API v2 spec:
  - `manage_bar`: `start_date` → `starts_on`, `end_date` → `ends_on`
  - `manage_milestone`: `name` → `title`
  - `manage_launch_task`: `assignee_id` → `assigned_user_id`, `completed` (bool) → `status` (enum: to_do, in_progress, completed, blocked)
- **Actions removed** from existing tools:
  - `manage_bar_link`: removed `update` action (only `create` and `delete`)
  - `manage_opportunity`: removed `delete` action (only `create` and `update`)
- Tool count: 52 → 47 (35 read, 12 write)

### Fixed
- `get_roadmap_legends` now extracts legends from the roadmap response instead of calling non-existent `/roadmaps/{id}/legends` endpoint

### Added
- Weekly API drift monitoring via GitHub Action (`.github/workflows/api-check.yml`)
- Integration smoke test for all API endpoints (`go test -tags integration`)
- `make check` target (lint + tests) and `make check-api` target
- API endpoint snapshot (`testdata/api-endpoints.json`) for drift detection
- Panic recovery and ToolAnnotations support
- Gosec security linter in golangci-lint config

### Changed
- CI consolidated: `test` and `lint` jobs replaced with single `make check` job

## [4.11.1] - 2025-01-13

### Fixed
- Build configuration now uses `cmd/productplan` entry point, ensuring release binaries include all 52 tools (was incorrectly building from root `main.go` with only 37 tools)

## [4.11.0] - 2025-01-13

### Added
- `get_launch_section` tool to get details of a single launch section
- `get_launch_task` tool to get details of a single launch task

### Improved
- Optimized tool descriptions for token efficiency (shorter, clearer for AI agents)
- Added helper functions (`setIfNotEmpty`, `setIfNotNil`, `setIfNotEmptySlice`) to reduce code duplication
- Extracted `addBarOptionalFields` helper for cleaner bar payload building

### Fixed
- Switch statement default cases now return proper errors instead of silent nil

### Changed
- Tool count increased from 50 to 52 (37 READ + 15 WRITE)

## [4.10.0] - 2025-01-13

### Added
- 11 new tools completing ProductPlan API coverage:
  - **Roadmap**: `get_roadmap_comments` - roadmap-level discussion threads
  - **OKRs**: `get_key_result` - individual key result details
  - **Discovery**: `list_all_customers`, `list_all_tags` - global customer and tag lists
  - **Launches**: `manage_launch` (CRUD), `get_launch_sections`, `manage_launch_section` (CRUD), `get_launch_tasks`, `manage_launch_task` (CRUD) - full launch checklist management
  - **Utility**: `list_users`, `list_teams` - account user and team data

### Changed
- Tool count increased from 39 to 50 (35 READ + 15 WRITE)
- Launches now support full CRUD operations (was read-only)

## [4.9.0] - 2025-01-13

### Added
- `get_roadmap_legends` tool to list available bar colors for a roadmap
- `legend_id` parameter on `manage_bar` to change bar colors
- Additional bar fields on `manage_bar`:
  - `percent_done`: Progress percentage (0-100)
  - `container`: Whether bar is a container for child bars
  - `parked`: Whether bar is parked (not actively scheduled)
  - `parent_id`: Parent bar ID for nesting under containers
  - `strategic_value`: Strategic importance text
  - `notes`: Additional notes
  - `effort`: Effort estimate

### Changed
- Tool count increased from 38 to 39 (27 READ + 12 WRITE)

## [4.8.1] - 2025-12-31

### Improved
- Added "Shared Principles" section to README documenting output conventions across agent skills

## [4.8.0] - 2025-12-31

### Added
- Agent skills for persona-based workflows:
  - `productplan-workflows`: General workflow patterns and tool reference
  - `productplan-pm`: Full PM toolkit with all 36 tools
  - `productplan-leadership`: Strategic cross-roadmap views for executives
  - `productplan-customer-facing`: Customer-ready roadmap info for Sales/CS
- All skills validated with SkillCheck (47-49/50 checks passed)

## [4.7.0] - 2025-12-27

### Added
- `health_check` tool with deep API checks, rate limit status, and cache statistics
- `get_roadmap_complete` tool for parallel fetching of roadmap details, bars, lanes, and milestones
- AI-friendly response formatting with summaries (e.g., "Found 3 roadmaps", "Bar created successfully")
- Comprehensive tests for rate limiter and health checker (100% coverage on health.go)

### Improved
- Tool descriptions with examples, return values, and use cases for better AI understanding

## [4.6.2] - 2025-12-26

### Fixed
- Go module version format (`go 1.24.0` with patch version) for Go proxy compatibility
- Go Report Card now achieves A+ rating with 0 issues across 48 files

## [4.6.1] - 2025-12-26

### Added
- Go Report Card badge in README

### Fixed
- Code formatting with gofmt across all packages

## [4.6.0] - 2025-12-26

### Changed
- Restructured monolithic `main.go` (1750 lines) into clean layered architecture
- Entry point reduced to ~100 lines in `cmd/productplan/main.go`
- Tool dispatch now uses registry pattern instead of 348-line switch statement

### Added
- `internal/api/` package: HTTP client, 40+ endpoint methods, response formatters
- `internal/mcp/` package: JSON-RPC server, registry-based handler dispatch
- `internal/tools/` package: 36 tool handlers organized by domain
- `internal/cli/` package: CLI commands extracted from main
- `internal/logging/` package: structured JSON logger with request IDs
- Integration tests for full MCP protocol flow
- Benchmarks for formatters, handlers, and sessions
- 68 LLM evaluation tests with difficulty levels (easy/medium/hard)
- CI validation script for evaluation suite

### Improved
- Test coverage: all internal packages at 90%+ (mcp: 97%, logging: 97%, api: 95%, cli: 95%, tools: 90%)
- Relocated utilities from `productplan/` to `pkg/productplan/`

### No Breaking Changes
- All 36 tool names unchanged
- All tool schemas identical
- MCP protocol version 2025-11-25
- Environment variable `PRODUCTPLAN_API_TOKEN` unchanged

## [4.2.0] - 2024-12-05

### Added
- Full CRUD support for ideas in the Discovery module
- Full CRUD support for opportunities
- Idea customers management (add/remove)
- Idea tags management (add/remove)
- Read-only access to idea forms

### Changed
- Total tool count increased to 36 (24 READ, 12 WRITE)

## [4.1.0] - 2024-12-04

### Added
- Bar children retrieval (`get_bar_children`)
- Bar comments with create support
- Bar connections with create/delete support
- Bar links with full CRUD support

### Changed
- Total tool count increased to 26

## [4.0.0] - 2024-12-03

### Changed
- Redesigned tool architecture with granular READ tools
- Introduced action-based WRITE tools (`manage_bar`, `manage_lane`, etc.)
- Bars now include lane names in responses for better context

### Added
- 14 granular READ tools
- 5 action-based WRITE tools

## [3.0.0] - 2024-12-02

### Changed
- Consolidated 58 tools into 15 (74% token reduction)
- Added response summarization for cleaner AI outputs

### Improved
- Token efficiency for AI assistant context windows

## [2.0.0] - 2024-12-01

### Added
- Initial public release
- Full ProductPlan API v2 coverage
- Support for roadmaps, bars, lanes, milestones
- OKR management (objectives and key results)
- Launch tracking with checklists
- CLI mode for direct terminal usage
- MCP server mode for AI assistant integration
