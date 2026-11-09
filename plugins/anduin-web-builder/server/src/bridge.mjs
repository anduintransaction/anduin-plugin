// The local end of the Anduin landing page builder bridge: builder tabs connect here over WebSocket, the agent pairs
// with one of them by the code the user hands over, and the builder's WebMCP tools are relayed to that tab.
import { WebSocketServer } from "ws";

export const PROTOCOL_VERSION = 1;
export const PORTS = [47811, 47812, 47813, 47814, 47815];

const HOST = "127.0.0.1";
const MAX_PENDING = 20;
const MAX_PAYLOAD = 8 * 1024 * 1024;
const HELLO_TIMEOUT_MS = 5000;
const CODE = /^[ABCDEFGHJKMNPQRSTUVWXYZ2-9]{6}$/;

export const Text = {
  notPaired:
    "No builder connected. If you were given a builder code earlier in this conversation, call connect_builder with " +
    'it again. Otherwise ask the user to click "Connect your agent" in the Anduin landing page builder and give you ' +
    "the code it shows, then call connect_builder.",
  wrongCode: (code) =>
    `No builder is waiting with code ${code}. Ask the user for the code the builder shows; it changes each ` +
    'time they click "Connect your agent". If the builder still shows that code, wait a few seconds and ' +
    "try again: the builder keeps looking for agents.",
  connected: (title) =>
    `Connected to the builder${title ? ` (page titled ${JSON.stringify(title)})` : ""}. Call list_builder_tools, ` +
    "then call_builder_tool with get_guide and follow it.",
  disconnected:
    "The builder disconnected before replying. The change may or may not have applied. Call connect_builder again " +
    "with the code you were given (wait a few seconds and retry if it isn't found); if it still isn't found, ask " +
    'the user to click "Connect your agent" and give you the new code. Then call read_page.',
  cancelled: "Cancelled. The builder may still finish the change: call read_page before you retry.",
  portsBusy:
    `Ports ${PORTS[0]}-${PORTS[PORTS.length - 1]} are all in use, so the builder cannot reach this agent. Ask the ` +
    "user to close another agent session that has the Anduin Web Builder plugin, then restart this one.",
  // Shown to the user in the builder's connect dialog.
  tooManyTabs: "Too many builder tabs are waiting for an agent. Close some; this one connects once there's room.",
  codeTaken: "That code is already in use. Cancel or exit AI mode, then click Connect your agent again.",
  pluginTooOld: "Your agent's Anduin Web Builder plugin is older than this builder. Update the plugin and restart your agent.",
  builderTooOld: "This builder page is older than your agent's plugin. Reload the page."
};

