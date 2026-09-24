# ProductPlan MCP Server Constitution

This document holds the governance articles for the ProductPlan MCP Server. These articles are **non-negotiable** and **not subject to per-feature override**. They apply to every commit, pull request, and release regardless of urgency or scope.

This document does not change without an explicit constitutional amendment: a dedicated pull request that modifies only this file, reviewed by the maintainer. A feature pull request that would violate an article does not get an exception; it either changes to comply, or it waits behind an amendment.

**Every article below codifies something the repository already does.** No article invents a new requirement. Each names the file or pattern it is drawn from, and each states honestly whether a linter, a test, or a CI job enforces it, or whether it rests on review alone. An article that claims enforcement it does not have is worse than one that admits it has none, because the false claim stops anyone from adding the missing check.

Written 27-08-2026 against `main` at go-sdk v1.7.0, 47 registered tools.

---

## Article I: Tool registration is declarative and single-entry

Adding a tool means adding one `mcp.Tool` to the matching per-category builder (`roadmapTools`, `barTools`, `objectiveTools`, `ideaTools`, `launchTools`, `utilityTools` across `internal/tools/definitions*.go`) and one entry to `handlerConstructors` in `internal/tools/registry.go:36-94`. `BuildAllTools` (`internal/tools/definitions.go:36-62`) assembles the list and runs `annotateTool` over every entry; `RegisterAll` wires each tool to its handler through `createHandler`. Handlers MUST NOT be registered around that path.

Tool names carry no `productplan_` server prefix, deliberately: they are bare operation names (`list_roadmaps`, `manage_bar`). The prefix that is load-bearing is the operation prefix. `annotateTool` and `readOnlyToolName` (`internal/tools/definitions.go:64-116`) derive annotations and output schemas from `get_`/`list_`/`check_`/`manage_`/`health_check`, so a new tool MUST use one of those prefixes or declare explicit annotations.

**Enforcement: mechanically checked, with one hole.** `TestBuildAllTools` pins the count at 50, `TestBuildAllToolsNames` enumerates every expected name, and the per-category counts are pinned by `TestRoadmapTools` (10), `TestBarTools` (11), `TestObjectiveTools` (6), `TestIdeaTools` (10), `TestLaunchTools` (9), `TestUtilityTools` (4), all in `internal/tools/definitions_test.go`. `TestRegisterAll` and `TestCreateHandlerReturnsValidHandlers` in `internal/tools/registry_test.go` spot-check five and six tools respectively. All run in CI via `make check` in `.github/workflows/ci.yml`. `TestEveryToolHasAHandlerAndNoOrphans` walks every name `BuildAllTools` returns through `handlerConstructors` and fails on a missing entry or an orphan constructor (amendment 24-09-2026; this closed the first of the two holes originally named here). The hole that remains: a tool named outside the annotated prefixes gets nil annotations unless it declares its own, and the annotation tests scan by prefix. The three `bulk_*_bars` tools are the live instance; they declare explicit annotations in `internal/tools/definitions_bulk.go`, and `TestBulkToolAnnotations` pins them, but nothing forces the next unprefixed tool to do the same.

---

## Article II: Handlers never panic out

Every tool call dispatches through `Registry.Call` (`internal/mcp/handler.go:83-101`), which uses **named return values** and a deferred `recover()`. A recovered panic is logged server-side with the panic value and full stack, and the caller receives `internal error in <tool>` as an error result. The SDK adapter `handlerFor` (`internal/mcp/sdk_server.go:122-142`) deliberately routes through `Registry.Call` rather than reaching the handler directly, so the recovery cannot be bypassed by construction; the comment at that site cites HG-1 in `rules/code-review-prompts.md`.

Tool failures are reported as an error result with `IsError` set, never as a returned Go error, because a JSON-RPC protocol error tells the caller the request was malformed rather than that the tool ran and failed.

