# ProductPlan MCP Server

[![CI](https://github.com/olgasafonova/productplan-mcp-server/actions/workflows/ci.yml/badge.svg)](https://github.com/olgasafonova/productplan-mcp-server/actions/workflows/ci.yml)
![lint](https://github.com/olgasafonova/productplan-mcp-server/actions/workflows/lint.yml/badge.svg)
[![CodeScene Average Code Health](https://codescene.io/projects/83043/status-badges/average-code-health)](https://codescene.io/projects/83043)

**Talk to your roadmaps using AI.** Ask questions, create ideas, check OKR progress, and manage launches through natural conversation with Claude, Cursor, or other AI assistants.

## What can you do with this?

Instead of clicking through ProductPlan's interface, just ask:

> "What's on our Q1 roadmap?"

> "Show me all objectives that are behind schedule"

> "Create a new idea for mobile app improvements"

> "What launches are coming up this month?"

> "List all ideas tagged 'customer-request'"

> "Color every eFormidling bar with the Customer satisfaction legend"

> "Move these ten bars to the Platform lane and tag them Q3"

The AI fetches your real ProductPlan data and responds in seconds. Changes to many bars go through one validated call: every bar is checked before anything is written, and you can ask for a dry run first.

## Who is this for?

- **Product Managers** who want faster access to roadmap data
- **Team leads** who need quick status updates without context-switching
- **Anyone using AI assistants** (Claude, Cursor, etc.) who wants ProductPlan integrated into their workflow

No coding required. You'll copy a file and paste some settings.

---

## Quick start (5 minutes)

### Step 1: Get your ProductPlan API token

1. Log into [ProductPlan](https://app.productplan.com)
2. Go to **Settings** → **API** (or visit [this link](https://app.productplan.com/settings/api) directly)
3. Copy your API token

### Step 2: Download the app

Go to the [Releases page](https://github.com/olgasafonova/productplan-mcp-server/releases) and download the right file for your computer:

| Your Computer | Download This |
|---------------|---------------|
| Mac (M1, M2, M3, M4) | `productplan-darwin-arm64` |
| Mac (Intel) | `productplan-darwin-amd64` |
| Windows | `productplan-windows-amd64.exe` |
| Linux | `productplan-linux-amd64` |

**On Mac/Linux**, open Terminal and run these two commands (replace the filename with what you downloaded):

```bash
chmod +x ~/Downloads/productplan-darwin-arm64
sudo mv ~/Downloads/productplan-darwin-arm64 /usr/local/bin/productplan
```

You'll be asked for your password. This is normal.

**On Windows**:

1. Create a folder for the binary (if it doesn't exist):
   ```
   mkdir C:\Tools
   ```

2. Move the downloaded `.exe` to that folder and rename it:
   ```
   move %USERPROFILE%\Downloads\productplan-windows-amd64.exe C:\Tools\productplan.exe
   ```

3. Use the full path `C:\Tools\productplan.exe` in your AI assistant config (shown in Step 3)

> **Note**: You can skip adding to PATH. Just use the full file path in your configuration.

### Step 3: Connect to your AI assistant

Pick the tool you use:

<details>
<summary><strong>Claude Desktop</strong> (click to expand)</summary>

1. Find your config file:
   - **Mac**: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

2. Open it in any text editor and add this (replace `your-token` with your actual API token):

**Mac/Linux:**
```json
{
  "mcpServers": {
    "productplan": {
      "command": "/usr/local/bin/productplan",
      "env": {
        "PRODUCTPLAN_API_TOKEN": "your-token"
      }
    }
  }
}
```

**Windows:**
```json
{
  "mcpServers": {
    "productplan": {
      "command": "C:\\Tools\\productplan.exe",
      "env": {
        "PRODUCTPLAN_API_TOKEN": "your-token"
      }
    }
  }
}
```

3. Restart Claude Desktop

</details>

<details>
<summary><strong>Claude Code (Terminal)</strong></summary>

Add to your config file:
- **Mac/Linux**: `~/.claude.json`
- **Windows**: `%USERPROFILE%\.claude.json`

**Mac/Linux:**
```json
{
  "mcpServers": {
    "productplan": {
      "command": "/usr/local/bin/productplan",
      "env": {
        "PRODUCTPLAN_API_TOKEN": "your-token"
      }
    }
  }
}
```

**Windows:**
```json
{
  "mcpServers": {
    "productplan": {
      "command": "C:\\Tools\\productplan.exe",
      "env": {
        "PRODUCTPLAN_API_TOKEN": "your-token"
      }
    }
  }
}
```

</details>

<details>
<summary><strong>Cursor</strong></summary>

1. Open Cursor
2. Go to **Settings** → **MCP Servers**
3. Add this configuration:

**Mac/Linux:**
```json
{
  "productplan": {
    "command": "/usr/local/bin/productplan",
    "env": {
      "PRODUCTPLAN_API_TOKEN": "your-token"
    }
  }
}
```

**Windows:**
```json
{
  "productplan": {
    "command": "C:\\Tools\\productplan.exe",
    "env": {
      "PRODUCTPLAN_API_TOKEN": "your-token"
    }
  }
}
```

> **Windows users**: Use double backslashes (`\\`) in the path. This is required because backslash is an escape character in JSON.

</details>

<details>
<summary><strong>VS Code + Cline</strong></summary>

1. Install the [Cline extension](https://marketplace.visualstudio.com/items?itemName=saoudrizwan.claude-dev)
2. Open VS Code settings (JSON) and add:

**Mac/Linux:**
```json
{
  "cline.mcpServers": {
    "productplan": {
      "command": "/usr/local/bin/productplan",
      "env": {
        "PRODUCTPLAN_API_TOKEN": "your-token"
      }
    }
  }
}
```

**Windows:**
```json
{
  "cline.mcpServers": {
    "productplan": {
      "command": "C:\\Tools\\productplan.exe",
      "env": {
        "PRODUCTPLAN_API_TOKEN": "your-token"
      }
    }
  }
}
```

</details>

<details>
<summary><strong>VS Code + Continue</strong></summary>

1. Install the [Continue extension](https://marketplace.visualstudio.com/items?itemName=continue.continue)
2. Add to your config file:
   - **Mac/Linux**: `~/.continue/config.json`
   - **Windows**: `%USERPROFILE%\.continue\config.json`

**Mac/Linux:**
```json
{
  "mcpServers": [
    {
      "name": "productplan",
      "command": "/usr/local/bin/productplan",
      "env": {
        "PRODUCTPLAN_API_TOKEN": "your-token"
      }
    }
  ]
}
```

**Windows:**
```json
{
  "mcpServers": [
    {
      "name": "productplan",
      "command": "C:\\Tools\\productplan.exe",
      "env": {
        "PRODUCTPLAN_API_TOKEN": "your-token"
      }
    }
  ]
}
```

</details>

<details>
<summary><strong>n8n (Workflow Automation)</strong></summary>

1. Set environment variable on your n8n instance:
   ```
   N8N_COMMUNITY_PACKAGES_ALLOW_TOOL_USAGE=true
   ```
2. Add an **MCP Client** node to your workflow
3. Configure:
   - **Command**:
     - Mac/Linux: `/usr/local/bin/productplan`
     - Windows: `C:\Tools\productplan.exe`
   - **Environment Variables**: `PRODUCTPLAN_API_TOKEN=your-token`
4. Connect to an **AI Agent** node

Example workflow: `Slack Trigger → AI Agent (with MCP Client) → Slack Response`

</details>

### Step 4: Start asking questions

Open your AI assistant and try:

- "List my ProductPlan roadmaps"
- "What bars are on roadmap [name]?"
- "Show me our OKRs"
- "What ideas are in discovery?"

---

## Real-world use cases

### Morning standup prep
> "Summarize what changed on our Product Roadmap in the last week"

### Stakeholder updates
> "List all Q1 objectives and their progress"

### Idea triage
> "Show me all ideas tagged 'enterprise' that don't have a priority set"

### Launch coordination
> "What tasks are still incomplete for the January launch?"

### Quick lookups
> "When is the 'Mobile App v2' bar scheduled to start?"

---

## What ProductPlan data can you access?

| Feature | View | Create | Edit | Delete |
|---------|------|--------|------|--------|
| **Roadmaps** | Yes | - | - | - |
| **Roadmap Comments** | Yes | - | - | - |
| **Bars** (roadmap items) | Yes | Yes | Yes | Yes |
| **Bulk bar edits** (up to 100 per call) | - | Yes | Yes | Yes |
| **Bar Comments** | Yes | - | - | - |
| **Bar Connections** | Yes | Yes | - | Yes |
| **Bar Links** | Yes | Yes | - | Yes |
| **Lanes** (categories) | Yes | Yes | Yes | Yes |
| **Legends** (bar colors) | Yes | - | Assign to bars | - |
| **Milestones** | Yes | Yes | Yes | Yes |
| **Ideas** (Discovery) | Yes | Yes | Yes | - |
| **Idea Customers** | Yes | - | - | - |
| **Idea Tags** | Yes | - | - | - |
| **Opportunities** | Yes | Yes | Yes | - |
| **Idea Forms** | Yes | - | - | - |
| **Objectives** (OKRs) | Yes | Yes | Yes | Yes |
| **Key Results** | Yes | Yes | Yes | Yes |
| **Launches** | Yes | Yes | Yes | Yes |
| **Launch Sections** | Yes | Yes | Yes | Yes |
| **Launch Tasks** | Yes | Yes | Yes | Yes |
| **Users** | Yes | - | - | - |
| **Teams** | Yes | - | - | - |

List tools take filters, so you can ask for exactly what you need: bars by name, dates, lane, legend or tag, and ideas, opportunities, launches and roadmaps by name or status.

Some things the ProductPlan API itself does not allow, so no tool can do them: creating legends, setting lane colors, writing comments, removing a bar from its container once nested, and creating Azure DevOps or Jira integration links (links made through the API are plain web links).

---

## How it works

```
┌─────────────────┐      spawns       ┌─────────────────┐      API calls     ┌─────────────────┐
│   AI Assistant  │ ───────────────── │   MCP Server    │ ─────────────────▶ │   ProductPlan   │
│ (Claude, Cursor)│ ◀───────────────▶ │   (this binary) │ ◀───────────────── │      API        │
└─────────────────┘   stdin/stdout    └─────────────────┘     JSON data      └─────────────────┘
      your computer                        your computer                         cloud
```

**Why does this need to run on your computer?**

MCP (Model Context Protocol) works through a subprocess model. Your AI assistant doesn't connect to a remote server; it spawns the binary as a local process and communicates via stdin/stdout. This architecture means:

1. **The binary must exist locally** because your AI assistant runs it as a child process
2. **Your API token stays on your machine**, never passing through third-party servers
3. **Real-time, synchronous communication** without network latency between AI and the MCP server
4. **Works offline** for cached data (though ProductPlan API calls still need internet)

When you ask "What's on our Q1 roadmap?", here's what happens:

1. Your AI assistant recognizes it needs ProductPlan data
2. It sends a structured request to the MCP server process
3. The binary translates this into ProductPlan API calls
4. ProductPlan returns JSON data
5. The binary formats and returns results to your AI
6. Your AI presents the answer in natural language

---

## Agent Skills

Pre-built workflow guides that teach AI assistants how to use ProductPlan tools effectively. Each skill targets a specific persona with tailored workflows.

| Skill | Audience | Focus |
|-------|----------|-------|
| [productplan-workflows](skills/productplan-workflows/) | General | Core patterns and tool reference |
| [productplan-pm](skills/productplan-pm/) | Product Managers | Full toolkit: roadmaps, OKRs, ideas, launches |
| [productplan-leadership](skills/productplan-leadership/) | Executives | Portfolio health, cross-roadmap views |
| [productplan-customer-facing](skills/productplan-customer-facing/) | Sales & CS | Customer-ready roadmap timelines |

### Shared Principles

All skills follow these output conventions:
- **No raw JSON** - Format responses as readable text and tables
- **Human-readable dates** - Use "March 2025" or "Q1 2025", not "2025-03-15"
- **Summarize large lists** - Don't overwhelm with 50 items; offer to expand

Persona-specific variations:
- **PM** includes `bar_id` for follow-up actions
- **Leadership** leads with executive summary, hides implementation details
- **Customer-facing** omits internal IDs, lane names, and OKRs entirely

**To use a skill**, copy the `SKILL.md` file to your Claude Code skills directory:

```bash
# Copy a skill (example: PM skill)
cp skills/productplan-pm/SKILL.md ~/.claude/skills/productplan-pm.md
```

Or reference skills directly in your prompts:

> "Use the productplan-pm workflow to show me our Q1 roadmap"

---

## Troubleshooting

**"Command not found" or "spawn ENOENT"**

Your AI assistant can't find the binary. This means:
- **Mac/Linux**: The file isn't at `/usr/local/bin/productplan`, or you forgot to run `chmod +x`
- **Windows**: The path in your config doesn't match where you saved the `.exe`

Fix: Verify the binary exists at the path in your config. Run `ls -la /usr/local/bin/productplan` (Mac/Linux) or check if `C:\Tools\productplan.exe` exists (Windows).

**Windows path issues**

Common mistakes on Windows:

| Wrong | Correct |
|-------|---------|
| `/usr/local/bin/productplan` | `C:\\Tools\\productplan.exe` |
| `C:\Tools\productplan.exe` (single backslash in JSON) | `C:\\Tools\\productplan.exe` |
| `productplan` (no path) | `C:\\Tools\\productplan.exe` |
| Missing `.exe` extension | Include `.exe` in the path |

Windows uses backslashes (`\`) for paths, but JSON treats backslash as an escape character. You must double them (`\\`) in your config file.

**"Invalid API token"**

Double-check your token at [ProductPlan Settings → API](https://app.productplan.com/settings/api). Tokens can expire or be regenerated. Make sure you copied the full token without extra spaces.

**"No roadmaps found"**

Your API token only accesses data you have permission to see in ProductPlan. Check that your account has access to the roadmaps you're looking for.

**AI assistant doesn't see ProductPlan tools**

MCP servers load when your AI assistant starts, not when configs change. After editing your config file, fully quit and restart the application. On Mac, use Cmd+Q (not just closing the window).

**A bar's color didn't change**

Colors are set by legend *name* (for example "Committed"), not by an ID. Ask your assistant to list the roadmap's legends first. Versions before 6.0.0 sent a `legend_id`, which ProductPlan treats as "remove the color"; upgrade if colors keep disappearing.

**"unknown argument" errors**

Since 6.0.0 the server refuses arguments a tool doesn't have, and suggests the closest valid one ("did you mean legend?"). That usually means the assistant guessed a parameter name; it can retry with the suggestion.

**"Permission denied" on Mac/Linux**

The binary needs execute permission. Run:
```bash
chmod +x /usr/local/bin/productplan
```

---

## Command line (optional)

You can also use this tool directly in Terminal without an AI assistant:

```bash
# First, set your token
export PRODUCTPLAN_API_TOKEN="your-token"

# Then run commands
productplan status           # Check connection
productplan roadmaps         # List all roadmaps
productplan bars 12345       # List bars in roadmap #12345
productplan objectives       # List all OKRs
productplan ideas            # List all ideas
productplan opportunities    # List all opportunities
productplan launches         # List all launches
```

**Optional setting:** `PRODUCTPLAN_CACHE_TTL` controls how long read results are cached in memory (default `60s`; `0` turns the cache off). Any change you make through the server clears the cache immediately.

---

## Background info

### What is MCP?

[Model Context Protocol (MCP)](https://modelcontextprotocol.io/) is an open standard that lets AI assistants connect to external tools. Anthropic created it; other AI providers are adopting it. This server implements MCP so your AI assistant can read and write ProductPlan data.

### What is ProductPlan?

[ProductPlan](https://www.productplan.com/) is roadmap software used by 4,000+ product teams. It handles roadmaps, OKRs, idea discovery, and launch coordination.

---

## For Developers

<details>
<summary>Project structure</summary>

```
productplan-mcp-server/
├── cmd/productplan/main.go      # Entry point: token check, then MCP server or CLI
├── internal/
│   ├── api/                     # ProductPlan API client
│   │   ├── client.go            # HTTP client: auth, rate limiting, error wrapping
│   │   ├── transport.go         # Tuned HTTP transport
│   │   ├── cache.go             # In-process TTL read cache (singleflight, write invalidation)
│   │   ├── list.go              # Paged collection GETs and Ransack query encoding
│   │   ├── ids.go               # Typed resource IDs (BarID, RoadmapID, ...), each validated into a path segment
│   │   ├── path.go              # Routes and request paths built only from typed IDs
│   │   ├── safeseg.go           # Path-segment validation for user-supplied IDs
│   │   ├── endpoints*.go        # Endpoint methods (roadmaps, bars, ideas, launches, OKRs)
│   │   ├── bars_read.go         # Roadmap bars with lane enrichment and client-side filters
│   │   ├── bar_schema.go        # Roadmap legends/lanes/custom fields for bar writes
│   │   └── formatters.go        # Response projection for AI
│   ├── mcp/                     # MCP wiring over the official go-sdk
│   │   ├── sdk_server.go        # Serves the registry via go-sdk (stdio)
│   │   ├── sdk.go               # Converts local Tool -> SDK tool
│   │   ├── handler.go           # Registry: dispatch and panic recovery
│   │   ├── argkeys.go           # Rejects undeclared argument keys (did-you-mean)
│   │   ├── editdistance.go      # Levenshtein distance for suggestions
│   │   └── types.go             # Tool-authoring types
│   ├── tools/                   # Tool definitions and handlers
│   │   ├── registry.go          # Tool registration (name -> handler)
│   │   ├── definitions*.go      # Tool schemas and descriptions
│   │   ├── helpers.go           # typedHandler, manage-action dispatch
│   │   ├── filters.go           # List-tool filters -> Ransack predicates
│   │   ├── formatter.go         # List/item/action response summaries
│   │   ├── item_type.go         # Item nouns for summaries
│   │   ├── bar_planner.go       # Validates bar writes against the roadmap
│   │   ├── bar_names.go         # Legend/lane/custom field name resolution
│   │   ├── bulk_bars.go         # bulk_update/create/delete_bars
│   │   ├── roadmaps.go, bars.go, ideas.go, objectives.go, launches.go, utility.go  # handlers
│   │   └── types*.go            # Typed argument structs for handlers
│   ├── cli/                     # CLI commands (status, roadmaps, etc.)
│   │   └── cli.go
│   └── logging/                 # slog JSON handler setup (ts/level/msg)
│       └── logger.go
├── pkg/productplan/             # Reusable utilities
│   ├── retry.go                 # Exponential backoff with jitter
│   ├── ratelimit.go             # Adaptive rate limiting
│   ├── batch.go                 # Batched operations
│   ├── health.go                # Health reporting
│   ├── requestid.go             # Request tracing
│   ├── validation.go            # ID validation (Field.RequireID)
│   └── errors.go                # APIError and error suggestions
└── evals/                       # LLM evaluation test suite
    ├── runner.go, types.go
    ├── tool_selection.json
    ├── confusion_pairs.json
    └── argument_correctness.json
```

</details>

<details>
<summary>Build from source</summary>

Requires Go 1.26 or newer.

```bash
git clone https://github.com/olgasafonova/productplan-mcp-server.git
cd productplan-mcp-server
go build -o productplan ./cmd/productplan
```

Build for all platforms:
```bash
# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o dist/productplan-darwin-arm64 ./cmd/productplan

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o dist/productplan-darwin-amd64 ./cmd/productplan

# Linux
GOOS=linux GOARCH=amd64 go build -o dist/productplan-linux-amd64 ./cmd/productplan

# Windows
GOOS=windows GOARCH=amd64 go build -o dist/productplan-windows-amd64.exe ./cmd/productplan
```

</details>

<details>
<summary>Testing</summary>

Run all tests:
```bash
go test ./...
```

Run with coverage:
```bash
go test ./... -cover
```

Run benchmarks:
```bash
go test ./internal/... -bench=. -benchmem
```

Run evaluation suite:
```bash
./scripts/run-evals.sh
```

**Coverage** (measured 24-09-2026, `go test ./... -cover`):

| Package | Coverage |
|---------|----------|
| internal/logging | 100% |
| pkg/productplan | 94.5% |
| internal/mcp | 92.5% |
| internal/cli | 92.3% |
| evals | 89.5% |
| internal/api | 87.4% |
| internal/tools | 82.3% |
| cmd/productplan | 34.9% |

Every production file scores 10.0 in CodeScene Code Health (two type-only files can't be scored).

</details>

<details>
<summary>MCP tool reference</summary>

50 tools available: 35 READ tools and 15 WRITE tools (12 action-based `manage_*` plus 3 `bulk_*` bar tools):

**Read tools:**
- Roadmaps: `list_roadmaps`, `get_roadmap`, `get_roadmap_bars`, `get_roadmap_lanes`, `get_roadmap_milestones`, `get_roadmap_legends`, `get_roadmap_comments`, `get_roadmap_complete`
- Bars: `get_bar`, `get_bar_children`, `get_bar_comments`, `get_bar_connections`, `get_bar_links`
- OKRs: `list_objectives`, `get_objective`, `list_key_results`, `get_key_result`
- Discovery: `list_ideas`, `get_idea`, `list_all_customers`, `list_all_tags`, `list_opportunities`, `get_opportunity`, `list_idea_forms`, `get_idea_form`
- Launches: `list_launches`, `get_launch`, `get_launch_sections`, `get_launch_section`, `get_launch_tasks`, `get_launch_task`
- Admin: `check_status`, `health_check`, `list_users`, `list_teams`

**Write tools:**
- Roadmaps: `manage_bar`, `manage_lane`, `manage_milestone`
- Bar relationships: `manage_bar_connection`, `manage_bar_link`
- Bulk bars: `bulk_update_bars`, `bulk_create_bars`, `bulk_delete_bars` (up to 100 bars per call, validated up front, `dry_run` supported)
- OKRs: `manage_objective`, `manage_key_result`
- Discovery: `manage_idea`, `manage_opportunity`
- Launches: `manage_launch`, `manage_launch_section`, `manage_launch_task`

Example:
```json
{"tool": "list_roadmaps", "arguments": {}}
{"tool": "manage_bar", "arguments": {"action": "create", "roadmap_id": "123", "lane": "Backend", "name": "New feature", "legend": "Committed"}}
{"tool": "bulk_update_bars", "arguments": {"set": {"legend": "Committed"}, "items": [{"bar_id": "901"}, {"bar_id": "902"}], "dry_run": true}}
{"tool": "manage_idea", "arguments": {"action": "create", "name": "Mobile app improvements"}}
```

</details>

<details>
<summary>Architecture</summary>

The server uses a clean layered architecture:

```
┌──────────────────────────────────────────────────────────────┐
│                        cmd/productplan                        │
│                     (entry point, DI)                         │
└──────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│  internal/cli │    │  internal/mcp │    │internal/tools │
│  (CLI cmds)   │    │  (MCP / SDK)  │    │  (handlers)   │
└───────────────┘    └───────────────┘    └───────────────┘
                              │                     │
                              └──────────┬──────────┘
                                         ▼
                              ┌───────────────────┐
                              │   internal/api    │
                              │  (HTTP client)    │
                              └───────────────────┘
                                         │
                                         ▼
                              ┌───────────────────┐
                              │  ProductPlan API  │
                              └───────────────────┘
```

**Key interfaces:**

```go
// Tool handler interface (internal/mcp)
type Handler interface {
    Handle(ctx context.Context, args map[string]any) (json.RawMessage, error)
}

// Logging (internal/logging): a *slog.Logger with a JSON handler on stderr
logger := logging.New(slog.LevelInfo)
```

**Logging format:**
```json
{"ts":"2026-09-24T10:30:00.123456789Z","level":"debug","msg":"API response","endpoint":"/roadmaps/5","status_code":200,"dur_ms":245}
```

</details>

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for release history and detailed changes.

---

## Like This Project?

If this server saved you time, consider giving it a ⭐ on GitHub. It helps others discover the project.

---

## More MCP Servers

Check out my other MCP servers:

| Server | Description | Stars |
|--------|-------------|-------|
| [gleif-mcp-server](https://github.com/olgasafonova/gleif-mcp-server) | Access GLEIF LEI database. Look up company identities, verify legal entities. | ![GitHub stars](https://img.shields.io/github/stars/olgasafonova/gleif-mcp-server?style=flat) |
| [mediawiki-mcp-server](https://github.com/olgasafonova/mediawiki-mcp-server) | Connect AI to any MediaWiki wiki. Search, read, edit wiki content. | ![GitHub stars](https://img.shields.io/github/stars/olgasafonova/mediawiki-mcp-server?style=flat) |
| [miro-mcp-server](https://github.com/olgasafonova/miro-mcp-server) | Control Miro whiteboards with AI. Boards, diagrams, mindmaps, and more. | ![GitHub stars](https://img.shields.io/github/stars/olgasafonova/miro-mcp-server?style=flat) |
| [nordic-registry-mcp-server](https://github.com/olgasafonova/nordic-registry-mcp-server) | Access Nordic business registries. Look up companies across Norway, Denmark, Finland, Sweden. | ![GitHub stars](https://img.shields.io/github/stars/olgasafonova/nordic-registry-mcp-server?style=flat) |
| [tilbudstrolden-mcp](https://github.com/olgasafonova/tilbudstrolden-mcp) | Nordic grocery deal hunting. Find offers, plan meals, track spending. | ![GitHub stars](https://img.shields.io/github/stars/olgasafonova/tilbudstrolden-mcp?style=flat) |

---

## License

MIT License - see [LICENSE](LICENSE)
