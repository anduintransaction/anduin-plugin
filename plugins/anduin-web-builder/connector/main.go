// The Anduin Web Builder plugin's MCP server: the agent's three tools, backed by the local builder bridge. A single
// native program, so the plugin needs no Node.js or other runtime on the user's computer.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const version = "0.3.0" // keep in step with .claude-plugin/plugin.json and manifest.json

type connectInput struct {
	Code string `json:"code" jsonschema:"The 6-character code from the builder, e.g. K7Q2PX"`
}

type callInput struct {
	Name  string         `json:"name" jsonschema:"The builder tool's name, from list_builder_tools"`
	Input map[string]any `json:"input,omitempty" jsonschema:"The tool's input, as an object"`
}

func newServer(bridge *Bridge) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "anduin-builder", Version: version}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name: "connect_builder",
		Description: "Connects to the Anduin landing page builder tab the user opened for AI. The user clicks " +
			"\"Connect your agent\" in the builder and gives you the 6-character code it shows. Call this first.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in connectInput) (*mcp.CallToolResult, any, error) {
		return toolResult(bridge.Connect(in.Code)), nil, nil
	})
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_builder_tools",
		Description: "Lists the connected builder's tools: name, description and input schema. Call after connect_builder.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return toolResult(bridge.ListTools(ctx)), nil, nil
	})
	mcp.AddTool(server, &mcp.Tool{
		Name: "call_builder_tool",
		Description: "Runs one of the builder's tools on the connected tab and returns its result. Start with " +
			"get_guide and follow it.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in callInput) (*mcp.CallToolResult, any, error) {
		return toolResult(bridge.CallTool(ctx, in.Name, in.Input)), nil, nil
	})
	return server
}

func toolResult(r Result) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: r.Text}}, IsError: r.IsError}
}

func main() {
	bridge := Start(Options{})
	if port := bridge.Port(); port != 0 {
		fmt.Fprintf(os.Stderr, "[anduin-builder] listening on 127.0.0.1:%d\n", port)
	} else {
		fmt.Fprintln(os.Stderr, "[anduin-builder] no free port")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Returns when the agent closes stdin, which is how its host ends the server.
	if err := newServer(bridge).Run(ctx, &mcp.StdioTransport{}); err != nil && ctx.Err() == nil {
		fmt.Fprintf(os.Stderr, "[anduin-builder] %v\n", err)
	}
	bridge.Close()
}