Each recovered panic gets an 8-byte `crypto/rand` hex reference, logged with the panic value and stack, and returned to the caller as `internal error in <tool> (ref <id>)`, so the caller has a token to hand an operator (amendment 24-09-2026; this closed the gap originally named here). `TestRegistryCall_PanicCarriesReference` asserts the reference appears in both the log and the result and that the panic value never reaches the caller.

**Enforcement: mechanically checked.** `TestSDKServerRecoversHandlerPanic` in `internal/mcp/sdk_server_test.go` drives a panicking tool through a real in-memory MCP session and asserts three things: the panic is not a protocol error, not a silent success, and the panic value does not leak to the caller; it then verifies the server still answers calls afterwards. `TestSDKServerReportsToolFailureAsErrorResult` covers the error-result rule. Both run in CI.

---

## Article III: Anything that does I/O takes `context.Context` first

Every endpoint method on `api.Client` across `internal/api/endpoints*.go`, and every handler behind the `mcp.Handler` interface (`Handle(ctx context.Context, args map[string]any)`), accepts `context.Context` as its first parameter and propagates it to the HTTP request via `http.NewRequestWithContext` (`internal/api/client.go:150`).

The only exempt methods are the three in-process accessors that touch no network: `CacheStats`, `RateLimiter` and `SetLogger` (`internal/api/client.go`). `CacheStats` was added 24-09-2026 with the in-process read cache and reads counters only. This exemption is exhaustive. A new method that reaches the ProductPlan API and does not take a context is a violation, not a new exemption.

**Enforcement: none mechanical.** No linter in `.golangci.yml` checks parameter ordering. The `Handler` interface forces the shape on handlers; the client methods rest on review.

---

## Article IV: Errors are never silently discarded

An operation MUST NOT swallow an error. If an error cannot be handled where it occurs it is logged with enough context to identify the failing object, and propagated or surfaced. A deliberate discard is written as `_ =` with the reason apparent at the line (the `resp.Body.Close()` and CLI `Fprintln` discards in `internal/api/client.go:202` and `internal/cli/cli.go` are the pattern).

This article has a receipt in this repository: the 4.11.0 entry in `CHANGELOG.md` records switch-statement default cases that returned silent `nil` instead of errors — a caller sending an unknown `action` got quiet success. The 5.1.0 entry records a second species of the same failure: the ID validators in `pkg/productplan/validation.go` existed as dead code while nothing called them.

**Enforcement: mechanically checked.** `errcheck` is in the enabled linter list in `.golangci.yml`, with `check-type-assertions: true`, and runs on every push and pull request via both `lint.yml` and `make check` in `ci.yml`. Test files are excluded by an explicit rule in `.golangci.yml`; production code is not. The `gosec` exclusion for `G104` carries the comment "Covered by errcheck", and unlike the sister repo where that comment was once false, here it is true. Verified 27-08-2026: `golangci-lint run ./...` reports **0 issues**, and the probe `golangci-lint run --no-config --default=none --enable=errcheck ./...` reports **12 hits, every one in a `_test.go` file, zero in production code**.

---

## Article V: List responses are bounded and honest about truncation, and an empty result says so

Every list response passes through `FormatList` (`internal/tools/formatter.go`), which wraps the payload in a `FormattedResponse` with a human-readable summary. Three behaviours are the contract:

1. **The cap.** `defaultListCap = 50` (`internal/tools/formatter.go:19`). A longer array is clipped and the data re-marshalled so the payload carries only what the summary reports. Raising the cap is a change to this article's terms and belongs in an amendment.
2. **Honest truncation.** A clipped list reports `Showing first 50 of N <items> (refine to narrow)` with the true total, so the caller knows the page is partial.
3. **Explicit zero results.** An empty list reports a literal `No <items> found` rather than a bare empty array. This server implements the zero-result principle from `rules/agent-interface-design.md` principle 5, which most of its sister servers do not.