export async function startBridge({
  ports = PORTS,
  host = HOST,
  maxPending = MAX_PENDING,
  maxPayload = MAX_PAYLOAD,
  helloTimeoutMs = HELLO_TIMEOUT_MS
} = {}) {
  const wss = await listenOnFirstFree(ports, host, maxPayload);
  const pending = new Map(); // code -> { ws, title }
  const calls = new Map(); // id -> { ws, done }
  const helloed = new WeakSet(); // sockets that have sent their one hello
  let paired = null; // { ws, title, code, origin }
  // The last pairing outlives its socket, so the same tab coming back after a drop re-pairs without the agent.
  let last = null; // { code, origin }
  let nextId = 1;

  const send = (ws, message) => {
    if (ws.readyState === ws.OPEN) ws.send(JSON.stringify(message));
  };
  const refuse = (ws, message) => {
    send(ws, { type: "error", message });
    ws.close(4001, "refused");
  };

  const origins = new WeakMap(); // socket -> its page's Origin header, which the browser sets and a page can't fake

  wss?.on("connection", (ws, req) => {
    origins.set(ws, String(req.headers.origin ?? ""));
    // A bad frame (too big, bad UTF-8) makes ws emit 'error', which would crash the process without a listener.
    // 'close' still follows, and gone() answers the socket's in-flight calls.
    ws.on("error", () => {});
    // A socket that never says hello would hold a connection forever, so it is closed after helloTimeoutMs.
    const helloTimer = setTimeout(() => ws.close(4002, "no hello"), helloTimeoutMs);
    ws.on("message", (data) => {
      let message;
      try {
        message = JSON.parse(data.toString());
      } catch {
        ws.close(4002, "not json");
        return;
      }
      if (message?.type === "hello") {
        clearTimeout(helloTimer);
        hello(ws, message);
      } else if (message?.type === "knows") {
        // Lets the builder ask, before connecting, whether this agent would re-pair its code. It pairs nothing, and
        // answers yes only for the last pairing's code from the same site, so it tells a page elsewhere nothing.
        send(ws, { type: "knows", known: last?.code === message.code && last.origin === origins.get(ws) });
      } else if (message?.type === "tools" || message?.type === "result") answer(ws, message);
    });
    ws.on("close", () => {
      clearTimeout(helloTimer);
      gone(ws);
    });
  });

  function hello(ws, { v, code, title }) {
    // One hello per socket: a second one, including from the paired tab, is ignored.
    if (helloed.has(ws)) return;
    helloed.add(ws);
    if (v !== PROTOCOL_VERSION) {
      refuse(ws, typeof v === "number" && v > PROTOCOL_VERSION ? Text.pluginTooOld : Text.builderTooOld);
      return;
    }
    if (typeof code !== "string" || !CODE.test(code)) {
      ws.close(4002, "bad code");
      return;
    }
    // The tab this agent was paired with, back after a drop: same code, same site. The site check stops a page
    // elsewhere from guessing its way into a pairing.
    if (last?.code === code && last.origin === origins.get(ws)) {
      const previous = paired;
      paired = { ws, title: typeof title === "string" ? title.slice(0, 120) : "", code, origin: last.origin };
      if (previous && previous.ws !== ws) previous.ws.close(4000, "replaced");
      send(ws, { type: "paired" });
      return;
    }
    if (pending.size >= maxPending) {
      refuse(ws, Text.tooManyTabs);
      return;
    }
    if (pending.has(code)) {
      refuse(ws, Text.codeTaken);
      return;
    }
    pending.set(code, { ws, title: typeof title === "string" ? title.slice(0, 120) : "", origin: origins.get(ws) });
  }

  function answer(ws, message) {
    const call = calls.get(message.id);
    if (!call || call.ws !== ws) return;
    calls.delete(message.id);
    call.done(message.type === "tools" ? { tools: message.tools ?? [] } : { text: String(message.text ?? "") });
  }

  function gone(ws) {
    for (const [code, entry] of pending) if (entry.ws === ws) pending.delete(code);
    if (paired?.ws === ws) paired = null;
    for (const [id, call] of calls) {
      if (call.ws === ws) {
        calls.delete(id);
        call.done({ text: Text.disconnected, isError: true });
      }
    }
  }

  function connect(rawCode) {
    if (!wss) return { text: Text.portsBusy, isError: true };
    const code = String(rawCode ?? "").trim().toUpperCase();
    if (paired?.code === code) return { text: Text.connected(paired.title) };
    const entry = pending.get(code);
    if (!entry) return { text: Text.wrongCode(code), isError: true };
    pending.delete(code);
    // One tab at a time: the tab paired before is closed and goes back to waiting with its own code.
    if (paired) paired.ws.close(4000, "replaced");
    paired = { ...entry, code };
    last = { code, origin: entry.origin };
    send(entry.ws, { type: "paired" });
    return { text: Text.connected(entry.title) };
  }

  // No timeout: a theme change waits for the user's answer in the builder. A call ends when the tab answers, the
  // tab disconnects, or the agent cancels.
  function request(message, signal) {
    if (!paired) return Promise.resolve({ text: Text.notPaired, isError: true });
    const { ws } = paired;
    const id = nextId++;
    return new Promise((resolve) => {
      if (signal?.aborted) {
        resolve({ text: Text.cancelled, isError: true });
        return;
      }
      const onAbort = () => {
        if (!calls.delete(id)) return;
        send(ws, { type: "cancel", id });
        resolve({ text: Text.cancelled, isError: true });
      };
      calls.set(id, {
        ws,
        done: (result) => {
          signal?.removeEventListener("abort", onAbort);
          resolve(result);
        }
      });
      signal?.addEventListener("abort", onAbort, { once: true });
      send(ws, { ...message, id });
    });
  }

  async function listTools(signal) {
    const result = await request({ type: "list" }, signal);
    return result.tools ? { text: JSON.stringify(result.tools, null, 2) } : result;
  }

  const callTool = (name, input, signal) => request({ type: "call", tool: String(name), input: input ?? {} }, signal);

  const close = () =>
    new Promise((resolve) => {
      if (!wss) {
        resolve();
        return;
      }
      for (const ws of wss.clients) ws.terminate();
      wss.close(() => resolve());
    });

  return {
    port: wss ? wss.address().port : null,
    address: () => wss?.address() ?? null,
    connect,
    listTools,
    callTool,
    close
  };
}

async function listenOnFirstFree(ports, host, maxPayload) {
  for (const port of ports) {
    const wss = await new Promise((resolve) => {
      const server = new WebSocketServer({ host, port, maxPayload });
      server.once("listening", () => resolve(server));
      server.once("error", () => resolve(null));
    });
    if (wss) return wss;
  }
  return null;
}
