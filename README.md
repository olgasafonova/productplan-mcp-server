# ProductPlan CLI & MCP Server

A single binary that provides both CLI access and MCP server integration for [ProductPlan](https://www.productplan.com/).

- **CLI mode**: Query ProductPlan directly from terminal
- **MCP server mode**: Integrate with Claude Code, Cursor, VS Code, etc.
- **No dependencies**: Single ~5MB binary
- **Cross-platform**: macOS, Linux, Windows

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

## Available MCP Tools (60+)

### Roadmaps
| MCP Tool | Description |
|----------|-------------|
| `list_roadmaps` | List all roadmaps |
| `get_roadmap` | Get roadmap details |
| `get_roadmap_bars` | Get all bars in a roadmap |
| `get_roadmap_lanes` | Get all lanes in a roadmap |
| `get_roadmap_milestones` | Get all milestones |
| `get_roadmap_comments` | Get roadmap comments |

### Lanes
| MCP Tool | Description |
|----------|-------------|
| `create_lane` | Create a new lane |
| `update_lane` | Update lane properties |
| `delete_lane` | Delete a lane |

### Milestones
| MCP Tool | Description |
|----------|-------------|
| `create_milestone` | Create a new milestone |
| `update_milestone` | Update milestone properties |
| `delete_milestone` | Delete a milestone |

### Bars
| MCP Tool | Description |
|----------|-------------|
| `get_bar` | Get bar details |
| `create_bar` | Create a new bar |
| `update_bar` | Update bar properties |
| `delete_bar` | Delete a bar |
| `get_bar_child_bars` | Get child bars |
| `get_bar_comments` | Get bar comments |
| `get_bar_connections` | Get bar connections |
| `get_bar_links` | Get bar external links |

### Ideas (Discovery)
| MCP Tool | Description |
|----------|-------------|
| `list_ideas` | List all ideas |
| `get_idea` | Get idea details |
| `create_idea` | Create a new idea |
| `update_idea` | Update idea properties |
| `get_idea_customers` | Get idea customers |
| `get_idea_tags` | Get idea tags |

### Opportunities (Discovery)
| MCP Tool | Description |
|----------|-------------|
| `list_opportunities` | List all opportunities |
| `get_opportunity` | Get opportunity details |
| `create_opportunity` | Create an opportunity |
| `update_opportunity` | Update opportunity |

### Idea Forms
| MCP Tool | Description |
|----------|-------------|
| `list_idea_forms` | List idea forms |
| `get_idea_form` | Get form details |

### Objectives (OKRs)
| MCP Tool | Description |
|----------|-------------|
| `list_objectives` | List all objectives |
| `get_objective` | Get objective details |
| `create_objective` | Create an objective |
| `update_objective` | Update objective |
| `delete_objective` | Delete an objective |

### Key Results (OKRs)
| MCP Tool | Description |
|----------|-------------|
| `list_key_results` | List key results |
| `get_key_result` | Get key result details |
| `create_key_result` | Create a key result |
| `update_key_result` | Update key result |
| `delete_key_result` | Delete a key result |

### Launches
| MCP Tool | Description |
|----------|-------------|
| `list_launches` | List all launches |
| `get_launch` | Get launch details |
| `create_launch` | Create a launch |
| `update_launch` | Update launch |
| `delete_launch` | Delete a launch |

### Checklist Sections
| MCP Tool | Description |
|----------|-------------|
| `list_checklist_sections` | List checklist sections |
| `get_checklist_section` | Get section details |
| `create_checklist_section` | Create a section |
| `update_checklist_section` | Update section |
| `delete_checklist_section` | Delete a section |

### Tasks
| MCP Tool | Description |
|----------|-------------|
| `list_launch_tasks` | List tasks in a launch |
| `get_task` | Get task details |
| `create_task` | Create a task |
| `update_task` | Update task |
| `delete_task` | Delete a task |

### Organization
| MCP Tool | Description |
|----------|-------------|
| `list_users` | List all users |
| `list_teams` | List all teams |
| `check_status` | Check API status |

## API Coverage

| Feature | Read | Create | Update | Delete |
|---------|------|--------|--------|--------|
| Roadmaps | ✅ | - | - | - |
| Lanes | ✅ | ✅ | ✅ | ✅ |
| Milestones | ✅ | ✅ | ✅ | ✅ |
| Bars | ✅ | ✅ | ✅ | ✅ |
| Bar Comments | ✅ | - | - | - |
| Bar Connections | ✅ | - | - | - |
| Bar Links | ✅ | - | - | - |
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

## License

MIT License - see [LICENSE](LICENSE) for details.
