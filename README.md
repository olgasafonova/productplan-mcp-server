# ProductPlan CLI & MCP Server

A single binary that provides both CLI access and MCP server integration for [ProductPlan](https://www.productplan.com/).

- **CLI mode**: Query ProductPlan directly from terminal
- **MCP server mode**: Integrate with Claude Code, Cursor, VS Code, etc.
- **No dependencies**: Single ~5MB binary
- **Cross-platform**: macOS, Linux, Windows
- **Token-optimized**: Consolidated tools and summarized responses for efficient AI usage

## What is ProductPlan?

[ProductPlan](https://www.productplan.com/) is a roadmap software used by product teams to plan, visualize, and communicate strategy. Over 4,000 companies use it to align teams around product direction.

**Core features:**
- **Roadmaps** - Visual timelines with bars representing initiatives, organized into lanes (themes, teams, or categories)
- **OKRs** - Strategic objectives and measurable key results to track progress
- **Discovery** - Capture and prioritize ideas before they hit the roadmap
- **Launches** - Coordinate go-to-market activities with checklists and tasks

## What is MCP?

[Model Context Protocol (MCP)](https://modelcontextprotocol.io/) is an open standard for connecting AI assistants to external tools and data sources. Anthropic developed it; OpenAI, Google, and others are adopting it.

This server lets AI assistants interact with your ProductPlan data through natural language:

```
You: "What's on our Q1 roadmap?"
AI: [queries ProductPlan API, returns roadmap items]

You: "Show me all objectives that are behind schedule"
AI: [fetches OKRs, filters by status]

You: "Create a new idea for mobile app improvements"
AI: [creates idea in ProductPlan Discovery]
```

## Supported Tools

| Tool | Support |
|------|---------|
| Terminal (CLI) | ✅ Direct |
| Claude Code | ✅ Native |
| Cursor | ✅ Native |
| Claude Desktop | ✅ Native |
| VS Code + Cline | ✅ Via extension |
| VS Code + Continue | ✅ Via extension |
| VS Code + Roo Code | ✅ Via extension |
| n8n | ✅ Native MCP Client |

## Installation

### Download Binary

Download from [Releases](https://github.com/olgasafonova/productplan-mcp-server/releases):

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `productplan-darwin-arm64` |
| macOS (Intel) | `productplan-darwin-amd64` |
| Linux (x64) | `productplan-linux-amd64` |
| Linux (ARM) | `productplan-linux-arm64` |
| Windows | `productplan-windows-amd64.exe` |

```bash
# macOS/Linux
chmod +x productplan-*
sudo mv productplan-* /usr/local/bin/productplan
```

### Build from Source

```bash
git clone https://github.com/olgasafonova/productplan-mcp-server.git
cd productplan-mcp-server
go build -o productplan .
```

## Configuration

Get your API token from [ProductPlan Settings → API](https://app.productplan.com/settings/api).

```bash
export PRODUCTPLAN_API_TOKEN="your-api-token"
```

## CLI Usage

```bash
# Check connection
productplan status

# Roadmaps
productplan roadmaps              # List all roadmaps
productplan roadmaps 12345        # Get roadmap details

# Bars, Lanes, Milestones
productplan bars 12345            # List bars in roadmap
productplan lanes 12345           # List lanes in roadmap
productplan milestones 12345      # List milestones in roadmap

# OKRs
productplan objectives            # List all objectives
productplan objectives 67890      # Get objective details
productplan key-results 67890     # List key results for objective

# Discovery
productplan ideas                 # List all ideas
productplan ideas 11111           # Get idea details
productplan opportunities         # List all opportunities
productplan opportunities 22222   # Get opportunity details

# Launches
productplan launches              # List all launches
productplan launches 33333        # Get launch details
productplan tasks 33333           # List tasks for launch

# Organization
productplan users                 # List users
productplan teams                 # List teams
```

## MCP Server Configuration

### Claude Code

Add to `~/.claude.json`:

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

### Cursor

Add to Cursor's MCP settings (Settings → MCP Servers):

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

### Claude Desktop

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

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

### VS Code + Cline

1. Install [Cline](https://marketplace.visualstudio.com/items?itemName=saoudrizwan.claude-dev) extension
2. Add to VS Code settings:

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

### VS Code + Continue

1. Install [Continue](https://marketplace.visualstudio.com/items?itemName=continue.continue) extension
2. Add to `~/.continue/config.json`:

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

### VS Code + Roo Code

1. Install [Roo Code](https://marketplace.visualstudio.com/items?itemName=RooVeterinaryInc.roo-cline) extension
2. Add to settings:

```json
{
  "roo-cline.mcpServers": {
    "productplan": {
      "command": "/usr/local/bin/productplan",
      "env": {
        "PRODUCTPLAN_API_TOKEN": "your-token"
      }
    }
  }
}
```

### n8n

[n8n](https://n8n.io/) has native MCP support via the MCP Client node.

1. Set environment variable on your n8n instance:
   ```
   N8N_COMMUNITY_PACKAGES_ALLOW_TOOL_USAGE=true
   ```

2. Add **MCP Client** node to your workflow

3. Configure connection:
   - **Command**: `/usr/local/bin/productplan`
   - **Environment Variables**: `PRODUCTPLAN_API_TOKEN=your-token`

4. Connect to an **AI Agent** node to enable natural language queries against your ProductPlan data

Example workflow: `Slack Trigger → AI Agent (with MCP Client) → Slack Response`

The AI agent can query roadmaps, create ideas, or fetch OKR status from conversational input.

## Available MCP Tools (15 Consolidated)

v3.0 consolidates 58 individual tools into 15 action-based tools, reducing token consumption by ~74%.

| Tool | Actions | Description |
|------|---------|-------------|
| `roadmaps` | list, get, get_bars, get_comments | Manage roadmaps |
| `lanes` | list, create, update, delete | Manage lanes in roadmaps |
| `milestones` | list, create, update, delete | Manage milestones |
| `bars` | get, create, update, delete, get_children, get_comments, get_connections, get_links | Manage bars |
| `bar_connections` | list, create, delete | Manage bar connections |
| `bar_links` | list, create, delete | Manage external links on bars |
| `ideas` | list, get, create, update, get_customers, get_tags | Manage ideas (Discovery) |
| `opportunities` | list, get, create, update | Manage opportunities |
| `idea_forms` | list, get | View idea forms |
| `objectives` | list, get, create, update, delete | Manage OKRs |
| `key_results` | list, get, create, update, delete | Manage key results |
| `launches` | list, get, create, update, delete | Manage launches |
| `checklist_sections` | list, get, create, update, delete | Manage checklist sections |
| `tasks` | list, get, create, update, delete | Manage launch tasks |
| `organization` | users, teams, status | Organization info |

### Example Tool Usage

```json
// List all roadmaps
{"tool": "roadmaps", "arguments": {"action": "list"}}

// Get specific roadmap
{"tool": "roadmaps", "arguments": {"action": "get", "id": "12345"}}

// Get bars in a roadmap
{"tool": "roadmaps", "arguments": {"action": "get_bars", "id": "12345"}}

// Create a new idea
{"tool": "ideas", "arguments": {"action": "create", "title": "Mobile app redesign", "description": "..."}}
```

### Response Optimization

List operations return summarized responses to reduce token usage:

```json
{
  "count": 17,
  "items": [
    {"id": 498227, "name": "Product Roadmap", "updated_at": "2025-12-05T00:57:59Z"},
    {"id": 592160, "name": "Process Platform", "updated_at": "2025-12-04T22:19:03Z"}
  ]
}
```

## API Coverage

| Feature | Read | Create | Update | Delete |
|---------|------|--------|--------|--------|
| Roadmaps | ✅ | - | - | - |
| Lanes | ✅ | ✅ | ✅ | ✅ |
| Milestones | ✅ | ✅ | ✅ | ✅ |
| Bars | ✅ | ✅ | ✅ | ✅ |
| Bar Comments | ✅ | - | - | - |
| Bar Connections | ✅ | ✅ | - | ✅ |
| Bar Links | ✅ | ✅ | - | ✅ |
| Ideas | ✅ | ✅ | ✅ | - |
| Idea Customers | ✅ | - | - | - |
| Idea Tags | ✅ | - | - | - |
| Idea Forms | ✅ | - | - | - |
| Opportunities | ✅ | ✅ | ✅ | - |
| Objectives | ✅ | ✅ | ✅ | ✅ |
| Key Results | ✅ | ✅ | ✅ | ✅ |
| Launches | ✅ | ✅ | ✅ | ✅ |
| Checklist Sections | ✅ | ✅ | ✅ | ✅ |
| Tasks | ✅ | ✅ | ✅ | ✅ |
| Users | ✅ | - | - | - |
| Teams | ✅ | - | - | - |

## Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Create release archives
make release
```

## Changelog

### v3.0.0
- Consolidated 58 tools into 15 action-based tools (74% reduction)
- Added response summarization for list operations
- Compact JSON responses in MCP mode
- Improved token efficiency for AI assistants

### v2.0.0
- Initial public release
- Full ProductPlan API v2 coverage
- CLI and MCP server modes

## License

MIT License - see [LICENSE](LICENSE) for details.