Bounded scope, stated so the gap is visible: single-item reads (`FormatItem`) and `get_roadmap_complete` pass the upstream payload through verbatim, and the `data` field is deliberately not projected into view types — the `readOutputSchema` comment in `internal/tools/definitions.go:20-23` records that the wrapper is the contract and the per-endpoint body passes through. Projection is not this server's design, and this article does not claim it.

**Enforcement: partially mechanical.** `TestFormatList_CapsLongList` in `internal/tools/formatter_hg2_test.go` pins the cap and the truncation summary; `TestFormatList_Empty` pins the zero-result message. Nothing forces a new list handler to route through `FormatList`; that rests on review.

---

## Article VI: Meaningful exit codes, loud failure on unknown input, and every error suggests a next action

The command-line paths return 0 on success and 1 on failure, consistently: no arguments prints usage and exits 1, an unknown subcommand prints usage and exits 1, a missing required ID prints the subcommand's usage line and exits 1, an API failure prints `Error: ...` to stderr and exits 1 (`internal/cli/cli.go:99-125`). A missing `PRODUCTPLAN_API_TOKEN` is refused at startup with a named-variable error and exit 1 (`cmd/productplan/main.go:35-42,52-55`), not on the tenth request.

Every API error that reaches the caller carries actionable guidance: `APIError.Suggestion()` (`pkg/productplan/errors.go`) maps 401/403/404/400/422 to a concrete next step (404 says to verify the ID via the `list_*` tools; 429 says how long to wait, read from `Retry-After`), and `handleResponse` (`internal/api/client.go:161-173`) appends it to the error text.

Unknown arguments fail loud. `Registry.Call`, the single dispatch point, checks every argument key against the tool's declared `InputSchema` properties before the handler runs. An unknown key is refused with an error that names it, suggests the closest declared key within edit distance 2, and lists the valid keys. Deprecated arguments stay declared so they reach their explanatory errors instead of this generic one: `legend_id`, `effort`, `parent_id` and `container` on `manage_bar`, and `color` on `manage_lane`. Removing a declared argument is breaking under Article XII. (Amendment 24-09-2026. Before it, unknown arguments were silently ignored and an agent that invented a parameter learned nothing.)

**Enforcement: mechanically checked.** `TestCLI_Run_NoArgs`, `TestCLI_Run_UnknownCommand`, and `TestCLI_APIError` in `internal/cli/cli_test.go` assert exit code 1 and the stderr message. `TestAPIError_Suggestion` in `pkg/productplan/errors_test.go` covers the suggestion map. `TestMissingTokenExitsOneAndNamesVariable` and `TestEmptyTokenExitsOne` in `cmd/productplan` cover the missing-token exit. `TestDispatchRejectsUnknownArgumentsForEveryTool` drives the unknown-argument refusal across every registered tool, `TestSchemasDeclareEveryArgTheHandlerReads` proves no schema omits an argument its handler reads, and `TestManageBarKeepsDeprecatedArgsDeclared` pins the deprecated arguments.

---

## Article VII: A tool description is a public contract with the agent

The description is the only thing an agent reads before deciding whether to call a tool. Changing it changes behaviour for every caller, invisibly, with no version bump and no error.

Descriptions follow the established shape: a first line stating what the tool does, a `USE WHEN` clause with example phrasings, and where a confusable sibling exists, a cross-reference of the form "For X, use `other_tool` instead" (`check_status` vs `health_check`, `list_users` vs `list_teams`, and their kin). Every one of the 50 descriptions carries a `USE WHEN` clause today (verified by count across `internal/tools/definitions*.go`). Those cross-references are load-bearing disambiguation and MUST NOT be dropped when a description is shortened; removing one, or renaming a tool, is a breaking change under Article XIII whatever happened to the code behind it.

The evidence that the distinctions are subtle enough to lose in an edit is the eval suite: `evals/confusion_pairs.json` pins 13 confusable pairs across 52 tests, alongside 83 tool-selection tests and 33 argument-correctness tests.

