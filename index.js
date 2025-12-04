#!/usr/bin/env node

const { Server } = require("@modelcontextprotocol/sdk/server/index.js");
const { StdioServerTransport } = require("@modelcontextprotocol/sdk/server/stdio.js");
const {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} = require("@modelcontextprotocol/sdk/types.js");

const API_BASE = "https://app.productplan.com/api/v2";
const API_TOKEN = process.env.PRODUCTPLAN_API_TOKEN;

async function apiRequest(method, endpoint, body = null) {
  const url = `${API_BASE}${endpoint}`;
  const options = {
    method,
    headers: {
      "Authorization": `Bearer ${API_TOKEN}`,
      "Content-Type": "application/json",
    },
  };

  if (body) {
    options.body = JSON.stringify(body);
  }

  const response = await fetch(url, options);

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`API error ${response.status}: ${errorText}`);
  }

  return response.json();
}

const server = new Server(
  {
    name: "productplan-mcp-server",
    version: "1.0.0",
  },
  {
    capabilities: {
      tools: {},
    },
  }
);

// Define available tools
server.setRequestHandler(ListToolsRequestSchema, async () => {
  return {
    tools: [
      // Roadmaps
      {
        name: "list_roadmaps",
        description: "List all roadmaps in your ProductPlan account",
        inputSchema: {
          type: "object",
          properties: {},
        },
      },
      {
        name: "get_roadmap",
        description: "Get details of a specific roadmap",
        inputSchema: {
          type: "object",
          properties: {
            id: { type: "string", description: "Roadmap ID" },
          },
          required: ["id"],
        },
      },
      {
        name: "get_roadmap_bars",
        description: "Get all bars (items) from a roadmap",
        inputSchema: {
          type: "object",
          properties: {
            roadmap_id: { type: "string", description: "Roadmap ID" },
          },
          required: ["roadmap_id"],
        },
      },
      {
        name: "get_roadmap_lanes",
        description: "Get all lanes from a roadmap",
        inputSchema: {
          type: "object",
          properties: {
            roadmap_id: { type: "string", description: "Roadmap ID" },
          },
          required: ["roadmap_id"],
        },
      },
      {
        name: "get_roadmap_milestones",
        description: "Get all milestones from a roadmap",
        inputSchema: {
          type: "object",
          properties: {
            roadmap_id: { type: "string", description: "Roadmap ID" },
          },
          required: ["roadmap_id"],
        },
      },
      // Bars
      {
        name: "get_bar",
        description: "Get details of a specific bar",
        inputSchema: {
          type: "object",
          properties: {
            id: { type: "string", description: "Bar ID" },
          },
          required: ["id"],
        },
      },
      {
        name: "create_bar",
        description: "Create a new bar on a roadmap",
        inputSchema: {
          type: "object",
          properties: {
            roadmap_id: { type: "string", description: "Roadmap ID" },
            lane_id: { type: "string", description: "Lane ID" },
            name: { type: "string", description: "Bar name" },
            start_date: { type: "string", description: "Start date (YYYY-MM-DD)" },
            end_date: { type: "string", description: "End date (YYYY-MM-DD)" },
            description: { type: "string", description: "Bar description" },
          },
          required: ["roadmap_id", "lane_id", "name"],
        },
      },
      {
        name: "update_bar",
        description: "Update an existing bar",
        inputSchema: {
          type: "object",
          properties: {
            id: { type: "string", description: "Bar ID" },
            name: { type: "string", description: "Bar name" },
            start_date: { type: "string", description: "Start date (YYYY-MM-DD)" },
            end_date: { type: "string", description: "End date (YYYY-MM-DD)" },
            description: { type: "string", description: "Bar description" },
          },
          required: ["id"],
        },
      },
      // Discovery
      {
        name: "list_ideas",
        description: "List all ideas in Discovery",
        inputSchema: {
          type: "object",
          properties: {},
        },
      },
      {
        name: "get_idea",
        description: "Get details of a specific idea",
        inputSchema: {
          type: "object",
          properties: {
            id: { type: "string", description: "Idea ID" },
          },
          required: ["id"],
        },
      },
      {
        name: "create_idea",
        description: "Create a new idea",
        inputSchema: {
          type: "object",
          properties: {
            title: { type: "string", description: "Idea title" },
            description: { type: "string", description: "Idea description" },
          },
          required: ["title"],
        },
      },
      {
        name: "list_opportunities",
        description: "List all opportunities in Discovery",
        inputSchema: {
          type: "object",
          properties: {},
        },
      },
      // Strategy
      {
        name: "list_objectives",
        description: "List all strategic objectives",
        inputSchema: {
          type: "object",
          properties: {},
        },
      },
      {
        name: "get_objective",
        description: "Get details of a specific objective",
        inputSchema: {
          type: "object",
          properties: {
            id: { type: "string", description: "Objective ID" },
          },
          required: ["id"],
        },
      },
      {
        name: "list_key_results",
        description: "List key results for an objective",
        inputSchema: {
          type: "object",
          properties: {
            objective_id: { type: "string", description: "Objective ID" },
          },
          required: ["objective_id"],
        },
      },
      // Launches
      {
        name: "list_launches",
        description: "List all launches",
        inputSchema: {
          type: "object",
          properties: {},
        },
      },
      {
        name: "get_launch",
        description: "Get details of a specific launch",
        inputSchema: {
          type: "object",
          properties: {
            id: { type: "string", description: "Launch ID" },
          },
          required: ["id"],
        },
      },
      {
        name: "list_launch_tasks",
        description: "List tasks for a launch",
        inputSchema: {
          type: "object",
          properties: {
            launch_id: { type: "string", description: "Launch ID" },
          },
          required: ["launch_id"],
        },
      },
      // Utility
      {
        name: "list_users",
        description: "List all users in the account",
        inputSchema: {
          type: "object",
          properties: {},
        },
      },
      {
        name: "list_teams",
        description: "List all teams in the account",
        inputSchema: {
          type: "object",
          properties: {},
        },
      },
      {
        name: "check_status",
        description: "Check ProductPlan API status",
        inputSchema: {
          type: "object",
          properties: {},
        },
      },
    ],
  };
});

