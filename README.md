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
| **Bar Comments** | Yes | Yes | - | - |
| **Bar Connections** | Yes | Yes | - | Yes |
| **Bar Links** | Yes | Yes | Yes | Yes |
| **Lanes** (categories) | Yes | Yes | Yes | Yes |
| **Milestones** | Yes | Yes | Yes | Yes |
| **Ideas** (Discovery) | Yes | Yes | Yes | - |
| **Idea Customers** | Yes | Yes | - | Yes |
| **Idea Tags** | Yes | Yes | - | Yes |
| **Opportunities** | Yes | Yes | Yes | Yes |
| **Idea Forms** | Yes | - | - | - |
| **Objectives** (OKRs) | Yes | Yes | Yes | Yes |
| **Key Results** | Yes | Yes | Yes | Yes |
| **Launches** | Yes | - | - | - |

---

## Troubleshooting

**"Command not found" or "spawn ENOENT"**

Your AI assistant can't find the binary. This means:
- **Mac/Linux**: The file isn't at `/usr/local/bin/productplan`, or you forgot to run `chmod +x`
- **Windows**: The path in your config doesn't match where you saved the `.exe`

Fix: Verify the binary exists at the path in your config. Run `ls -la /usr/local/bin/productplan` (Mac/Linux) or check if `C:\Tools\productplan.exe` exists (Windows).

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

v4.2 provides 24 READ tools and 12 WRITE tools (action-based):

**Read tools:**
- Roadmaps: `list_roadmaps`, `get_roadmap`, `get_roadmap_bars`, `get_roadmap_lanes`, `get_roadmap_milestones`
- Bars: `get_bar`, `get_bar_children`, `get_bar_comments`, `get_bar_connections`, `get_bar_links`
- OKRs: `list_objectives`, `get_objective`, `list_key_results`
- Discovery: `list_ideas`, `get_idea`, `get_idea_customers`, `get_idea_tags`, `list_opportunities`, `get_opportunity`, `list_idea_forms`, `get_idea_form`
- Launches: `list_launches`, `get_launch`
- Admin: `check_status`

**Write tools:**
- Roadmaps: `manage_bar`, `manage_lane`, `manage_milestone`
- Bar relationships: `manage_bar_comment`, `manage_bar_connection`, `manage_bar_link`
- OKRs: `manage_objective`, `manage_key_result`
- Discovery: `manage_idea`, `manage_idea_customer`, `manage_idea_tag`, `manage_opportunity`

Example:
```json
{"tool": "list_roadmaps", "arguments": {}}
{"tool": "manage_bar", "arguments": {"action": "create", "roadmap_id": "123", "lane_id": "456", "name": "New feature"}}
{"tool": "manage_idea", "arguments": {"action": "create", "name": "Mobile app improvements"}}
```

</details>

---

## Changelog

**v4.2.0** - Discovery module: full CRUD for ideas, opportunities, idea customers/tags; read-only idea forms (36 tools total)
**v4.1.0** - Bar relationships: children, comments, connections, links with full CRUD (26 tools)
**v4.0.0** - Redesigned tool architecture: 14 granular READ tools + 5 action-based WRITE tools; bars now include lane names
**v3.0.0** - Consolidated 58 tools into 15 (74% fewer tokens), added response summarization
**v2.0.0** - Initial public release with full ProductPlan API v2 coverage

## License

MIT License - see [LICENSE](LICENSE)
