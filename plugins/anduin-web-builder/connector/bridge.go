// The local end of the Anduin landing page builder bridge: builder tabs connect here over WebSocket, the agent pairs
// with one of them by the code the user hands over, and the builder's WebMCP tools are relayed to that tab.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// ProtocolVersion is the "v" a builder tab says hello with. Bump it, with the builder, only for a change the other
// side cannot read.
const ProtocolVersion = 1

// Ports the builder looks for agents on; this bridge listens on the first free one.
var Ports = []int{47811, 47812, 47813, 47814, 47815}

const writeTimeout = 10 * time.Second

var codePattern = regexp.MustCompile(`^[ABCDEFGHJKMNPQRSTUVWXYZ2-9]{6}$`)

// Close codes the builder reads.
const (
	closeReplaced  websocket.StatusCode = 4000
	closeRefused   websocket.StatusCode = 4001
	closeMalformed websocket.StatusCode = 4002
)

const (
	textNotPaired = "No builder connected. If you were given a builder code earlier in this conversation, call " +
		"connect_builder with it again. Otherwise ask the user to click \"Connect your agent\" in the Anduin landing " +
		"page builder and give you the code it shows, then call connect_builder."
	textDisconnected = "The builder disconnected before replying. The change may or may not have applied. Call " +
		"connect_builder again with the code you were given (wait a few seconds and retry if it isn't found); if it " +
		"still isn't found, ask the user to click \"Connect your agent\" and give you the new code. Then call read_page."
	textCancelled = "Cancelled. The builder may still finish the change: call read_page before you retry."
	// Shown to the user in the builder's connect dialog.
	textTooManyTabs   = "Too many builder tabs are waiting for an agent. Close some; this one connects once there's room."
	textCodeTaken     = "That code is already in use. Cancel or exit AI mode, then click Connect your agent again."
	textPluginTooOld  = "Your agent's Anduin Web Builder plugin is older than this builder. Update the plugin and restart your agent."
	textBuilderTooOld = "This builder page is older than your agent's plugin. Reload the page."
)

var textPortsBusy = fmt.Sprintf("Ports %d-%d are all in use, so the builder cannot reach this agent. Ask the user to "+
	"close another agent session that has the Anduin Web Builder plugin, then restart this one.",
	Ports[0], Ports[len(Ports)-1])

func textWrongCode(code string) string {
	return fmt.Sprintf("No builder is waiting with code %s. Ask the user for the code the builder shows; it changes "+
		"each time they click \"Connect your agent\". If the builder still shows that code, wait a few seconds and "+
		"try again: the builder keeps looking for agents.", code)
}

func textConnected(title string) string {
	named := ""
	if title != "" {
		named = " (page titled " + quote(title) + ")"
	}
	return "Connected to the builder" + named + ". Call list_builder_tools, then call_builder_tool with get_guide " +
		"and follow it."
}

// Result is what a tool call returns to the agent.
type Result struct {
	Text    string
	IsError bool
}

// Options tune the bridge; zero values take the defaults.
type Options struct {
	Ports        []int
	Host         string
	MaxPending   int
	MaxPayload   int64
	HelloTimeout time.Duration
}

// Bridge pairs one builder tab at a time with the agent and relays its tools.
type Bridge struct {
	opts     Options
	listener net.Listener // nil when every port was taken
	server   *http.Server

	mu      sync.Mutex
	tabs    map[*tab]struct{}
	pending map[string]*tab // code -> a tab waiting for the agent
	paired  *pairing
	// The last pairing outlives its socket, so the same tab coming back after a drop re-pairs without the agent.
	last   *lastPairing
	calls  map[int64]*call
	nextID int64
}

type tab struct {
	ws *websocket.Conn
	// Its page's Origin header, which the browser sets and a page can't fake.
	origin  string
	title   string
	helloed bool // a socket says hello once
}

type pairing struct {
	tab  *tab
	code string
}

type lastPairing struct{ code, origin string }

type call struct {
	tab  *tab
	done chan reply
}

type reply struct {
	Result
	tools json.RawMessage // set when the tab answered a list
}

// Start listens on the first free port. With none free it still returns a bridge, whose connect says so.
func Start(opts Options) *Bridge {
	if opts.Ports == nil {
		opts.Ports = Ports
	}
	if opts.Host == "" {
		opts.Host = "127.0.0.1"
	}
	if opts.MaxPending == 0 {
		opts.MaxPending = 20
	}
	if opts.MaxPayload == 0 {
		opts.MaxPayload = 8 << 20
	}
	if opts.HelloTimeout == 0 {
		opts.HelloTimeout = 5 * time.Second
	}
	b := &Bridge{
		opts:    opts,
		tabs:    map[*tab]struct{}{},
		pending: map[string]*tab{},
		calls:   map[int64]*call{},
	}
	for _, port := range opts.Ports {
		if l, err := net.Listen("tcp", net.JoinHostPort(opts.Host, strconv.Itoa(port))); err == nil {
			b.listener = l
			break
		}
	}
	if b.listener != nil {
		b.server = &http.Server{Handler: http.HandlerFunc(b.accept), ReadHeaderTimeout: 10 * time.Second}
		go func() { _ = b.server.Serve(b.listener) }()
	}
	return b
}

