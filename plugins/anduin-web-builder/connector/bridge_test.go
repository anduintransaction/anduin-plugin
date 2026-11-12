package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// Not the real range, so a running connector does not interfere.
var testPorts = []int{48811, 48812}

const appOrigin = "https://app.anduin.test"

func tick() { time.Sleep(30 * time.Millisecond) }

func startBridge(t *testing.T, opts Options) *Bridge {
	t.Helper()
	opts.Ports = testPorts
	b := Start(opts)
	t.Cleanup(b.Close)
	return b
}

// A fake builder tab: it queues what the bridge sends and records how its socket closed.
type fakeTab struct {
	t      *testing.T
	ws     *websocket.Conn
	inbox  chan map[string]any
	closed chan websocket.StatusCode
}

func dial(t *testing.T, port int, origin string) *fakeTab {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	header := http.Header{}
	if origin != "" {
		header.Set("Origin", origin)
	}
	ws, _, err := websocket.Dial(ctx, "ws://127.0.0.1:"+strconv.Itoa(port), &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	f := &fakeTab{t: t, ws: ws, inbox: make(chan map[string]any, 16), closed: make(chan websocket.StatusCode, 1)}
	go func() {
		for {
			_, data, err := ws.Read(context.Background())
			if err != nil {
				f.closed <- websocket.CloseStatus(err)
				return
			}
			var message map[string]any
			_ = json.Unmarshal(data, &message)
			f.inbox <- message
		}
	}()
	t.Cleanup(func() { _ = ws.CloseNow() })
	return f
}

type hello struct {
	code, title, origin string
	v                   any
}

func defaultHello() hello {
	return hello{code: "ABC234", title: "Home page", origin: appOrigin, v: ProtocolVersion}
}

// A tab that has said hello, as the builder does once its socket opens.
func openTab(t *testing.T, port int, edit ...func(*hello)) *fakeTab {
	t.Helper()
	h := defaultHello()
	for _, e := range edit {
		e(&h)
	}
	f := dial(t, port, h.origin)
	f.send(map[string]any{"type": "hello", "v": h.v, "code": h.code, "title": h.title})
	tick()
	return f
}

func withCode(code string) func(*hello)     { return func(h *hello) { h.code = code } }
func withOrigin(origin string) func(*hello) { return func(h *hello) { h.origin = origin } }
func withVersion(v any) func(*hello)        { return func(h *hello) { h.v = v } }

func (f *fakeTab) send(message any) {
	data, _ := json.Marshal(message)
	if err := f.ws.Write(context.Background(), websocket.MessageText, data); err != nil {
		f.t.Fatalf("send: %v", err)
	}
}

func (f *fakeTab) next() map[string]any {
	f.t.Helper()
	select {
	case message := <-f.inbox:
		return message
	case <-time.After(2 * time.Second):
		f.t.Fatal("no message from the bridge")
		return nil
	}
}

func (f *fakeTab) waitClosed() websocket.StatusCode {
	f.t.Helper()
	select {
	case code := <-f.closed:
		return code
	case <-time.After(3 * time.Second):
		f.t.Fatal("the socket was not closed")
		return 0
	}
}

func eq(t *testing.T, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

// Runs a blocking bridge call in the background, like the agent waiting on a tool.
func async(do func() Result) <-chan Result {
	out := make(chan Result, 1)
	go func() { out <- do() }()
	return out
}

func await(t *testing.T, results <-chan Result) Result {
	t.Helper()
	select {
	case r := <-results:
		return r
	case <-time.After(2 * time.Second):
		t.Fatal("the call did not finish")
		return Result{}
	}
}

func listen(t *testing.T, port int) net.Listener {
	t.Helper()
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = l.Close() })
	return l
}

func TestListensOnLoopbackOnTheFirstFreePort(t *testing.T) {
	listen(t, testPorts[0])
	b := startBridge(t, Options{})
	eq(t, b.Port(), testPorts[1])
	eq(t, b.listener.Addr().(*net.TCPAddr).IP.String(), "127.0.0.1")
}

func TestEveryPortTakenConnectSaysSo(t *testing.T) {
	for _, port := range testPorts {
		listen(t, port)
	}
	b := startBridge(t, Options{})
	eq(t, b.Port(), 0)
	eq(t, b.Connect("ABC234"), Result{Text: textPortsBusy, IsError: true})
}

func TestPairsOnlyWithTheHandedOverCodeCaseInsensitively(t *testing.T) {
	b := startBridge(t, Options{})
	tab := openTab(t, b.Port())
	wrong := b.Connect("ZZZ999")
	if !wrong.IsError || !strings.Contains(wrong.Text, "No builder is waiting with code ZZZ999") {
		t.Fatalf("wrong code: %#v", wrong)
	}
	right := b.Connect(" abc234 ")
	if right.IsError || !strings.Contains(right.Text, `(page titled "Home page")`) {
		t.Fatalf("right code: %#v", right)
	}
	eq(t, tab.next(), map[string]any{"type": "paired"})
	// Connecting again with the same code is not an error.
	eq(t, b.Connect("ABC234").IsError, false)
}

