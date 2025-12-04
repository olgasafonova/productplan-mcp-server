# ProductPlan CLI & MCP Server

A single binary that provides both CLI access and MCP server integration for [ProductPlan](https://www.productplan.com/).

- **CLI mode**: Query ProductPlan directly from terminal
- **MCP server mode**: Integrate with Claude Code, Cursor, VS Code, etc.
- **No dependencies**: Single ~5MB binary
- **Cross-platform**: macOS, Linux, Windows

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

# List roadmaps
productplan roadmaps

# Get roadmap details
productplan roadmaps 12345

# List bars in a roadmap
productplan bars 12345

# List all objectives (OKRs)
productplan objectives

# Get objective details
productplan objectives 67890

# List key results
productplan key-results 67890

# List ideas, launches, users, teams
productplan ideas
productplan launches
productplan users
productplan teams
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

## Available Commands / MCP Tools

| CLI Command | MCP Tool | Description |
|-------------|----------|-------------|
| `roadmaps` | `list_roadmaps` | List all roadmaps |
| `roadmaps <id>` | `get_roadmap` | Get roadmap details |
| `bars <roadmap_id>` | `get_roadmap_bars` | Get roadmap bars |
| - | `get_roadmap_lanes` | Get roadmap lanes |
| - | `get_roadmap_milestones` | Get roadmap milestones |
| - | `get_bar` | Get bar details |
| - | `create_bar` | Create a bar |
| - | `update_bar` | Update a bar |
| `objectives` | `list_objectives` | List all OKRs |
| `objectives <id>` | `get_objective` | Get objective details |
| `key-results <id>` | `list_key_results` | Get key results |
| `ideas` | `list_ideas` | List ideas |
| - | `get_idea` | Get idea details |
| - | `create_idea` | Create an idea |
| - | `list_opportunities` | List opportunities |
| `launches` | `list_launches` | List launches |
| - | `get_launch` | Get launch details |
| - | `list_launch_tasks` | Get launch tasks |
| `users` | `list_users` | List users |
| `teams` | `list_teams` | List teams |
| `status` | `check_status` | Check API status |

## API Coverage

| Feature | Read | Create | Update | Delete |
|---------|------|--------|--------|--------|
| Roadmaps | ✅ | ❌ | ❌ | ❌ |
| Bars | ✅ | ✅ | ✅ | ❌ |
| Lanes | ✅ | ❌ | ❌ | ❌ |
| Milestones | ✅ | ❌ | ❌ | ❌ |
| Ideas | ✅ | ✅ | ❌ | ❌ |
| Opportunities | ✅ | ❌ | ❌ | ❌ |
| Objectives | ✅ | ❌ | ❌ | ❌ |
| Key Results | ✅ | ❌ | ❌ | ❌ |
| Launches | ✅ | ❌ | ❌ | ❌ |

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