// Port is the port listened on, or 0 when every port was taken.
func (b *Bridge) Port() int {
	if b.listener == nil {
		return 0
	}
	return b.listener.Addr().(*net.TCPAddr).Port
}

// Close stops listening and drops every tab.
func (b *Bridge) Close() {
	if b.server == nil {
		return
	}
	_ = b.server.Close()
	// Also directly: a Close before Serve has started would leave the port bound until Serve gets to it.
	_ = b.listener.Close()
	b.mu.Lock()
	tabs := make([]*tab, 0, len(b.tabs))
	for t := range b.tabs {
		tabs = append(tabs, t)
	}
	b.mu.Unlock()
	// At once for all: one mid close handshake can hold its CloseNow for seconds.
	for _, t := range tabs {
		go func() { _ = t.ws.CloseNow() }()
	}
}

func (b *Bridge) accept(w http.ResponseWriter, r *http.Request) {
	// Builder pages are on Anduin's sites, never on this host, so the browser's cross-origin check would refuse them
	// all. The Origin is kept instead: re-pairing and knows answer only the site that paired.
	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	ws.SetReadLimit(b.opts.MaxPayload)
	t := &tab{ws: ws, origin: r.Header.Get("Origin")}
	b.mu.Lock()
	b.tabs[t] = struct{}{}
	b.mu.Unlock()
	// A socket that never says hello would hold a connection forever.
	helloTimer := time.AfterFunc(b.opts.HelloTimeout, func() { t.close(closeMalformed, "no hello") })
	defer func() {
		helloTimer.Stop()
		b.gone(t)
	}()
	for {
		// A frame over the read limit ends the socket here, and gone() answers its in-flight calls.
		_, data, err := ws.Read(context.Background())
		if err != nil {
			return
		}
		var message map[string]json.RawMessage
		if json.Unmarshal(data, &message) != nil {
			if !json.Valid(data) {
				go t.close(closeMalformed, "not json")
			}
			continue
		}
		switch kind, _ := str(message["type"]); kind {
		case "hello":
			helloTimer.Stop()
			b.hello(t, message)
		case "knows":
			// Lets the builder ask, before connecting, whether this agent would re-pair its code. It pairs nothing,
			// and answers yes only for the last pairing's code from the same site, so it tells a page elsewhere nothing.
			code, isString := str(message["code"])
			b.mu.Lock()
			known := isString && b.last != nil && b.last.code == code && b.last.origin == t.origin
			b.mu.Unlock()
			t.send(map[string]any{"type": "knows", "known": known})
		case "tools", "result":
			b.answer(t, kind, message)
		}
	}
}

func (b *Bridge) hello(t *tab, message map[string]json.RawMessage) {
	// One hello per socket: a second one, including from the paired tab, is ignored.
	b.mu.Lock()
	again := t.helloed
	t.helloed = true
	b.mu.Unlock()
	if again {
		return
	}
	if v, isNumber := number(message["v"]); !isNumber || v != ProtocolVersion {
		if isNumber && v > ProtocolVersion {
			t.refuse(textPluginTooOld)
		} else {
			t.refuse(textBuilderTooOld)
		}
		return
	}
	code, isString := str(message["code"])
	if !isString || !codePattern.MatchString(code) {
		go t.close(closeMalformed, "bad code")
		return
	}
	title, _ := str(message["title"])
	if runes := []rune(title); len(runes) > 120 {
		title = string(runes[:120])
	}

	b.mu.Lock()
	// The tab this agent was paired with, back after a drop: same code, same site. The site check stops a page
	// elsewhere from guessing its way into a pairing.
	if b.last != nil && b.last.code == code && b.last.origin == t.origin {
		previous := b.paired
		t.title = title
		b.paired = &pairing{tab: t, code: code}
		b.mu.Unlock()
		if previous != nil && previous.tab != t {
			go previous.tab.close(closeReplaced, "replaced")
		}
		t.send(map[string]any{"type": "paired"})
		return
	}
	refusal := ""
	if len(b.pending) >= b.opts.MaxPending {
		refusal = textTooManyTabs
	} else if _, taken := b.pending[code]; taken {
		refusal = textCodeTaken
	} else {
		t.title = title
		b.pending[code] = t
	}
	b.mu.Unlock()
	if refusal != "" {
		t.refuse(refusal)
	}
}

func (b *Bridge) answer(t *tab, kind string, message map[string]json.RawMessage) {
	id, isNumber := number(message["id"])
	if !isNumber {
		return
	}
	b.mu.Lock()
	c := b.calls[int64(id)]
	if c == nil || c.tab != t || float64(int64(id)) != id {
		b.mu.Unlock()
		return
	}
	delete(b.calls, int64(id))
	b.mu.Unlock()
	if kind == "tools" {
		tools := message["tools"]
		if len(tools) == 0 || string(tools) == "null" {
			tools = json.RawMessage("[]")
		}
		c.done <- reply{tools: tools}
	} else {
		c.done <- reply{Result: Result{Text: textOf(message["text"])}}
	}
}

