# ProductPlan MCP Server

[![CI](https://github.com/olgasafonova/productplan-mcp-server/actions/workflows/ci.yml/badge.svg)](https://github.com/olgasafonova/productplan-mcp-server/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/olgasafonova/productplan-mcp-server?v=1)](https://goreportcard.com/report/github.com/olgasafonova/productplan-mcp-server)

**Talk to your roadmaps using AI.** Ask questions, create ideas, check OKR progress, and manage launches through natural conversation with Claude, Cursor, or other AI assistants.

## What can you do with this?

Instead of clicking through ProductPlan's interface, just ask:

> "What's on our Q1 roadmap?"

> "Show me all objectives that are behind schedule"

> "Create a new idea for mobile app improvements"

> "What launches are coming up this month?"

> "List all ideas tagged 'customer-request'"

The AI fetches your real ProductPlan data and responds in seconds.

## Who is this for?

- **Product Managers** who want faster access to roadmap data
- **Team leads** who need quick status updates without context-switching
- **Anyone using AI assistants** (Claude, Cursor, etc.) who wants ProductPlan integrated into their workflow

No coding required. You'll copy a file and paste some settings.

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
| **Bar Comments** | Yes | Yes | - | - |
| **Bar Connections** | Yes | Yes | - | Yes |
| **Bar Links** | Yes | Yes | Yes | Yes |
| **Lanes** (categories) | Yes | Yes | Yes | Yes |
| **Legends** (bar colors) | Yes | - | - | - |
| **Milestones** | Yes | Yes | Yes | Yes |
| **Ideas** (Discovery) | Yes | Yes | Yes | - |
| **Idea Customers** | Yes | Yes | - | Yes |
| **Idea Tags** | Yes | Yes | - | Yes |
| **Opportunities** | Yes | Yes | Yes | Yes |
| **Idea Forms** | Yes | - | - | - |
| **Objectives** (OKRs) | Yes | Yes | Yes | Yes |
| **Key Results** | Yes | Yes | Yes | Yes |
| **Launches** | Yes | Yes | Yes | Yes |
| **Launch Sections** | Yes | Yes | Yes | Yes |
| **Launch Tasks** | Yes | Yes | Yes | Yes |
| **Users** | Yes | - | - | - |
| **Teams** | Yes | - | - | - |

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
├── cmd/productplan/main.go      # Entry point (~100 lines)
├── internal/
│   ├── api/                     # ProductPlan API client
│   │   ├── client.go            # HTTP client with caching, retry, rate limiting
│   │   ├── endpoints.go         # 40+ API endpoint methods
│   │   └── formatters.go        # Response enrichment for AI
│   ├── mcp/                     # MCP protocol implementation
│   │   ├── server.go            # JSON-RPC server, stdio I/O
│   │   ├── handler.go           # Tool dispatch via registry
│   │   └── types.go             # Protocol types
│   ├── tools/                   # Tool definitions and handlers
│   │   ├── registry.go          # Tool registration and dispatch
│   │   └── types.go             # Typed argument structs for handlers
│   ├── cli/                     # CLI commands (status, roadmaps, etc.)
│   │   └── cli.go
│   └── logging/                 # Structured JSON logging
│       └── logger.go
├── pkg/productplan/             # Reusable utilities
│   ├── cache.go                 # LRU cache with TTL
│   ├── retry.go                 # Exponential backoff with jitter
│   ├── ratelimit.go             # Adaptive rate limiting
│   ├── registry.go              # ToolBuilder for schema generation
│   ├── requestid.go             # Request tracing
│   └── errors.go                # Error suggestions
└── evals/                       # LLM evaluation test suite
    ├── tool_selection.json
    ├── confusion_pairs.json
    └── argument_correctness.json
```

</details>

<details>
<summary>Build from source</summary>

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

**Coverage targets:**

| Package | Coverage |
|---------|----------|
| internal/mcp | 97% |
| internal/logging | 97% |
| internal/api | 95% |
| internal/cli | 95% |
| internal/tools | 90% |

</details>

<details>
<summary>MCP tool reference</summary>

52 tools available: 37 READ tools and 15 WRITE tools (action-based):

**Read tools:**
- Roadmaps: `list_roadmaps`, `get_roadmap`, `get_roadmap_bars`, `get_roadmap_lanes`, `get_roadmap_milestones`, `get_roadmap_legends`, `get_roadmap_comments`, `get_roadmap_complete`
- Bars: `get_bar`, `get_bar_children`, `get_bar_comments`, `get_bar_connections`, `get_bar_links`
- OKRs: `list_objectives`, `get_objective`, `list_key_results`, `get_key_result`
- Discovery: `list_ideas`, `get_idea`, `get_idea_customers`, `get_idea_tags`, `list_all_customers`, `list_all_tags`, `list_opportunities`, `get_opportunity`, `list_idea_forms`, `get_idea_form`
- Launches: `list_launches`, `get_launch`, `get_launch_sections`, `get_launch_section`, `get_launch_tasks`, `get_launch_task`
- Admin: `check_status`, `health_check`, `list_users`, `list_teams`

**Write tools:**
- Roadmaps: `manage_bar`, `manage_lane`, `manage_milestone`
- Bar relationships: `manage_bar_comment`, `manage_bar_connection`, `manage_bar_link`
- OKRs: `manage_objective`, `manage_key_result`
- Discovery: `manage_idea`, `manage_idea_customer`, `manage_idea_tag`, `manage_opportunity`
- Launches: `manage_launch`, `manage_launch_section`, `manage_launch_task`

Example:
```json
{"tool": "list_roadmaps", "arguments": {}}
{"tool": "manage_bar", "arguments": {"action": "create", "roadmap_id": "123", "lane_id": "456", "name": "New feature"}}
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
│  (CLI cmds)   │    │ (JSON-RPC IO) │    │  (handlers)   │
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

// Logger interface (internal/logging)
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
}
```

**Logging format:**
```json
{"ts":"2024-12-26T10:30:00Z","level":"info","req_id":"ab12","op":"get_roadmap_bars","dur_ms":245}
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
| [miro-mcp-server](https://github.com/olgasafonova/miro-mcp-server) | Control Miro whiteboards with AI. 77 tools for boards, diagrams, mindmaps. | ![GitHub stars](https://img.shields.io/github/stars/olgasafonova/miro-mcp-server?style=flat) |

---

## License

MIT License - see [LICENSE](LICENSE)
