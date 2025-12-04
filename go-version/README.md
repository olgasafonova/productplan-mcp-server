# ProductPlan CLI & MCP Server (Go)

A single binary that works as both a CLI tool and an MCP server for ProductPlan.

## Features

- **CLI mode**: Query ProductPlan directly from terminal
- **MCP server mode**: Integrate with Claude Code, Cursor, Claude Desktop, etc.
- **No dependencies**: Single binary, no Node.js/Python required
- **Cross-platform**: macOS, Linux, Windows

## Installation

### Download Binary

Download the appropriate binary for your platform from [Releases](https://github.com/olgasafonova/productplan-mcp-server/releases):

| Platform | File |
|----------|------|
| macOS (Intel) | `productplan-darwin-amd64` |
| macOS (Apple Silicon) | `productplan-darwin-arm64` |
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
cd productplan-mcp-server/go-version
go build -o productplan .
```

## Configuration

Set your ProductPlan API token as an environment variable:

```bash
export PRODUCTPLAN_API_TOKEN="your-api-token"
```

Get your token from [ProductPlan Settings > API](https://app.productplan.com/settings/api).

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

# List ideas
productplan ideas

# List launches
productplan launches

# List users and teams
productplan users
productplan teams
```

## MCP Server Usage

### Claude Code

Add to your Claude Code settings:

```json
{
  "mcpServers": {
    "productplan": {
      "command": "/usr/local/bin/productplan",
      "args": ["serve"],
      "env": {
        "PRODUCTPLAN_API_TOKEN": "your-token"
      }
    }
  }
}
```

Or without `serve` (default mode):

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

### Cursor IDE

Add to Cursor's MCP configuration (Settings > MCP Servers):

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

## Available Commands / MCP Tools

| Command | MCP Tool | Description |
|---------|----------|-------------|
| `roadmaps` | `list_roadmaps` | List all roadmaps |
| `roadmaps <id>` | `get_roadmap` | Get roadmap details |
| `bars <roadmap_id>` | `get_roadmap_bars` | Get roadmap bars |
| `objectives` | `list_objectives` | List all OKRs |
| `objectives <id>` | `get_objective` | Get objective details |
| `key-results <id>` | `list_key_results` | Get key results |
| `ideas` | `list_ideas` | List all ideas |
| `launches` | `list_launches` | List all launches |
| `users` | `list_users` | List users |
| `teams` | `list_teams` | List teams |
| `status` | `check_status` | Check API status |

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

MIT
