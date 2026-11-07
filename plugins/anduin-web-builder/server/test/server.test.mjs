import { test } from "node:test";
import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { createInterface } from "node:readline";
import { fileURLToPath } from "node:url";

// The committed bundle, spawned the way the agent's host does. The bridge takes the first free port of 47811-47815,
// so a real server that is already running only moves this one to the next port.
const BUNDLE = fileURLToPath(new URL("../dist/server.mjs", import.meta.url));

test("the bundled server lists its three tools and refuses a call before a builder is connected", { timeout: 20000 }, async () => {
  const child = spawn(process.execPath, [BUNDLE], { stdio: ["pipe", "pipe", "ignore"] });
  const waiting = new Map(); // JSON-RPC id -> resolve
  createInterface({ input: child.stdout }).on("line", (line) => {
    const reply = JSON.parse(line);
    waiting.get(reply.id)?.(reply);
  });
  const write = (message) => child.stdin.write(`${JSON.stringify({ jsonrpc: "2.0", ...message })}\n`);
  const request = (id, method, params) => {
    const reply = new Promise((resolve) => waiting.set(id, resolve));
    write({ id, method, params });
    return reply.then((message) => {
      assert.equal(message.error, undefined, JSON.stringify(message.error));
      return message.result;
    });
  };
  try {
    await request(1, "initialize", {
      protocolVersion: "2025-03-26",
      capabilities: {},
      clientInfo: { name: "server-test", version: "0" }
    });
    write({ method: "notifications/initialized" });
    const { tools } = await request(2, "tools/list", {});
    assert.deepEqual(tools.map((tool) => tool.name).sort(), ["call_builder_tool", "connect_builder", "list_builder_tools"]);
    const call = await request(3, "tools/call", { name: "call_builder_tool", arguments: { name: "read_page" } });
    assert.equal(call.isError, true);
    assert.match(call.content[0].text, /^No builder connected/);
  } finally {
    child.kill();
  }
});
