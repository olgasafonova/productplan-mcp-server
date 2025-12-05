# ProductPlan MCP Server

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

**On Windows**, move the `.exe` file to a folder in your PATH, or note down where you saved it.

### Step 3: Connect to your AI assistant

Pick the tool you use:

<details>
<summary><strong>Claude Desktop</strong> (click to expand)</summary>

1. Find your config file:
   - **Mac**: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

2. Open it in any text editor and add this (replace `your-token` with your actual API token):

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

3. Restart Claude Desktop

</details>

<details>
<summary><strong>Claude Code (Terminal)</strong></summary>

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

</details>

<details>
<summary><strong>Cursor</strong></summary>

1. Open Cursor
2. Go to **Settings** → **MCP Servers**
3. Add this configuration:

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

</details>

<details>
<summary><strong>VS Code + Cline</strong></summary>

1. Install the [Cline extension](https://marketplace.visualstudio.com/items?itemName=saoudrizwan.claude-dev)
2. Open VS Code settings (JSON) and add:

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

</details>

<details>
<summary><strong>VS Code + Continue</strong></summary>

1. Install the [Continue extension](https://marketplace.visualstudio.com/items?itemName=continue.continue)
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

</details>

<details>
<summary><strong>n8n (Workflow Automation)</strong></summary>

1. Set environment variable on your n8n instance:
   ```
   N8N_COMMUNITY_PACKAGES_ALLOW_TOOL_USAGE=true
   ```
2. Add an **MCP Client** node to your workflow
3. Configure:
   - **Command**: `/usr/local/bin/productplan`
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
| **Bars** (roadmap items) | Yes | Yes | Yes | Yes |
| **Lanes** (categories) | Yes | Yes | Yes | Yes |
| **Milestones** | Yes | Yes | Yes | Yes |
| **Ideas** (Discovery) | Yes | Yes | Yes | - |
| **Opportunities** | Yes | Yes | Yes | - |
| **Objectives** (OKRs) | Yes | Yes | Yes | Yes |
| **Key Results** | Yes | Yes | Yes | Yes |
| **Launches** | Yes | Yes | Yes | Yes |
| **Tasks** (launch checklists) | Yes | Yes | Yes | Yes |
| **Users & Teams** | Yes | - | - | - |

---

## Troubleshooting

**"Command not found"**
Make sure you ran the `chmod` and `mv` commands from Step 2. On Windows, ensure the .exe is in your PATH.

**"Invalid API token"**
Double-check your token at [ProductPlan Settings → API](https://app.productplan.com/settings/api). Tokens can expire or be regenerated.

**"No roadmaps found"**
Your API token only accesses data you have permission to see in ProductPlan. Check that your account has access to the roadmaps you're looking for.

**AI assistant doesn't see ProductPlan**
Restart your AI assistant after editing the config file. The MCP server only loads on startup.

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
<summary>Build from source</summary>

```bash
git clone https://github.com/olgasafonova/productplan-mcp-server.git
cd productplan-mcp-server
go build -o productplan .
```

Build for all platforms:
```bash
make build-all
make release
```

</details>

<details>
<summary>MCP tool reference</summary>

v3.0 consolidates 58 API operations into 15 action-based tools:

| Tool | Actions |
|------|---------|
| `roadmaps` | list, get, get_bars, get_comments |
| `lanes` | list, create, update, delete |
| `milestones` | list, create, update, delete |
| `bars` | get, create, update, delete, get_children, get_comments, get_connections, get_links |
| `bar_connections` | list, create, delete |
| `bar_links` | list, create, delete |
| `ideas` | list, get, create, update, get_customers, get_tags |
| `opportunities` | list, get, create, update |
| `idea_forms` | list, get |
| `objectives` | list, get, create, update, delete |
| `key_results` | list, get, create, update, delete |
| `launches` | list, get, create, update, delete |
| `checklist_sections` | list, get, create, update, delete |
| `tasks` | list, get, create, update, delete |
| `organization` | users, teams, status |

Example tool call:
```json
{"tool": "roadmaps", "arguments": {"action": "list"}}
```

</details>

---

## Changelog

**v3.0.0** - Consolidated 58 tools into 15 (74% fewer tokens), added response summarization
**v2.0.0** - Initial public release with full ProductPlan API v2 coverage

## License

MIT License - see [LICENSE](LICENSE)
