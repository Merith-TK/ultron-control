# Claude Notes — Ultron Control

## Project Overview
Go-based API server for controlling **ComputerCraft (CC:Tweaked) turtles** in Minecraft. Turtles connect via WebSocket and can be commanded via HTTP REST API. Think: remote control panel for Minecraft automation bots.

- **Repo**: github.com/merith-tk/ultron-control
- **Language**: Go 1.18
- **Default port**: 3300
- **Config**: `config.toml` (TOML format)

## Tech Stack
- `github.com/gorilla/mux` — HTTP routing
- `github.com/gorilla/websocket` — WebSocket (turtle↔server)
- `github.com/pelletier/go-toml` — config parsing
- Lua (CC:Tweaked) — turtle-side client scripts

## Architecture

### Server (Go)
```
main.go → config.ReadConfig() → api.CreateApiServer()
api/
  init.go   — server setup, WebSocket upgrader, sync command machinery
  setup.go  — route registration (InitModules)
  turtle.go — HTTP handlers + WS handler for turtles
  texture.go — texture/resourcepack endpoint
  global.go  — stub (not implemented)
config/
  config.go — TOML config read/write
```

### Routes
- `GET  /api/turtle`                — list all turtles
- `GET  /api/turtle/{id}`           — get turtle data
- `GET  /api/turtle/{id}/{action}`  — get specific field (pos, fuel, sight, inventory, etc.)
- `POST /api/turtle/{id}`           — send command (sync default, async via header)
- `GET  /api/turtle/ws`             — turtle WebSocket endpoint
- `GET  /api/texture/{asset}/{modid}/{texture}` — texture server
- `GET  /api/static/*`              — serve Lua files to turtles

### Command Execution Modes
- **Sync (default)**: Server sends `{command, requestId}` JSON over WS, waits up to 30s for response. Returns result inline.
- **Async (legacy)**: Appends to turtle's `CmdQueue`, turtle polls and executes on next heartbeat loop.
- Header: `X-Execution-Mode: sync|async`
- Sync only supports single command; async supports arrays.

### Concurrency Model
- Per-turtle `sync.Mutex` serializes commands (FIFO). Both sync and async commands use this.
- `pendingRequests map[string]chan CommandResponse` tracks in-flight sync commands by requestId.
- Turtle connections tracked in `ActiveConnections{turtleConnections map[int]*websocket.Conn}`.

### Turtle Struct
```go
Turtle { Name, ID, Inventory[16], SelectedSlot, Pos{X,Y,Z,R,Rname},
         Fuel{Current,Max}, Sight{Up,Down,Front}, CmdResult, CmdQueue[]string,
         Misc, HeartBeat }
```
Global `var Turtles []Turtle` — in-memory, no persistence.

## Lua Client (static/)
- `startup.lua` — entry: auto-updates scripts from server, runs turtle.lua
- `ultron.lua`  — core library: WS connection, processCmd, handleSyncCommand, config
- `turtle.lua`  — main loop: apiLoop (sends heartbeat data), recieveOrders, processCmdQueue
- `pastebin.lua` — pastebin helper

### Lua Loop Structure
```lua
parallel.waitForAll(apiLoop, ultron.recieveOrders, processCmdQueue)
-- apiLoop: sends turtle state every 0.5s (or on inventory event)
-- recieveOrders: reads WS messages, dispatches sync vs legacy commands
-- processCmdQueue: executes ultron.cmd sequentially
```

### Sync Command Flow (Lua side)
1. `recieveOrders` gets JSON `{command, requestId}` from WS
2. Calls `ultron.handleSyncCommand(parsedData)`
3. Runs `ultron.processCmd(command)` → executes via `load()` + `pcall()`
4. Sends back `{requestId, result, success}` JSON over WS
5. Go server's `handleCommandResponse()` routes response to waiting channel

## Current Branch: rework/websocket
The branch is implementing the sync command infrastructure (now done) and planning a multi-client GUI WebSocket broadcast system. Key TODOs from `.copilot/api-rework-plan.md`:
- [ ] `/api/gui/ws` endpoint for GUI clients
- [ ] `ClientManager` to broadcast turtle updates to GUI subscribers
- [ ] Event-driven Lua refactor (separate data collection from communication)
- [ ] Message typing system (`turtle_update`, `command_result`, `turtle_connected`)

## Known Issues / Notes
- `var Turtles []Turtle` is in-memory only — **intentional design**. Persisting turtle state caused desync errors (turtle "mini strokes"). State is authoritative on the turtle itself; API C&C clients are responsible for storing data. Early plans for a `api/world` world-cache endpoint were never built.
- `turtle.lua` line 123 has a comment/code merge artifact (comment on same line as `local function`)
- `handleLegacyTurtleData` has a comment: "Don't even know why I did this, or if it is even needed"
- `ultron.ws("close")` has a bug: uses `type` instead of `connectionType` for the close check
- `api/texture.go` — not reviewed yet
- Async commands hold the per-turtle mutex for the entire queue-append operation, which may cause unexpected blocking if a long sync command is in flight

## Self-Correction Rules
- DO NOT use `.devcontainer/do-not-commit/` — it is a symlink to `~/.claude` for devcontainer persistence. Writing there is confusing/wrong. Use `.claude/` in project root per CLAUDE.md.
