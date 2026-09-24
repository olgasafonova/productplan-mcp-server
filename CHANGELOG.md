# Changelog

All notable changes to the ProductPlan MCP Server are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [6.0.0] - 2026-09-24

A major release. The headline is a fix: every "change color" call through earlier versions **removed** the bar's color instead (see Fixed). It also adds bulk bar editing, list filters and a read cache, moves to go-sdk 1.8.0 and Go 1.26, and brings every scorable production file to CodeScene Code Health 10.0.

### Upgrade notes (breaking)

- **Unknown tool arguments are refused.** A call with a parameter the tool doesn't declare now fails with the valid names and a "did you mean" suggestion, instead of silently ignoring it.
- **Refused with an explanation instead of ignored:** `legend_id` (pass `legend` by name), `effort` on `manage_bar`, and `color` on `manage_lane`. `parent_id` and `container` still work as aliases for `container_bar_id` and `is_container`.
- **List responses changed shape.** Every list tool now returns the summary-plus-capped-data wrapper its output schema always declared, and bar fields use ProductPlan's names (`starts_on`, `ends_on`, `lane`, `legend`).
- **`pkg/productplan` exported API changed:** see Removed. Nothing outside this repository is known to import it.
- **Go 1.26 or newer** is required to build from source.

### Fixed

- **Numeric IDs were refused on every tool.** All ID arguments decode as strings, but ProductPlan returns IDs as JSON numbers, so an agent copying `"id": 36935840` from a read into a write got `cannot unmarshal number into ... of type string`. Whole numbers under `*_id` and `*_ids` keys, including inside bulk `items` and `set`, are now converted to strings before decoding. Found in a live round-trip on roadmap 641608, 24-09-2026.
- **`manage_bar` update with `legend_id` wiped the bar's color.** The payload builder forwarded `legend_id`, which the ProductPlan API reads as a request to clear the legend: `PATCH {legend_id:"1"}` answers 204 and sets the legend to null (probed live 24-09-2026). Every "change color" call through this server removed the color instead. Bars are colored by legend **name** via `legend`; `legend_id` is now rejected with the roadmap's valid legend names and is never forwarded.
- **`manage_bar` sent four other fields the API ignores or refuses.** `parent_id` and `container` were silently ignored (the documented fields are `container_bar_id` and `is_container`), `effort` is not a bar field, and custom fields sent as `{name,value}` got 422 `label must be present`. `parent_id` and `container` now map to the documented fields, custom fields go out as `{label,value}` (`name` still accepted as an alias), and `effort` is rejected with a pointer to custom fields rather than dropped silently.
- **`manage_bar` create returned the wrong thing.** ProductPlan answers `POST /bars` with `{"location":"/api/v2/bars/<id>"}`, not the bar. The ID is now parsed from `location` and the bar is read back.
- **`get_roadmap_legends` described data that does not exist.** Its description promised "ID, name, and hex color" and its formatter picked `id`/`label`/`color` from objects; the API returns bare legend names. It now returns the names, says so, and tells the caller to pass one as `legend`.
- **List calls no longer stop at the first page.** Every collection endpoint returns `{results, paging}` and the API documents a default `page_size` of 200. All list tools issued one bare GET, so a roadmap with more than 200 bars (or an account with more than 200 ideas, users, ...) was silently truncated. `GetList` (`internal/api/list.go`) now requests `page_size=500`, follows `paging.page_count` with bounded concurrency, and fails the whole call if any page fails. It stops at 50 pages and says so: the payload carries `incomplete`, the upstream `record_count`, and a note that the summary repeats.
- **`get_roadmap_bars` discarded the lanes error** (`lanes, _ :=` at `internal/api/endpoints.go:47`, an Article IV violation). A lanes failure is now logged and reported as a warning in the summary while the bars still return; a bars failure fails the call.
- **Bar projection matched a shape the API does not send.** It unmarshalled a bare array where the API returns an envelope, so projection failed and the raw, uncapped bars leaked through; it also read `start_date`/`end_date`/`lane_id`. Bars now project the documented fields: `starts_on`, `ends_on`, `lane_name` (from the bar), `lane_id` (joined from the lane list, never guessed when two lanes share a name), `legend`, `tags`, `percent_done`, `is_container`, `parked`. Description and custom fields stay behind `get_bar`.
- **List summaries were missing for most list tools.** `FormatList` only understood bare arrays, so enveloped and api-projected payloads passed through with no summary, no `No <items> found` message and, for raw envelopes, no 50-item cap (Article V). Both shapes are now summarised and capped. Plurals are fixed (`opportunities`, `launches`).
- **Projections picked fields that do not exist.** Lanes: `color` replaced by `description` and `position`. Milestones: `title` (was `name`). Launches: `launch_date` (was `date`) plus `progress`. Objectives: `risk_status`, `start_date`, `end_date`, `key_results_count` (was `status`, `time_frame`). Tool descriptions for `get_roadmap_lanes`, `get_roadmap_legends`, `list_ideas`, `list_launches` and `list_objectives` now name what is returned.
- **`manage_lane` sent `color`, which the lane endpoints do not accept.** `POST`/`PATCH /roadmaps/{id}/lanes` take `name`, `description` and `position` only. `manage_lane` now sends those three (`description` and `position` are new arguments); `color` stays declared but is rejected with an explanation instead of being sent and ignored.
- **`api.IsNotFound` matched message text.** `handleResponse` flattened `*productplan.APIError` with `%s` when appending a suggestion; it now wraps with `%w` (same message) and `IsNotFound` uses `errors.As` only.

