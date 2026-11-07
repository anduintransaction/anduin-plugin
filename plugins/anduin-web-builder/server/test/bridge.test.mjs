import { test } from "node:test";
import assert from "node:assert/strict";
import net from "node:net";
import { WebSocket } from "ws";
import { startBridge, PROTOCOL_VERSION, Text } from "../src/bridge.mjs";

// Not the real range, so a running plugin server does not interfere.
const PORTS = [48811, 48812];
const tick = () => new Promise((resolve) => setTimeout(resolve, 30));

/** A fake builder tab: connects, says hello, and queues what the server sends. */
async function tab(port, { code = "ABC234", v = PROTOCOL_VERSION, title = "Home page" } = {}) {
  const ws = new WebSocket(`ws://127.0.0.1:${port}`);
  const inbox = [];
  const waiters = [];
  ws.on("message", (data) => {
    const message = JSON.parse(data.toString());
    const waiter = waiters.shift();
    if (waiter) waiter(message);
    else inbox.push(message);
  });
  const closed = new Promise((resolve) => ws.once("close", resolve));
  await new Promise((resolve, reject) => {
    ws.once("open", resolve);
    ws.once("error", reject);
  });
  ws.send(JSON.stringify({ type: "hello", v, code, title }));
  await tick();
  return {
    ws,
    inbox,
    closed,
    next: () => (inbox.length ? Promise.resolve(inbox.shift()) : new Promise((resolve) => waiters.push(resolve))),
    send: (message) => ws.send(JSON.stringify(message))
  };
}

async function withBridge(options, body) {
  const bridge = await startBridge({ ports: PORTS, ...options });
  try {
    await body(bridge);
  } finally {
    await bridge.close();
  }
}

test("listens on 127.0.0.1 only, on the first free port of the range", async () => {
  const blocker = net.createServer();
  await new Promise((resolve) => blocker.listen(PORTS[0], "127.0.0.1", resolve));
  try {
    await withBridge({}, async (bridge) => {
      assert.equal(bridge.port, PORTS[1]);
      assert.equal(bridge.address().address, "127.0.0.1");
    });
  } finally {
    await new Promise((resolve) => blocker.close(resolve));
  }
});

test("every port taken: connect says so", async () => {
  const blockers = await Promise.all(
    PORTS.map(
      (port) =>
        new Promise((resolve) => {
          const server = net.createServer();
          server.listen(port, "127.0.0.1", () => resolve(server));
        })
    )
  );
  try {
    await withBridge({}, async (bridge) => {
      assert.equal(bridge.port, null);
      assert.deepEqual(bridge.connect("ABC234"), { text: Text.portsBusy, isError: true });
    });
  } finally {
    await Promise.all(blockers.map((server) => new Promise((resolve) => server.close(resolve))));
  }
});

test("pairs only with the code the user handed over, case-insensitively", async () => {
  await withBridge({}, async (bridge) => {
    const t = await tab(bridge.port, { code: "ABC234" });
    const wrong = bridge.connect("ZZZ999");
    assert.equal(wrong.isError, true);
    assert.match(wrong.text, /No builder is waiting with code ZZZ999/);
    const right = bridge.connect(" abc234 ");
    assert.equal(right.isError, undefined);
    assert.match(right.text, /Home page/);
    assert.deepEqual(await t.next(), { type: "paired" });
    // Connecting again with the same code is not an error.
    assert.equal(bridge.connect("ABC234").isError, undefined);
  });
});

test("a tab that is not paired never receives calls", async () => {
  await withBridge({}, async (bridge) => {
    const t = await tab(bridge.port);
    assert.deepEqual(await bridge.callTool("read_page", {}), { text: Text.notPaired, isError: true });
    await tick();
    assert.equal(t.inbox.length, 0);
  });
});

test("relays list and call to the paired tab and returns its answers", async () => {
  await withBridge({}, async (bridge) => {
    const t = await tab(bridge.port);
    bridge.connect("ABC234");
    await t.next(); // paired

    const listing = bridge.listTools();
    const list = await t.next();
    assert.equal(list.type, "list");
    t.send({ type: "tools", id: list.id, tools: [{ name: "read_page", description: "Reads", inputSchema: {} }] });
    assert.match((await listing).text, /"read_page"/);

    const calling = bridge.callTool("edit_page", { ops: [1] });
    const call = await t.next();
    assert.deepEqual({ type: call.type, tool: call.tool, input: call.input }, {
      type: "call",
      tool: "edit_page",
      input: { ops: [1] }
    });
    t.send({ type: "result", id: call.id, text: "Applied 1 op." });
    assert.deepEqual(await calling, { text: "Applied 1 op." });
  });
});

test("an agent cancellation reaches the tab as cancel", async () => {
  await withBridge({}, async (bridge) => {
    const t = await tab(bridge.port);
    bridge.connect("ABC234");
    await t.next();
    const controller = new AbortController();
    const calling = bridge.callTool("update_theme", {}, controller.signal);
    const call = await t.next();
    controller.abort();
    assert.deepEqual(await t.next(), { type: "cancel", id: call.id });
    assert.equal((await calling).isError, true);
  });
});

test("the tab going away mid-call answers the call as disconnected", async () => {
  await withBridge({}, async (bridge) => {
    const t = await tab(bridge.port);
    bridge.connect("ABC234");
    await t.next();
    const calling = bridge.callTool("edit_page", {});
    await t.next();
    t.ws.close();
    assert.deepEqual(await calling, { text: Text.disconnected, isError: true });
    assert.deepEqual(await bridge.callTool("read_page", {}), { text: Text.notPaired, isError: true });
  });
});

test("a new pairing replaces the old one: the old tab is closed", async () => {
  await withBridge({}, async (bridge) => {
    const a = await tab(bridge.port, { code: "ABC234" });
    const b = await tab(bridge.port, { code: "XYZ789" });
    bridge.connect("ABC234");
    await a.next();
    bridge.connect("XYZ789");
    assert.deepEqual(await b.next(), { type: "paired" });
    await a.closed;
  });
});

test("a tab on another protocol version is refused with a reason", async () => {
  await withBridge({}, async (bridge) => {
    const t = await tab(bridge.port, { v: PROTOCOL_VERSION + 1 });
    assert.deepEqual(await t.next(), { type: "error", message: Text.pluginTooOld });
    await t.closed;
  });
});

test("no more than maxPending tabs wait at once", async () => {
  await withBridge({ maxPending: 2 }, async (bridge) => {
    await tab(bridge.port, { code: "AAAAA2" });
    await tab(bridge.port, { code: "AAAAA3" });
    const third = await tab(bridge.port, { code: "AAAAA4" });
    assert.deepEqual(await third.next(), { type: "error", message: Text.tooManyTabs });
    await third.closed;
  });
});