func TestAnUnpairedTabNeverReceivesCalls(t *testing.T) {
	b := startBridge(t, Options{})
	tab := openTab(t, b.Port())
	eq(t, b.CallTool(context.Background(), "read_page", nil), Result{Text: textNotPaired, IsError: true})
	tick()
	eq(t, len(tab.inbox), 0)
}

func TestRelaysListAndCallAndReturnsTheAnswers(t *testing.T) {
	b := startBridge(t, Options{})
	tab := openTab(t, b.Port())
	b.Connect("ABC234")
	tab.next() // paired

	listing := async(func() Result { return b.ListTools(context.Background()) })
	list := tab.next()
	eq(t, list["type"], "list")
	tab.send(map[string]any{"type": "tools", "id": list["id"], "tools": []any{
		map[string]any{"name": "read_page", "description": "Reads", "inputSchema": map[string]any{}},
	}})
	if listed := await(t, listing); !strings.Contains(listed.Text, `"read_page"`) || listed.IsError {
		t.Fatalf("listed: %#v", listed)
	}

	calling := async(func() Result {
		return b.CallTool(context.Background(), "edit_page", map[string]any{"ops": []any{1}})
	})
	c := tab.next()
	eq(t, []any{c["type"], c["tool"], c["input"]}, []any{"call", "edit_page", map[string]any{"ops": []any{1.0}}})
	tab.send(map[string]any{"type": "result", "id": c["id"], "text": "Applied 1 op."})
	eq(t, await(t, calling), Result{Text: "Applied 1 op."})
}

func TestAnAgentCancellationReachesTheTabAsCancel(t *testing.T) {
	b := startBridge(t, Options{})
	tab := openTab(t, b.Port())
	b.Connect("ABC234")
	tab.next()
	ctx, cancel := context.WithCancel(context.Background())
	calling := async(func() Result { return b.CallTool(ctx, "update_theme", nil) })
	c := tab.next()
	cancel()
	eq(t, tab.next(), map[string]any{"type": "cancel", "id": c["id"]})
	eq(t, await(t, calling), Result{Text: textCancelled, IsError: true})
}

func TestTheTabGoingAwayMidCallAnswersItAsDisconnected(t *testing.T) {
	b := startBridge(t, Options{})
	tab := openTab(t, b.Port())
	b.Connect("ABC234")
	tab.next()
	calling := async(func() Result { return b.CallTool(context.Background(), "edit_page", nil) })
	tab.next()
	_ = tab.ws.Close(websocket.StatusNormalClosure, "")
	eq(t, await(t, calling), Result{Text: textDisconnected, IsError: true})
	eq(t, b.CallTool(context.Background(), "read_page", nil), Result{Text: textNotPaired, IsError: true})
}

func TestANewPairingReplacesTheOldOneAndClosesTheOldTab(t *testing.T) {
	b := startBridge(t, Options{})
	first := openTab(t, b.Port(), withCode("ABC234"))
	second := openTab(t, b.Port(), withCode("XYZ789"))
	b.Connect("ABC234")
	first.next()
	b.Connect("XYZ789")
	eq(t, second.next(), map[string]any{"type": "paired"})
	eq(t, first.waitClosed(), closeReplaced)
}

func TestANewerProtocolIsRefusedWithAReason(t *testing.T) {
	b := startBridge(t, Options{})
	tab := openTab(t, b.Port(), withVersion(ProtocolVersion+1))
	eq(t, tab.next(), map[string]any{"type": "error", "message": textPluginTooOld})
	eq(t, tab.waitClosed(), closeRefused)
}

func TestAnOlderProtocolIsRefusedAsBuilderTooOld(t *testing.T) {
	b := startBridge(t, Options{})
	tab := openTab(t, b.Port(), withVersion(0))
	eq(t, tab.next(), map[string]any{"type": "error", "message": textBuilderTooOld})
	eq(t, tab.waitClosed(), closeRefused)
}

func TestNoMoreThanMaxPendingTabsWait(t *testing.T) {
	b := startBridge(t, Options{MaxPending: 2})
	openTab(t, b.Port(), withCode("AAAAA2"))
	openTab(t, b.Port(), withCode("AAAAA3"))
	third := openTab(t, b.Port(), withCode("AAAAA4"))
	eq(t, third.next(), map[string]any{"type": "error", "message": textTooManyTabs})
	eq(t, third.waitClosed(), closeRefused)
}

func TestAFrameOverTheSizeCapClosesThatSocketAndTheBridgeKeepsServing(t *testing.T) {
	b := startBridge(t, Options{MaxPayload: 1024})
	big := dial(t, b.Port(), appOrigin)
	big.send(strings.Repeat("x", 4096))
	eq(t, big.waitClosed(), websocket.StatusMessageTooBig)
	tab := openTab(t, b.Port(), withCode("AAAAA2"))
	eq(t, b.Connect("AAAAA2").IsError, false)
	eq(t, tab.next(), map[string]any{"type": "paired"})
}

