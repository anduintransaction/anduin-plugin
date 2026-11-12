package main

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// The server as the agent's host sees it: its three tools, and a call refused before a builder is connected.
func TestServerListsItsToolsAndRefusesACallBeforePairing(t *testing.T) {
	ctx := context.Background()
	b := startBridge(t, Options{})
	clientEnd, serverEnd := mcp.NewInMemoryTransports()
	if _, err := newServer(b).Connect(ctx, serverEnd, nil); err != nil {
		t.Fatalf("server: %v", err)
	}
	session, err := mcp.NewClient(&mcp.Implementation{Name: "server-test", Version: "0"}, nil).Connect(ctx, clientEnd, nil)
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer session.Close()

	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}
	var names []string
	for _, tool := range listed.Tools {
		names = append(names, tool.Name)
	}
	slices.Sort(names)
	eq(t, names, []string{"call_builder_tool", "connect_builder", "list_builder_tools"})

	called, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "call_builder_tool",
		Arguments: map[string]any{"name": "read_page"},
	})
	if err != nil {
		t.Fatalf("tools/call: %v", err)
	}
	text := called.Content[0].(*mcp.TextContent).Text
	if !called.IsError || !strings.HasPrefix(text, "No builder connected") {
		t.Fatalf("call before pairing: %v %q", called.IsError, text)
	}
}

func TestRawInputKeepsTheAgentsBytes(t *testing.T) {
	eq(t, string(rawInput(json.RawMessage(`{"name":"edit_page","input":{"b":1,"a":{"d":2,"c":3}}}`))),
		`{"b":1,"a":{"d":2,"c":3}}`)
	eq(t, rawInput(json.RawMessage(`{"name":"read_page"}`)), json.RawMessage(nil))
}