**Enforcement: partially mechanical.** `TestBuildAllToolsHaveDescriptions` checks non-emptiness only. The eval suites are load-validated and count-checked by `evals/runner_test.go` under `go test ./...` in CI. Nothing checks that a description edit preserved its `USE WHEN` clause or its cross-references.

---

## Article VIII: Annotations tell the truth about what a tool does

`ReadOnlyHint`, `DestructiveHint`, and `IdempotentHint` become MCP tool hints that clients use to decide whether to prompt a human or retry a call. A wrong hint costs a user their data.

The rules, all derived centrally in `annotateTool` (`internal/tools/definitions.go:96-116`) rather than per-tool: every `get_`/`list_`/`check_`/`health_check` tool is `ReadOnlyHint=true` and never destructive; every `manage_*` tool is `DestructiveHint=true` because each dispatches across `action=create/update/delete` with documented cascade deletes; and `IdempotentHint` is claimed by **no tool at all** — not on writes, whose idempotency varies per action (`create` twice makes two records, `delete` twice 404s), and not on reads, where the hint is meaningless and misleads retry-aware clients.

This article is incident-born. The 5.1.0 entry in `CHANGELOG.md` records that all twelve `manage_*` tools shipped without `DestructiveHint` (a repo-wide grep found one hit: the type definition) and with a blanket `IdempotentHint=true` that was simply a lie. Commit `9850390` later removed the same lie from the 35 read-only tools.

**Enforcement: mechanically checked; the best-enforced article in this document.** `TestManageToolsHaveDestructiveHint` (the named HG-3 regression test), `TestManageToolsDoNotClaimIdempotency`, and `TestReadOnlyToolsAnnotation` in `internal/tools/annotations_test.go`; `TestConversionAddsBareBoolHintsToAnnotatedTools` in `internal/tools/sdk_conversion_test.go` verifies the hints survive conversion to the SDK's tool type. All run in CI.

---

## Article IX: The API client fails closed in both directions

Nothing agent-supplied reaches the wire unvalidated, and nothing upstream reaches the agent unsanitized. Three mechanisms, each incident-born or graduated from a portfolio audit:

1. **Path-segment validation.** Every agent-supplied ID is one of the typed IDs in `internal/api/ids.go` (`BarID`, `RoadmapID` and 12 more), and request paths can only be built from them (`route.with` in `internal/api/path.go`). Each type validates itself through `safeSeg` (`internal/api/safeseg.go`), which enforces `^[A-Za-z0-9_-]+$` via `productplan.Field.RequireID` and `url.PathEscape`s the result. Since 24-09-2026 an unvalidated ID cannot reach a URL because the code would not compile; before that the rule rested on review. The 5.1.0 entry in `CHANGELOG.md` records why: before this, `bar_id="../../strategy/objectives/SECRET"` pivoted the request to a different resource.
2. **Redirects refused.** The HTTP client sets `CheckRedirect` to `http.ErrUseLastResponse` (`internal/api/client.go:122-124`) — the full-refuse form of the graduated HG-4 rule, appropriate because the ProductPlan API is the only legitimate upstream. Without it, an upstream `Location: http://169.254.169.254/...` would be followed silently.
3. **Response bodies sanitized.** `sanitizeBodyForCaller`/`sanitizeFieldForCaller` in `pkg/productplan/errors.go` strip error-body excerpts at the first newline and cap them at 200 characters before they can reach `Error()` output, on both the non-JSON and the JSON path (HG-2; the 5.0.1 entry records the incident where raw HTML error pages flowed through unbounded).