### Added

- **Three bulk bar tools** (tool count 47 → 50): `bulk_update_bars`, `bulk_create_bars`, `bulk_delete_bars`. Up to 100 bars per call. A shared `set` object applies to every item, so "color these 30 bars Committed" is one small call. All items are validated before any write (one roadmap fetch per distinct roadmap), `dry_run:true` returns the exact payloads without writing, writes run 4 at a time under the existing adaptive rate limiter, and each call returns per-item `{index, bar_id, ok, error}` with a summary such as "Updated 28 of 30 bars; 2 failed". The result is an error only when nothing succeeded; a partial failure is a normal result, because some writes landed and resending a whole create would duplicate them. `bulk_delete_bars` requires `confirm:true` and verifies each delete by reading the bar back and expecting 404. Annotations are explicit: update and delete are destructive, create is not, all three are open-world, and none claims idempotency (Article VIII). Concurrency reuses `pkg/productplan/batch.go`'s `Execute`, which nothing had called until now.
- **`manage_bar` accepts the documented contract:** `legend` (name), `lane` (name, alongside `lane_id`), `is_container`, `container_bar_id`. `legend:""` or `clear_legend:true` clears the color (sent as JSON null); omitted fields are never sent.
- **Names are validated before writing.** Legend, lane, custom text field labels, and dropdown values are checked against the roadmap (one `GET /roadmaps/{id}` per call) and matched case-insensitively; an invalid name fails with the list of valid ones, and nothing is sent.
- **Dated bars land on the timeline.** ProductPlan parks new bars by default; a create with `starts_on` and `ends_on` and no `parked` now defaults to `parked:false`. A nested bar inherits its container's parked state, because the API requires them to match.
- **Nesting pre-checks** for the API's 422 cases: `container_bar_id` on a bar without both dates, and a parked state that differs from the container's, fail early with the fix named.
- `get_roadmap` description now points agents at the embedded `custom_text_fields` and `custom_dropdown_fields` (with `allowed_values`) that bar writes are validated against.
- **Filters on list tools, no new tools.** One table in `internal/tools/filters.go` maps friendly args to ProductPlan `q[...]` predicates and generates the schema and sort allowlist. Server-side: `get_roadmap_bars` (`name_contains`, `starts_after`, `starts_before`, `ends_after`, `ends_before`, `is_container`, `sort`), `list_roadmaps` (`name_contains`, `sort`), `list_ideas` (`name_contains`, `channel`, `sort`), `list_opportunities` (`problem_contains`, `workflow_status`, `sort`), `list_launches` (`name_contains`, `status`, `launch_after`, `launch_before`, `sort`). Client-side on `get_roadmap_bars`, applied before the 50-item cap: `lane` (name or ID), `legend`, `tag`. Dates must be real `YYYY-MM-DD` dates; an unknown sort field errors with the allowed list; an empty filtered result reads `No <items> matched the filters`.
- **In-process read cache.** GETs are cached per path+query for 60s (`PRODUCTPLAN_CACHE_TTL`: `90s`, `2m`, or seconds; `0` disables; a malformed value is refused at startup). Concurrent identical GETs share one upstream call. Any POST, PATCH or DELETE clears the whole cache, success or not, and an in-flight read that straddles a write is never stored. `check_status` bypasses the cache. `health_check` reports `cache.{enabled, ttl_seconds, entries, hits, misses, coalesced, invalidations}`.
- **Recovered panics carry a reference.** A panicking tool now returns `internal error in <tool> (ref <16 hex>)`, and the same ref is logged with the panic value and stack. The panic value still never reaches the caller.