// Handle tool calls
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  try {
    let result;

    switch (name) {
      // Roadmaps
      case "list_roadmaps":
        result = await apiRequest("GET", "/roadmaps");
        break;
      case "get_roadmap":
        result = await apiRequest("GET", `/roadmaps/${args.id}`);
        break;
      case "get_roadmap_bars":
        result = await apiRequest("GET", `/roadmaps/${args.roadmap_id}/bars`);
        break;
      case "get_roadmap_lanes":
        result = await apiRequest("GET", `/roadmaps/${args.roadmap_id}/lanes`);
        break;
      case "get_roadmap_milestones":
        result = await apiRequest("GET", `/roadmaps/${args.roadmap_id}/milestones`);
        break;

      // Bars
      case "get_bar":
        result = await apiRequest("GET", `/bars/${args.id}`);
        break;
      case "create_bar":
        result = await apiRequest("POST", "/bars", {
          roadmap_id: args.roadmap_id,
          lane_id: args.lane_id,
          name: args.name,
          start_date: args.start_date,
          end_date: args.end_date,
          description: args.description,
        });
        break;
      case "update_bar":
        const updateData = {};
        if (args.name) updateData.name = args.name;
        if (args.start_date) updateData.start_date = args.start_date;
        if (args.end_date) updateData.end_date = args.end_date;
        if (args.description) updateData.description = args.description;
        result = await apiRequest("PATCH", `/bars/${args.id}`, updateData);
        break;

      // Discovery
      case "list_ideas":
        result = await apiRequest("GET", "/discovery/ideas");
        break;
      case "get_idea":
        result = await apiRequest("GET", `/discovery/ideas/${args.id}`);
        break;
      case "create_idea":
        result = await apiRequest("POST", "/discovery/ideas", {
          title: args.title,
          description: args.description,
        });
        break;
      case "list_opportunities":
        result = await apiRequest("GET", "/discovery/opportunities");
        break;

      // Strategy
      case "list_objectives":
        result = await apiRequest("GET", "/strategy/objectives");
        break;
      case "get_objective":
        result = await apiRequest("GET", `/strategy/objectives/${args.id}`);
        break;
      case "list_key_results":
        result = await apiRequest("GET", `/strategy/objectives/${args.objective_id}/key-results`);
        break;

      // Launches
      case "list_launches":
        result = await apiRequest("GET", "/launches");
        break;
      case "get_launch":
        result = await apiRequest("GET", `/launches/${args.id}`);
        break;
      case "list_launch_tasks":
        result = await apiRequest("GET", `/launches/${args.launch_id}/tasks`);
        break;

      // Utility
      case "list_users":
        result = await apiRequest("GET", "/users");
        break;
      case "list_teams":
        result = await apiRequest("GET", "/teams");
        break;
      case "check_status":
        result = await apiRequest("GET", "/status");
        break;

      default:
        throw new Error(`Unknown tool: ${name}`);
    }

    return {
      content: [
        {
          type: "text",
          text: JSON.stringify(result, null, 2),
        },
      ],
    };
  } catch (error) {
    return {
      content: [
        {
          type: "text",
          text: `Error: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
});

async function main() {
  if (!API_TOKEN) {
    console.error("Error: PRODUCTPLAN_API_TOKEN environment variable is required");
    process.exit(1);
  }

  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("ProductPlan MCP Server running on stdio");
}

main().catch(console.error);