func (b *Bridge) gone(t *tab) {
	b.mu.Lock()
	delete(b.tabs, t)
	for code, waiting := range b.pending {
		if waiting == t {
			delete(b.pending, code)
		}
	}
	if b.paired != nil && b.paired.tab == t {
		b.paired = nil
	}
	var orphaned []*call
	for id, c := range b.calls {
		if c.tab == t {
			delete(b.calls, id)
			orphaned = append(orphaned, c)
		}
	}
	b.mu.Unlock()
	for _, c := range orphaned {
		c.done <- reply{Result: Result{Text: textDisconnected, IsError: true}}
	}
}

// Connect pairs the agent with the tab waiting with this code.
func (b *Bridge) Connect(rawCode string) Result {
	if b.listener == nil {
		return Result{Text: textPortsBusy, IsError: true}
	}
	code := strings.ToUpper(strings.TrimSpace(rawCode))
	b.mu.Lock()
	if b.paired != nil && b.paired.code == code {
		title := b.paired.tab.title
		b.mu.Unlock()
		return Result{Text: textConnected(title)}
	}
	t := b.pending[code]
	if t == nil {
		b.mu.Unlock()
		return Result{Text: textWrongCode(code), IsError: true}
	}
	delete(b.pending, code)
	// One tab at a time: the tab paired before is closed and goes back to waiting with its own code.
	previous := b.paired
	b.paired = &pairing{tab: t, code: code}
	b.last = &lastPairing{code: code, origin: t.origin}
	b.mu.Unlock()
	if previous != nil {
		go previous.tab.close(closeReplaced, "replaced")
	}
	t.send(map[string]any{"type": "paired"})
	return Result{Text: textConnected(t.title)}
}

// ListTools returns the paired tab's tools as indented JSON.
func (b *Bridge) ListTools(ctx context.Context) Result {
	r := b.request(ctx, map[string]any{"type": "list"})
	if r.tools == nil {
		return r.Result
	}
	var indented bytes.Buffer
	if json.Indent(&indented, r.tools, "", "  ") != nil {
		return Result{Text: string(r.tools)}
	}
	return Result{Text: indented.String()}
}

// CallTool runs one of the paired tab's tools with the agent's input, passed through as it came.
func (b *Bridge) CallTool(ctx context.Context, name string, input json.RawMessage) Result {
	if len(input) == 0 || string(input) == "null" {
		input = json.RawMessage("{}")
	}
	return b.request(ctx, map[string]any{"type": "call", "tool": name, "input": input}).Result
}

// No timeout: a theme change waits for the user's answer in the builder. A call ends when the tab answers, the tab
// disconnects, or the agent cancels.
func (b *Bridge) request(ctx context.Context, message map[string]any) reply {
	b.mu.Lock()
	if b.paired == nil {
		b.mu.Unlock()
		return reply{Result: Result{Text: textNotPaired, IsError: true}}
	}
	if ctx.Err() != nil {
		b.mu.Unlock()
		return reply{Result: Result{Text: textCancelled, IsError: true}}
	}
	t := b.paired.tab
	b.nextID++
	id := b.nextID
	c := &call{tab: t, done: make(chan reply, 1)}
	b.calls[id] = c
	b.mu.Unlock()

	message["id"] = id
	t.send(message)
	select {
	case r := <-c.done:
		return r
	case <-ctx.Done():
		b.mu.Lock()
		_, open := b.calls[id]
		delete(b.calls, id)
		b.mu.Unlock()
		if !open {
			return <-c.done // answered as the agent cancelled
		}
		t.send(map[string]any{"type": "cancel", "id": id})
		return reply{Result: Result{Text: textCancelled, IsError: true}}
	}
}

func (t *tab) send(message any) {
	data, err := json.Marshal(message)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()
	_ = t.ws.Write(ctx, websocket.MessageText, data)
}

func (t *tab) close(code websocket.StatusCode, reason string) {
	_ = t.ws.Close(code, reason)
}

// The reason first, then the close: the builder shows the reason when the socket closes as refused.
func (t *tab) refuse(message string) {
	go func() {
		t.send(map[string]any{"type": "error", "message": message})
		t.close(closeRefused, "refused")
	}()
}

func str(raw json.RawMessage) (string, bool) {
	var s string
	if len(raw) == 0 || raw[0] != '"' || json.Unmarshal(raw, &s) != nil {
		return "", false
	}
	return s, true
}

func number(raw json.RawMessage) (float64, bool) {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return 0, false
	}
	n, ok := v.(float64)
	return n, ok
}

// The text a tab sends back, whatever JSON it came as; none is empty.
func textOf(raw json.RawMessage) string {
	if s, ok := str(raw); ok {
		return s
	}
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	return string(raw)
}

// JSON-quoted, without escaping <, > and & as < etc.
func quote(s string) string {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(s)
	return strings.TrimSuffix(out.String(), "\n")
}
