package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// textResult wraps a value as a text MCP result.
func textResult(v any) *mcp.CallToolResult {
	text, _ := json.MarshalIndent(v, "", "  ")
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(text)}}}
}

// errResult wraps a message as an error MCP result.
func errResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}
}

// luaResult runs a Lua command and returns its result string directly,
// without the command metadata wrapper. Used for specialized query tools.
func luaResult(id int, cmd string) (*mcp.CallToolResult, error) {
	raw, err := executeCommandSync(id, cmd, 15*time.Second)
	if err != nil {
		return nil, err
	}
	m, _ := raw.(map[string]interface{})
	if success, _ := m["success"].(bool); !success {
		errMsg, _ := m["error"].(string)
		return errResult("Lua error: " + errMsg), nil
	}
	payload, _ := m["result"].(string)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: payload}}}, nil
}

func buildMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ultron-control",
		Version: "0.1.0",
	}, nil)

	// ── Instructions ───────────────────────────────────────────────────────────

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_instructions",
		Description: "CALL THIS FIRST. Returns the agent guide — do's, don'ts, crafting rules, run_command usage, and CC:Tweaked API reference. Required reading before operating turtles.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		content, err := os.ReadFile(filepath.Join("mcp", "docs", "ultron", "agent-guide.md"))
		if err != nil {
			return errResult("Error reading agent guide: " + err.Error()), nil, nil
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(content)}}}, nil, nil
	})

	// ── Turtle state ───────────────────────────────────────────────────────────

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_turtles",
		Description: "List all connected turtles and their current state (position, fuel, inventory, sight).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return textResult(Turtles), nil, nil
	})

	type IDArgs struct {
		ID int `json:"id" jsonschema:"The turtle's CC computer ID"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_turtle",
		Description: "Get the full current state of a specific turtle (position, fuel, inventory, sight, misc data).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args IDArgs) (*mcp.CallToolResult, any, error) {
		for _, t := range Turtles {
			if t.ID == args.ID {
				return textResult(t), nil, nil
			}
		}
		return errResult("Turtle " + strconv.Itoa(args.ID) + " not found or not connected"), nil, nil
	})

	// ── Command execution ──────────────────────────────────────────────────────

	type RunCommandArgs struct {
		ID      int    `json:"id"      jsonschema:"The turtle's CC computer ID"`
		Command string `json:"command" jsonschema:"Lua code to run. Use 'return' to get values back."`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "run_command",
		Description: "Execute Lua code on a turtle synchronously. Returns the full command result including timing, success flag, and output. Keep loops to ≤8 steps to stay within the 30s timeout.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args RunCommandArgs) (*mcp.CallToolResult, any, error) {
		result, err := executeCommandSync(args.ID, args.Command, 30*time.Second)
		if err != nil {
			return errResult("Error: " + err.Error()), nil, nil
		}
		return textResult(result), nil, nil
	})

	// ── Inventory ──────────────────────────────────────────────────────────────

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_inventory",
		Description: "Get a compact inventory snapshot: slot, name, displayName, count, hasEnchantments for every occupied slot. Prefer this over get_turtle when you only need inventory data.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args IDArgs) (*mcp.CallToolResult, any, error) {
		res, err := luaResult(args.ID, `
local slots = {}
for i = 1, 16 do
    local item = turtle.getItemDetail(i)
    if item then
        slots[#slots+1] = {
            slot            = i,
            name            = item.name,
            displayName     = item.displayName,
            count           = item.count,
            hasEnchantments = (item.enchantments ~= nil and #item.enchantments > 0),
        }
    end
end
return textutils.serializeJSON(slots)`)
		if err != nil {
			return errResult("Error: " + err.Error()), nil, nil
		}
		return res, nil, nil
	})

	type SlotArgs struct {
		ID   int `json:"id"   jsonschema:"The turtle's CC computer ID"`
		Slot int `json:"slot" jsonschema:"Inventory slot (1–16)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_inventory_slot",
		Description: "Get full detail for a single inventory slot including NBT data, enchantments, and durability. Uses turtle.getItemDetail(slot, true).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args SlotArgs) (*mcp.CallToolResult, any, error) {
		cmd := fmt.Sprintf(`
local item = turtle.getItemDetail(%d, true)
return item and textutils.serializeJSON(item) or "null"`, args.Slot)
		res, err := luaResult(args.ID, cmd)
		if err != nil {
			return errResult("Error: " + err.Error()), nil, nil
		}
		return res, nil, nil
	})

	// ── Waypoints ──────────────────────────────────────────────────────────────

	type WaypointArgs struct {
		ID   int    `json:"id"   jsonschema:"The turtle's CC computer ID"`
		Name string `json:"name" jsonschema:"Waypoint name"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_waypoint",
		Description: "Save the turtle's current position as a named waypoint in agentmcp.waypoints. Requires cfg/agenttools to be enabled on the turtle.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args WaypointArgs) (*mcp.CallToolResult, any, error) {
		cmd := fmt.Sprintf(`return agentmcp.setWaypoint(%q)`, args.Name)
		res, err := luaResult(args.ID, cmd)
		if err != nil {
			return errResult("Error: " + err.Error()), nil, nil
		}
		return res, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_waypoints",
		Description: "Return all named waypoints saved for a turtle as JSON. Requires cfg/agenttools to be enabled.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args IDArgs) (*mcp.CallToolResult, any, error) {
		res, err := luaResult(args.ID, `return agentmcp.getWaypoints()`)
		if err != nil {
			return errResult("Error: " + err.Error()), nil, nil
		}
		return res, nil, nil
	})

	// ── Documentation ──────────────────────────────────────────────────────────

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_docs",
		Description: "List available documentation sets (name, repo, root path).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		manifest, err := ReadManifest()
		if err != nil {
			return errResult("Error reading manifest: " + err.Error()), nil, nil
		}
		return textResult(manifest.Docs), nil, nil
	})

	type GetDocArgs struct {
		Name string `json:"name" jsonschema:"Doc set name (from list_docs)"`
		File string `json:"file" jsonschema:"File path relative to the doc set root, e.g. 'turtle.md'"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_doc",
		Description: "Get the contents of a documentation file. Files are served from mcp/docs/<name>/<root>/<file>.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args GetDocArgs) (*mcp.CallToolResult, any, error) {
		manifest, err := ReadManifest()
		if err != nil {
			return errResult("Error reading manifest: " + err.Error()), nil, nil
		}
		var root string
		for _, entry := range manifest.Docs {
			if entry.Name == args.Name {
				root = entry.Root
				break
			}
		}
		if root == "" {
			return errResult("Unknown doc set: " + args.Name), nil, nil
		}
		content, err := os.ReadFile(filepath.Join("mcp", "docs", args.Name, root, args.File))
		if err != nil {
			return errResult("File not found: " + err.Error()), nil, nil
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(content)}}}, nil, nil
	})

	// ── World map ──────────────────────────────────────────────────────────────

	type XYZArgs struct {
		X int `json:"x" jsonschema:"World X coordinate"`
		Y int `json:"y" jsonschema:"World Y coordinate"`
		Z int `json:"z" jsonschema:"World Z coordinate"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_block",
		Description: "Look up a specific world coordinate in the accumulated turtle sight map. Returns null if never observed.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args XYZArgs) (*mcp.CallToolResult, any, error) {
		b, err := GetBlock(args.X, args.Y, args.Z)
		if err != nil {
			return errResult("Error: " + err.Error()), nil, nil
		}
		return textResult(b), nil, nil
	})

	type FindBlockArgs struct {
		Name  string `json:"name"            jsonschema:"Block name or partial name, e.g. 'spruce_log'"`
		Limit int    `json:"limit,omitempty" jsonschema:"Max results (default 100, max 500)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "find_block",
		Description: "Search the world map for all known locations of a block type. Partial name matching supported (e.g. 'spruce_log' matches 'minecraft:spruce_log').",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args FindBlockArgs) (*mcp.CallToolResult, any, error) {
		if args.Limit == 0 {
			args.Limit = 100
		}
		results, err := FindBlock(args.Name, args.Limit)
		if err != nil {
			return errResult("Error: " + err.Error()), nil, nil
		}
		return textResult(results), nil, nil
	})

	type GetRegionArgs struct {
		X1 int `json:"x1" jsonschema:"Bounding box corner 1 X"`
		Y1 int `json:"y1" jsonschema:"Bounding box corner 1 Y"`
		Z1 int `json:"z1" jsonschema:"Bounding box corner 1 Z"`
		X2 int `json:"x2" jsonschema:"Bounding box corner 2 X"`
		Y2 int `json:"y2" jsonschema:"Bounding box corner 2 Y"`
		Z2 int `json:"z2" jsonschema:"Bounding box corner 2 Z"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_region",
		Description: "Get all known blocks within a bounding box from the world map. Results capped at 1000. Use to understand terrain before planning movement.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args GetRegionArgs) (*mcp.CallToolResult, any, error) {
		results, err := GetRegion(args.X1, args.Y1, args.Z1, args.X2, args.Y2, args.Z2)
		if err != nil {
			return errResult("Error: " + err.Error()), nil, nil
		}
		return textResult(results), nil, nil
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
