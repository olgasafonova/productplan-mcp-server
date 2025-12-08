# Changelog

All notable changes to the ProductPlan MCP Server are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