**Enforcement: mechanically checked.** `TestSafeSeg_RejectsInjection`, `TestEndpoint_RejectsInjectionBeforeHTTP`, and `TestEndpoint_RejectsInjectionInSecondID` in `internal/api/safeseg_test.go`; `TestParseAPIError_NonJSONBodyIsSanitized` and `TestParseAPIError_JSONFieldsAreSanitized` in `pkg/productplan/errors_test.go`. `TestEveryIDEndpoint_RejectsInjectionBeforeHTTP` sends `../` into every ID position of every endpoint and requires zero upstream requests; `TestIDTypes_NameTheirField` and `TestRouteWith_ValidatesEveryID` pin the types. A new endpoint cannot skip validation without building its path outside `route.with`, which review still has to catch.

---

## Article X: The wire surface tells clients the truth about shape and staleness

Two protocol-metadata rules, both about not advertising what the server cannot guarantee:

**Output schemas only where the shape is guaranteed.** Every read-only tool declares the `readOutputSchema` wrapper (`internal/tools/definitions.go:24-33`), making it Code Mode eligible, and the server emits `structuredContent` only for tools that declared a schema and only when the payload is valid JSON (`toolResult`, `internal/mcp/sdk_server.go:161-169`). `manage_*` tools deliberately declare no output schema, because their results vary per action and a schema would be a promise the server cannot keep.

**Cache hints with real TTLs.** SEP-2549 requires `ttlMs` and `cacheScope` on every list result, but the SDK's default leaves `ttlMs` at zero, which the spec reads as "immediately stale" — spec-compliant and useless. This server stamps a real 30-minute TTL on `tools/list` and `server/discover` through the SDK's own `ServerOptions.SetCacheable` hook (`setCacheHints` in `internal/mcp/sdk_server.go`, go-sdk v1.8.0), keeping any TTL a handler already set and leaving the scope to the SDK default of `public`; the number is a judgement about release cadence, recorded in the comment at that site. Until 24-09-2026 this ran through the `mcpcache` receiving middleware from `github.com/olgasafonova/mcp-cache-go`, which the SDK hook replaced; that dependency is gone.

**Enforcement: mechanically checked.** `TestReadOnlyToolsHaveOutputSchema` (pins 35 read tools) and `TestManageToolsHaveNoOutputSchema` in `internal/tools/annotations_test.go`; `TestSDKServerReturnsStructuredContentWhenSchemaDeclared`, `TestSDKServerReturnsTextForToolsWithoutOutputSchema`, and `TestSDKServerStampsCacheHintsOnListTools` and `TestSetCacheHintsPolicy` in `internal/mcp/sdk_server_test.go`. All run in CI.

---

## Article XI: Absent from the spec is not absent from the API, and drift is probed weekly

Tool behaviour is grounded in what the live ProductPlan API actually does, not in what its documentation says. The 5.0.0 entry in `CHANGELOG.md` is the receipt: five tools shipped for releases while calling endpoints that do not exist (`manage_bar_comment`, `get_idea_customers`, `manage_idea_customer`, `get_idea_tags`, `manage_idea_tag`), and `get_roadmap_legends` called a non-existent `/legends` endpoint when the data was in the roadmap response all along. Unit tests against imagined shapes proved nothing.

The standing countermeasure: `.github/workflows/api-check.yml` runs the integration suite (`go test -tags integration -run 'TestAPIEndpoints|TestAPIEndpointDiscovery' ./internal/api/`) against the live API every Monday, checking each endpoint in the `testdata/api-endpoints.json` snapshot, and files a deduplicated `api-drift` GitHub issue on failure.

**Enforcement: mechanical, and conditional on a secret.** The probe needs `PRODUCTPLAN_API_TOKEN`; when the secret is absent, `TestAPIEndpoints` calls `t.Skip` and the workflow goes green having probed nothing (`internal/api/api_integration_test.go:21-24`). A green weekly run is therefore not by itself proof the probes ran. The skip is deliberate — a missing secret should not file a false drift issue — and the cost is this caveat.

---

## Article XII: Semantic versioning, and the changelog is part of the change

