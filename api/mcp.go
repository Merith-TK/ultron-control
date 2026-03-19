package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func buildMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ultron-control",
		Version: "0.1.0",
	}, nil)

	// get_instructions — required reading, call this first
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_instructions",
		Description: "CALL THIS FIRST. Returns the agent guide for this MCP server — do's, don'ts, crafting rules, run_command usage, and CC:Tweaked API reference. Required reading before operating turtles.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		content, err := os.ReadFile(filepath.Join("mcp", "docs", "ultron", "agent-guide.md"))
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Error reading agent guide: " + err.Error()}},
				IsError: true,
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(content)}},
		}, nil, nil
	})

	// list_turtles
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_turtles",
		Description: "List all connected CC:Tweaked turtles and their current state (position, fuel, inventory, sight).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		text, _ := json.MarshalIndent(Turtles, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(text)}},
		}, nil, nil
	})

	// get_turtle
	type GetTurtleArgs struct {
		ID int `json:"id" jsonschema:"The turtle's CC computer ID"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_turtle",
		Description: "Get the full current state of a specific turtle by its computer ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GetTurtleArgs) (*mcp.CallToolResult, any, error) {
		for _, t := range Turtles {
			if t.ID == args.ID {
				text, _ := json.MarshalIndent(t, "", "  ")
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: string(text)}},
				}, nil, nil
			}
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Turtle " + strconv.Itoa(args.ID) + " not found or not connected"}},
			IsError: true,
		}, nil, nil
	})

	// run_command
	type RunCommandArgs struct {
		ID      int    `json:"id"      jsonschema:"The turtle's CC computer ID"`
		Command string `json:"command" jsonschema:"Lua code to run on the turtle. Use 'return' to get values back, e.g. 'return turtle.forward()'"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "run_command",
		Description: "Execute a Lua expression or statement on a turtle synchronously and return the result.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args RunCommandArgs) (*mcp.CallToolResult, any, error) {
		result, err := executeCommandSync(args.ID, args.Command, 30*time.Second)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Error: " + err.Error()}},
				IsError: true,
			}, nil, nil
		}
		text, _ := json.MarshalIndent(result, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(text)}},
		}, nil, nil
	})

	// list_docs
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_docs",
		Description: "List available documentation sets defined in the manifest (name, repo, root path).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		manifest, err := ReadManifest()
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Error reading manifest: " + err.Error()}},
				IsError: true,
			}, nil, nil
		}
		text, _ := json.MarshalIndent(manifest.Docs, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(text)}},
		}, nil, nil
	})

	// get_doc
	type GetDocArgs struct {
		Name string `json:"name" jsonschema:"Doc set name (from list_docs)"`
		File string `json:"file" jsonschema:"File path relative to the doc set root, e.g. 'turtle.md'"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_doc",
		Description: "Get the contents of a file from a documentation set. Files are served from mcp/docs/<name>/<root>/<file>.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GetDocArgs) (*mcp.CallToolResult, any, error) {
		manifest, err := ReadManifest()
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Error reading manifest: " + err.Error()}},
				IsError: true,
			}, nil, nil
		}
		var root string
		for _, entry := range manifest.Docs {
			if entry.Name == args.Name {
				root = entry.Root
				break
			}
		}
		if root == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Unknown doc set: " + args.Name}},
				IsError: true,
			}, nil, nil
		}
		content, err := os.ReadFile(filepath.Join("mcp", "docs", args.Name, root, args.File))
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "File not found: " + err.Error()}},
				IsError: true,
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(content)}},
		}, nil, nil
	})

	return server
}

var MCPHandler http.Handler

func init() {
	s := buildMCPServer()
	MCPHandler = mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s
	}, &mcp.StreamableHTTPOptions{
		Stateless:                  true,
		DisableLocalhostProtection: true,
	})
}
