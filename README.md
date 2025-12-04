# ProductPlan MCP Server

A Model Context Protocol (MCP) server that provides AI assistants with access to [ProductPlan](https://www.productplan.com/) roadmapping and strategy features.

## Two Versions Available

| Version | Location | Best For |
|---------|----------|----------|
| **Go** (recommended) | [`go-version/`](go-version/) | CLI + MCP server, single binary, no dependencies |
| **Node.js** | Root directory | Quick npm install |

**Go version** is recommended: single binary, works as both CLI and MCP server, no runtime dependencies.

### Supported Tools

| Tool | Support |
|------|---------|
| Claude Code (CLI) | ✅ Native |
| Cursor | ✅ Native |
| Claude Desktop | ✅ Native |
| VS Code + Cline | ✅ Via extension |
| VS Code + Continue | ✅ Via extension |
| VS Code + Roo Code | ✅ Via extension |
| Terminal (CLI) | ✅ Direct |

## Features

This MCP server exposes ProductPlan's API through standardized tools that any MCP-compatible AI assistant can use:

### Roadmaps
- `list_roadmaps` - List all roadmaps in your account
- `get_roadmap` - Get details of a specific roadmap
- `get_roadmap_bars` - Get all bars (items) from a roadmap
- `get_roadmap_lanes` - Get all lanes from a roadmap
- `get_roadmap_milestones` - Get all milestones from a roadmap

### Bars (Roadmap Items)
- `get_bar` - Get details of a specific bar
- `create_bar` - Create a new bar on a roadmap
- `update_bar` - Update an existing bar

### Discovery (Ideas & Opportunities)
- `list_ideas` - List all ideas
- `get_idea` - Get details of a specific idea
- `create_idea` - Create a new idea
- `list_opportunities` - List all opportunities

### Strategy (OKRs)
- `list_objectives` - List all strategic objectives
- `get_objective` - Get details of a specific objective
- `list_key_results` - List key results for an objective

### Launches
- `list_launches` - List all launches
- `get_launch` - Get details of a specific launch
- `list_launch_tasks` - List tasks for a launch

### Account
- `list_users` - List all users in the account
- `list_teams` - List all teams in the account
- `check_status` - Check ProductPlan API status

## Installation

### Prerequisites

- Node.js 18 or higher
- A ProductPlan account with API access
- A ProductPlan API token

### Getting Your API Token

1. Log in to ProductPlan
2. Go to **Settings** → **API** (or visit `https://app.productplan.com/settings/api`)
3. Generate a new API token
4. Copy the token for use in configuration

### Install from npm

```bash
npm install -g productplan-mcp-server
```

### Install from source

```bash
git clone https://github.com/YOUR_USERNAME/productplan-mcp-server.git
cd productplan-mcp-server
npm install
npm link  # Makes the command available globally
```

## Configuration

### For Claude Code / Claude Desktop

Add this to your Claude configuration file:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "productplan": {
      "command": "npx",
      "args": ["-y", "productplan-mcp-server"],
      "env": {
        "PRODUCTPLAN_API_TOKEN": "your-api-token-here"
      }
    }
  }
}
```

Or if installed globally:

```json
{
  "mcpServers": {
    "productplan": {
      "command": "productplan-mcp-server",
      "env": {
        "PRODUCTPLAN_API_TOKEN": "your-api-token-here"
      }
    }
  }
}
```

### For Other MCP Clients

Set the `PRODUCTPLAN_API_TOKEN` environment variable and run:

```bash
PRODUCTPLAN_API_TOKEN=your-token productplan-mcp-server
```

## Usage Examples

Once configured, you can ask your AI assistant questions like:

- "Show me all my roadmaps"
- "What objectives do we have in ProductPlan?"
- "Create a new idea titled 'Mobile app redesign'"
- "What are the key results for objective X?"
- "Add a bar to the Q1 roadmap for the authentication feature"

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
| Users | ✅ | ❌ | ❌ | ❌ |
| Teams | ✅ | ❌ | ❌ | ❌ |

> Note: Write operations are limited by ProductPlan's API. Contact ProductPlan for expanded API access.

## Development

```bash
# Clone the repository
git clone https://github.com/YOUR_USERNAME/productplan-mcp-server.git
cd productplan-mcp-server

# Install dependencies
npm install

# Run in development mode
PRODUCTPLAN_API_TOKEN=your-token npm start
```

## Troubleshooting

### "PRODUCTPLAN_API_TOKEN environment variable is required"

Make sure you've set the API token in your MCP configuration or environment.

### "API error 401: Unauthorized"

Your API token is invalid or expired. Generate a new one from ProductPlan settings.

### "API error 403: Forbidden"

Your account may not have API access enabled. Contact ProductPlan support.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Built with the [Model Context Protocol SDK](https://github.com/modelcontextprotocol/sdk)
- Powered by the [ProductPlan API](https://help.productplan.com/en/collections/1803015-productplan-api)
