// The Anduin Web Builder plugin's MCP server: the agent's three tools, backed by the local builder bridge.
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import { startBridge } from "./bridge.mjs";

const VERSION = "0.2.0"; // keep in step with .claude-plugin/plugin.json

const bridge = await startBridge();
console.error(`[anduin-builder] ${bridge.port ? `listening on 127.0.0.1:${bridge.port}` : "no free port"}`);

const server = new McpServer({ name: "anduin-builder", version: VERSION });
const reply = ({ text, isError }) => ({ content: [{ type: "text", text }], ...(isError ? { isError: true } : {}) });
// The SDK passes `extra` (with the request's AbortSignal) last, with or without parsed arguments before it.
const signalOf = (args) => args[args.length - 1]?.signal;

server.registerTool(
  "connect_builder",
  {
    description:
      'Connects to the Anduin landing page builder tab the user opened for AI. The user clicks "Open for AI to ' +
      'connect" in the builder and gives you the 6-character code shown in its top bar. Call this first.',
    inputSchema: { code: z.string().describe("The 6-character code from the builder, e.g. K7Q2PX") }
  },
  async ({ code }) => reply(bridge.connect(code))
);

server.registerTool(
  "list_builder_tools",
  {
    description: "Lists the connected builder's tools: name, description and input schema. Call after connect_builder.",
    inputSchema: {}
  },
  async (...args) => reply(await bridge.listTools(signalOf(args)))
);

server.registerTool(
  "call_builder_tool",
  {
    description:
      "Runs one of the builder's tools on the connected tab and returns its result. Start with get_guide and follow it.",
    inputSchema: {
      name: z.string().describe("The builder tool's name, from list_builder_tools"),
      input: z.record(z.string(), z.unknown()).optional().describe("The tool's input, as an object")
    }
  },
  async (...args) => {
    const [{ name, input }] = args;
    return reply(await bridge.callTool(name, input ?? {}, signalOf(args)));
  }
);

const shutdown = async () => {
  await bridge.close();
  process.exit(0);
};
process.stdin.on("end", shutdown);
process.on("SIGTERM", shutdown);
process.on("SIGINT", shutdown);

await server.connect(new StdioServerTransport());