func TestASocketSaysHelloOnceEvenWhenPaired(t *testing.T) {
	b := startBridge(t, Options{})
	tab := openTab(t, b.Port(), withCode("ABC234"))
	second := map[string]any{"type": "hello", "v": ProtocolVersion, "code": "XYZ789", "title": "Second"}
	tab.send(second)
	tick()
	eq(t, b.Connect("XYZ789").IsError, true)
	eq(t, b.Connect("ABC234").IsError, false)
	eq(t, tab.next(), map[string]any{"type": "paired"})
	tab.send(second)
	tick()
	eq(t, b.Connect("XYZ789").IsError, true)
}

func TestASocketThatNeverSaysHelloIsClosed(t *testing.T) {
	b := startBridge(t, Options{HelloTimeout: 50 * time.Millisecond})
	silent := dial(t, b.Port(), appOrigin)
	eq(t, silent.waitClosed(), closeMalformed)
}

func TestNotJSONClosesTheSocket(t *testing.T) {
	b := startBridge(t, Options{})
	tab := dial(t, b.Port(), appOrigin)
	if err := tab.ws.Write(context.Background(), websocket.MessageText, []byte("{")); err != nil {
		t.Fatalf("send: %v", err)
	}
	eq(t, tab.waitClosed(), closeMalformed)
}

func TestACodeOutsideTheAlphabetIsRefusedAndNeverPairs(t *testing.T) {
	b := startBridge(t, Options{})
	tab := openTab(t, b.Port(), withCode("ABC01L"))
	eq(t, tab.waitClosed(), closeMalformed)
	refused := b.Connect("ABC01L")
	if !refused.IsError || !strings.Contains(refused.Text, "No builder is waiting with code ABC01L") {
		t.Fatalf("refused: %#v", refused)
	}
}

func TestASecondTabWithAWaitingCodeIsRefusedAndTheFirstKeepsIt(t *testing.T) {
	b := startBridge(t, Options{})
	openTab(t, b.Port(), withCode("ABC234"))
	second := openTab(t, b.Port(), withCode("ABC234"))
	eq(t, second.next(), map[string]any{"type": "error", "message": textCodeTaken})
	eq(t, second.waitClosed(), closeRefused)
	eq(t, b.Connect("ABC234").IsError, false)
}

func TestAResultFromAnotherTabIsIgnored(t *testing.T) {
	b := startBridge(t, Options{})
	tab := openTab(t, b.Port(), withCode("ABC234"))
	b.Connect("ABC234")
	tab.next() // paired
	calling := async(func() Result { return b.CallTool(context.Background(), "read_page", nil) })
	c := tab.next()

	other := openTab(t, b.Port(), withCode("AAAAA2"))
	other.send(map[string]any{"type": "result", "id": c["id"], "text": "forged"})
	tick()
	select {
	case r := <-calling:
		t.Fatalf("answered by another tab: %#v", r)
	default:
	}

	tab.send(map[string]any{"type": "result", "id": c["id"], "text": "real"})
	eq(t, await(t, calling), Result{Text: "real"})
}

// Pairs a tab, then drops it, leaving the bridge remembering its code and site.
func pairAndDrop(t *testing.T, b *Bridge) {
	t.Helper()
	first := openTab(t, b.Port())
	b.Connect("ABC234")
	first.next()
	_ = first.ws.Close(websocket.StatusNormalClosure, "")
	first.waitClosed()
	tick()
}

func TestThePairedTabBackWithItsCodeRePairsWithoutTheAgent(t *testing.T) {
	b := startBridge(t, Options{})
	pairAndDrop(t, b)
	back := openTab(t, b.Port())
	eq(t, back.next(), map[string]any{"type": "paired"})
	listed := async(func() Result { return b.ListTools(context.Background()) })
	request := back.next()
	back.send(map[string]any{"type": "tools", "id": request["id"], "tools": []any{}})
	eq(t, await(t, listed), Result{Text: "[]"})
}

func TestTheLastCodeFromAnotherSiteDoesNotRePair(t *testing.T) {
	b := startBridge(t, Options{})
	pairAndDrop(t, b)
	other := openTab(t, b.Port(), withOrigin("https://elsewhere.test"))
	tick()
	eq(t, len(other.inbox), 0)
	eq(t, b.ListTools(context.Background()), Result{Text: textNotPaired, IsError: true})
}

func TestKnowsSaysYesOnlyForTheLastCodeFromTheSameSiteAndPairsNothing(t *testing.T) {
	b := startBridge(t, Options{})
	pairAndDrop(t, b)
	ask := func(code any, origin string) any {
		probe := dial(t, b.Port(), origin)
		probe.send(map[string]any{"type": "knows", "code": code})
		answer := probe.next()
		_ = probe.ws.Close(websocket.StatusNormalClosure, "")
		return answer["known"]
	}
	eq(t, ask("ABC234", appOrigin), true)
	eq(t, ask("ABC234", "https://elsewhere.test"), false)
	eq(t, ask("XYZ789", appOrigin), false)
	eq(t, ask(nil, appOrigin), false)
	eq(t, b.ListTools(context.Background()), Result{Text: textNotPaired, IsError: true})
}
