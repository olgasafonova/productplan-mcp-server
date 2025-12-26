# Changelog

All notable changes to the ProductPlan MCP Server are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
- MCP protocol version 2024-11-05
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