### Changed

- **SEP-2549 cache hints now set through the SDK's own hook.** go-sdk v1.8.0 added `ServerOptions.SetCacheable`, which replaces the `mcpcache` receiving middleware used on v1.7.0 to work around the SDK leaving `ttlMs` at 0 ("immediately stale"). Wire behaviour is unchanged: `tools/list` and `server/discover` still advertise a 30-minute `ttlMs` with `cacheScope: "public"`. The `github.com/olgasafonova/mcp-cache-go` dependency is dropped.
- **Minimum Go version is now 1.26.** `go.mod` declares `go 1.26.0`, the oldest supported release line; Go 1.25 is out of support.
- **Bars and lanes are fetched concurrently**, and `get_roadmap_complete`'s duplicate lanes fetch collapses to one call. Against a server sleeping 20ms per request, `BenchmarkGetRoadmapComplete` measures about 22ms/op for the handler against 108ms/op for the same calls made sequentially.
- **Tuned HTTP transport.** One transport per client, cloned from `http.DefaultTransport`, with HTTP/2 forced, 16 idle connections per host (stdlib default: 2), 90s idle timeout and a 10s TLS handshake timeout. Redirect refusal is unchanged.
- `golang.org/x/sync` is now a direct dependency (`errgroup`, `singleflight`); it was already in `go.sum` as indirect.
- **Unknown argument keys are rejected.** A tool call carrying a key its schema does not declare now fails before any API request, naming the key, suggesting the closest declared one (edit distance 2 or less) and listing the valid arguments. Previously such keys were silently ignored, so a caller relying on that will now get an error.
- **`pkg/productplan.RequireID(field, value)` is now `productplan.Field(field).RequireID(value)`** (breaking for importers of `pkg/productplan`; none known outside this repository). `Field` names the argument that error messages point at. The check and the error text are unchanged.
- **Endpoint methods take typed IDs.** `internal/api` defines `RoadmapID`, `LaneID`, `MilestoneID`, `BarID`, `ConnectionID`, `LinkID`, `ObjectiveID`, `KeyResultID`, `IdeaID`, `OpportunityID`, `IdeaFormID`, `LaunchID`, `SectionID` and `TaskID`, and every endpoint method takes them instead of strings. A request path is built only by `route.with`, which accepts nothing but these types and validates each one under its own field name, so an unvalidated ID can no longer reach a URL by construction (previously every method had to remember to call `safeSeg`). The client's raw `Request`, `Get`, `Post`, `Patch`, `Delete` and `GetList` methods are unexported; outside `internal/api` the API is reachable only through the typed endpoints. Validation and error text (`bar_id contains invalid characters ...`) are unchanged, and a new test drives `../` through every ID position of every endpoint against a server that must see zero requests.
- **Logging runs on `log/slog`.** `internal/logging` is now a `slog.JSONHandler` on stderr that renames the built-in keys to the ones the server has always written (`ts` in UTC RFC 3339 with nanoseconds, lower-case `level`, `msg`); every logger is a `*slog.Logger`. The hand-rolled `Logger` interface, `JSONLogger`, `Level`/`ParseLevel`, `Field`/`F` and the field constructors are gone; call sites build attributes with `slog` directly, keeping the same keys (`endpoint`, `method`, `status_code`, `dur_ms`, `error`, `tool`). One visible difference: keys now appear in call order (`ts`, `level`, `msg`, then attributes) instead of alphabetically, which JSON consumers do not see.
- Internal refactors: CodeScene Code Health is 10.0 on every scorable production file (56 of 58; `evals/types.go` and `internal/mcp/types.go` hold only type declarations and get no score). Behaviour is unchanged and pinned by the existing tests.
- **The MCP protocol layer is the official go-sdk** (PRs #54-#58, 31-07 to 03-08-2026; first recorded here). The hand-rolled JSON-RPC loop was replaced in five slices: go-sdk added with tool conversion (#54), the SDK took over transport, handshake, version negotiation and `server/discover` (#55), the old protocol layer was deleted, net -1,228 lines (#56), SEP-2549 cache hints were stamped on list results (#57), and the incoherent `idempotentHint` was dropped from the 35 read-only tools (#58).

### Removed

- **`pkg/productplan` validators nothing called:** `RequireNonEmpty`, `RequireRoadmapID`, `RequireBarID`, `RequireLaneID`, `RequireObjectiveID`, `RequireIdeaID`, `RequireAction`, `ValidateDate`, `ValidateColor`, `ValidateURL`, `ValidateEmail`, `GetString`, `GetStringSlice`. Only ID validation was ever used; the server validates dates, actions and names in `internal/tools`.
- **`pkg/productplan` tool registry** (`registry.go`): `ToolRegistry`, `NewToolRegistry`, `ToolDefinition`, `PropertyDef`, `ToolCategory` and its six `Category*` constants, `ToolBuilder` and `NewTool`. Nothing in the repository imported them outside their own tests; the server registers tools through `internal/tools` and `internal/mcp`. External importers of `pkg/productplan` (none known) lose these symbols.

### Infrastructure

- CI runs race tests on a Go 1.26.x / 1.27.x matrix (new `test` job, gated through `all-checks`); `check`, `security` and `build` run on 1.27.x
- `govulncheck` back on `@latest` (v1.8.0), removing the v1.7.0 pin that the 1.25 runner forced
- golangci-lint v2.7.2 → v2.13.2
- Release builds on Go 1.27.x; Docker builder image `golang:1.25-alpine` → `golang:1.27-alpine`
- Committed `server.json` snapshot refreshed from 4.8.2 to 5.1.0, with checksums verified against the v5.1.0 release assets
- Docker runtime image `alpine:3.20` (out of security support since 01-04-2026) → `alpine:3.24`
- New tests: every advertised tool has a handler and no constructor is orphaned; every tool schema declares every argument its handler reads; the missing-token startup exit; the README tool reference against `BuildAllTools`

### Dependencies

- `github.com/modelcontextprotocol/go-sdk` 1.7.0 → 1.8.0 (supersedes Dependabot #65)
- `golang.org/x/oauth2` 0.35.0 → 0.37.0, `golang.org/x/sync` 0.20.0 → 0.23.0, `golang.org/x/sys` 0.41.0 → 0.48.0, `golang.org/x/time` 0.15.0 → 0.16.0, `github.com/segmentio/asm` 1.1.3 → 1.2.1 (all indirect)

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