The released binary, the tool set, and every tool's arguments are versioned artifacts published to GitHub Releases (five platforms plus `.mcpb` bundles), GHCR (`.github/workflows/docker.yml`), and the MCP Registry (`.github/workflows/mcp-registry.yml`). Breaking changes MUST NOT ship in a patch or minor release.

On this server, "breaking" means what the 5.0.0 entry already treated as major-worthy: removing a tool, renaming a tool or an argument field, removing an action from a `manage_*` tool's enum, changing an argument's type, making an optional argument required, or the description changes named in Article VII. Adding a tool or an optional argument with behaviour-preserving defaults is not breaking.

Every user-visible change is recorded in `CHANGELOG.md` in Keep a Changelog format. The existing entries set the standard: 5.1.0 names the grep that proved the annotation gap, 5.0.1 names the file and line the raw body leaked through. The registry pipeline builds `server.json` fresh at publish time with checksums computed from the actual release assets; the `server.json` committed at the repo root is a snapshot the pipeline does not consume (refreshed to 5.1.0 on 24-09-2026, checksums matched to the v5.1.0 release assets).

Standing violation, named rather than hidden: the go-sdk migration (PRs #54-#58, August 2026), a protocol-layer replacement, is still not recorded in `CHANGELOG.md`. An `[Unreleased]` section now exists (24-09-2026) and covers the SDK 1.8 upgrade, the bar contract fix and the read-path work, but not the migration itself. Recording it is owed before the next release.

**Enforcement: none.** No CI job checks that a pull request touching `internal/tools/definitions*.go` also touched `CHANGELOG.md`, and the gap above shows the discipline is real but fallible.

---

## Article XIII: The supply chain is verified on every pull request

CI runs on every push to `main` and every pull request: `make check` (golangci-lint plus `go test -v -race ./...`) in the `check` job, `gosec` and `govulncheck` in the `security` job, and a tidy check in the `go-mod-tidy` job that diffs **both** `go.mod` and `go.sum` (`git diff --exit-code go.mod go.sum` in `.github/workflows/ci.yml`). The two-file diff is deliberate: a `go get` before the import leaves a stale `// indirect` in `go.mod` that build, test, and lint all pass over, and a `go.sum`-only diff reports clean — the failure that reddened two sister repos on 30-07-2026. `lint.yml` runs `golangci-lint` v2.13.2 independently. Since 24-09-2026 a `test` job also runs `go test -race` on a Go 1.26.x / 1.27.x matrix, gated through the `all-checks` aggregator so the matrix legs need no required contexts of their own; `check` stays a single job so its required context keeps reporting.

Three honest caveats. `govulncheck` is piped to `|| echo "::warning::..."`, so a known vulnerability produces an annotation and a green build; the stated rationale (stdlib findings resolve with Go patch updates) is true, and the cost is that a direct-dependency vulnerability also cannot fail the build. There is no `go mod verify` step, so checksum verification against the module database rests on the Go toolchain's default `GONOSUMCHECK`-off behaviour rather than an explicit CI gate. And the `cache: false # No external dependencies` comments throughout `ci.yml` describe the zero-dependency era that ended with the go-sdk migration in August 2026; they are harmless (they only skip build caching) but no longer true.

**Enforcement: mechanically checked, as described.** Verified locally 24-09-2026: `golangci-lint run ./...` reports 0 issues.

---

## Article XIV: No credentials in version control

API tokens and secrets MUST NOT be committed anywhere in this repository: not in code, not in tests, not in documentation examples. The ProductPlan token is read from `PRODUCTPLAN_API_TOKEN` at startup and never written to disk (`SECURITY.md` "API Token Handling"). Every documentation example uses `your-token`. The two MCP Registry token files that live at the repo root during publishing (`.mcpregistry_github_token`, `.mcpregistry_registry_token`) are excluded by the `.mcpregistry_*` rule in `.gitignore`, alongside `.env*` and `.beads-credential-key`, and `git ls-files` confirms neither is tracked.

**Enforcement: none in CI.** There is no secret scanner in any workflow, and `gosec`'s `G101` hardcoded-credential check is excluded in `.golangci.yml` with the documented reason "False positives on URL constants and test fixtures". The exclusion is defensible and its effect is that nothing automated stands behind this article: it rests on `.gitignore` and on review. The presence of live token files one `.gitignore` edit away from a commit is exactly why this article exists.

---

## Articles considered and rejected

**Test-first development, with coverage requirements as an article.** Coverage here is real and unusually even — the README's developer section quotes 90-97% per package — but there is no `CONTRIBUTING.md` to anchor a process rule, and pinning coverage numbers in a constitution turns a healthy practice into a stale claim the first time a package is added. The README table is the honest form and already lives where it belongs.

**Structured logging everywhere.** Rejected because the repository is mixed on purpose. `internal/logging` provides the structured JSON logger used throughout the server path (a thin `log/slog` setup since 24-09-2026); the CLI prints pretty JSON to stdout for a human, and `main.go` prints friendly errors to stderr. An article would need to carve out every CLI path, at which point it constrains almost nothing.

**Operations that grant durable access fail closed.** Not applicable in this repository: no tool shares, invites, grants a role, or writes a credential. The `manage_*` tools mutate roadmap data only, and Article VIII's destructive annotations plus Article IX's fail-closed client are where that intent actually lands here.

**Defending `pkg/productplan` as a reusable library boundary.** Rejected after checking what the server actually imports: only `AdaptiveRateLimiter`, `ParseAPIError`/`APIError`, and the ID validation (`Field.RequireID` since 24-09-2026, formerly the `Require*` validators). When this was written, five of the package's modules (`retry.go`, `batch.go`, `health.go`, `requestid.go`, and the `registry.go` ToolBuilder) were tested but imported by nothing in `internal/` or `cmd/`. On 24-09-2026 `batch.go` began running the `bulk_*_bars` tools and the ToolBuilder was deleted as dead code, so three remain unwired, and the LRU cache the README still lists at `pkg/productplan/cache.go` was extracted to `github.com/olgasafonova/mcp-cache-go`. An article defending a boundary mostly made of unwired code would codify aspiration, not practice.

**Docs and counts match the code.** A real drift risk with live instances: `evals/README.md` says 10 confusion pairs and 69 selection tests where the suites hold 13 and 83; the CLI usage text still describes "Design (v4.2)" with 24 read tools; the committed `server.json` says 4.8.2 against a 5.1.0 changelog; the README project-structure tree lists three `pkg/` files that no longer exist. Rejected as an article because it is too small for constitutional tier — it belongs as a generator or a test, and the registry pipeline already sidesteps the worst instance by regenerating `server.json` at publish time.

**Minimal attack surface.** Not applicable in the shape the source constitution meant: this binary serves stdio only — `SDKServer.run` is deliberately unexported with a comment saying stdio is the only transport this binary ships — so there is no HTTP listener, health endpoint, or bearer-token surface to bound.

---

## Amendment log

| Date | Change |
|------|--------|
| 27-08-2026 | Ratified. Fourteen articles, adapted from the `CONSTITUTION.md` in `gridctl/gridctl` (Apache-2.0, github.com/gridctl/gridctl). |
| 24-09-2026 | Amendment 1. Article VI now requires unknown tool arguments to be refused (closing its named gap). Articles I and II record their closed gaps (handler-map walk test, panic correlation ID). Article III adds `CacheStats` to the exhaustive exemption list. Article IX records that typed IDs make path validation compiler-enforced. Articles I, VII, X, XII and XIII updated to 50 tools, go-sdk v1.8.0 `SetCacheable`, the refreshed `server.json`, the Go 1.26/1.27 test matrix and golangci-lint v2.13.2. |
