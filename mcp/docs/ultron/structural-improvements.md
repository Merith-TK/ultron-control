# Ultron Control — Structural Improvements

This document describes proposed changes to improve agent situational awareness, area tracking, and persistent state management. It is written for developers implementing changes to the server or turtle firmware.

---

## Current Limitations

1. **No area memory.** The agent has no way to know which areas have been explored, where resources were found, or what terrain looks like beyond the current `sight` (3 adjacent blocks).
2. **No waypoint system.** Returning to a known position requires the agent to remember coordinates from the conversation context, which is lost across sessions.
3. **No structured misc schema.** `ultron.data.misc` is a free-form table with no conventions — agents in different sessions may write incompatible keys.
4. **Command sizing is guesswork.** There's no feedback mechanism to tell the agent it's close to timing out.

---

## Existing Mechanism: `ultron.data.misc`

The turtle firmware already supports arbitrary Lua table data in `ultron.data.misc`. This table is serialized and sent in every heartbeat, appearing as the `misc` field in `get_turtle` output.

This is the correct place for persistent agent state. All proposals below build on it.

---

## Proposed: Standardized `misc` Schema

All agent data lives under `ultron.data.misc.agentmcp` to avoid colliding with other uses of the misc table.

```lua
ultron.data.misc.agentmcp = {
    -- Waypoints: named positions the agent has recorded
    waypoints = {
        home = { x=7, y=73, z=-12, r=0 },
        base_chest = { x=5, y=73, z=-12 },
    },

    -- Current task state
    task = {
        name = "mining",       -- human-readable task name
        phase = "descending",  -- current phase within task
        target = { x=0, y=60, z=0 },  -- optional target position
    },

    -- Explored chunks: coarse grid (16x16 areas), each entry records what was found
    -- key = "cx,cz" where cx=math.floor(x/16), cz=math.floor(z/16)
    explored = {
        ["0,-1"] = {
            scanned_at = 1742300000,  -- os.epoch("utc")
            resources = {
                "minecraft:spruce_log",
                "minecraft:coal_ore",
            },
        },
        ["0,-2"] = {
            scanned_at = 1742300120,
            resources = {},  -- explored but nothing notable found
        },
    },
}
```

**Schema rules:**
- `agentmcp` — always the top-level namespace; never write directly into `misc` root
- `waypoints` — table of named `{x,y,z[,r]}` entries; use human-readable names
- `task` — current top-level task; overwrite when starting a new task
- `explored` — coarse 16×16 chunk grid; each entry records `scanned_at` timestamp and a `resources` list of item IDs found in that area. Prune entries that are very old or after a known depletion (e.g. a forest was fully cut).

---

## Proposed: MCP Tool Additions

### `set_waypoint(id, name, pos)`
Shorthand for writing to `ultron.data.misc.agentmcp.waypoints[name]`. Saves the current turtle position under a name without having to write full Lua.

```
set_waypoint(id=3, name="home")       -- uses current turtle pos
set_waypoint(id=3, name="mine_entrance", x=4, y=65, z=-8)
```

### `get_waypoints(id)`
Returns `ultron.data.misc.agentmcp.waypoints` from the latest heartbeat. Faster than reading full turtle state.

### `list_area(id)`
Returns `ultron.data.misc.agentmcp.explored` from the latest heartbeat. Each chunk entry includes its resource list, letting the agent answer "where is spruce wood?" by scanning the map rather than issuing a `run_command`.

---

## Proposed: Turtle-Side Area Scanning

Add a `scan_area(radius)` function to the turtle firmware that:
1. Walks a grid pattern within `radius` blocks
2. Calls `turtle.inspect()` at each position
3. Records notable blocks (non-air, non-stone, non-dirt) into the `resources` list of the relevant chunk key in `ultron.data.misc.agentmcp.explored`
4. Marks visited chunk regions with a `scanned_at` timestamp
5. Returns home after scan completes

This enables the agent to issue a single `run_command` like:
```lua
return scan_area(8)  -- scan 8-block radius, returns summary
```

---

## Proposed: Command Budget Feedback

Add a `__elapsed` field to command responses indicating how many milliseconds the command took. This lets the agent self-regulate loop sizes:

```json
{
  "result": "16",
  "__elapsed_ms": 12400
}
```

If `__elapsed_ms > 20000`, the agent should split the next similar operation into smaller chunks.

---

## Proposed: Agent State Lifecycle

Define a standard lifecycle for agent tasks using `misc.agentmcp.task`:

```lua
-- Before starting
ultron.data.misc.agentmcp.task = { name="chest_craft", phase="init" }
ultron.data.misc.agentmcp.waypoints.home = { x=7, y=73, z=-12, r=0 }

-- During execution
ultron.data.misc.agentmcp.task.phase = "finding_wood"

-- On completion
ultron.data.misc.agentmcp.task.phase = "done"

-- On failure
ultron.data.misc.agentmcp.task.phase = "failed"
ultron.data.misc.agentmcp.task.error = "ran out of fuel at x=2,y=73,z=-8"
```

This allows a new agent session to detect an interrupted task and resume or recover rather than starting blind.

---

## Implementation Status

| Item | Status | Notes |
|------|--------|-------|
| Standardized `agentmcp` schema | **Done** — convention established | Documented in agent-guide.md |
| `agentmcp.*` Lua helpers on turtle | **Done** — `static/agent.lua` | Downloaded on startup, available in every run_command |
| `moveTo` / `goTo` navigation | **Done** — `static/agent.lua` | Y-first, auto-digs, maxSteps guard |
| `scanSurroundings` + explored map | **Done** — `static/agent.lua` | Records notable blocks per chunk |
| `set_waypoint` / `get_waypoints` MCP tools | Not started | Would shortcut run_command round-trip |
| `list_area` MCP tool | Not started | Would expose explored map via get_turtle alternative |
| Command elapsed time in response | Not started | Low effort, helps agents self-regulate loop size |
| Turtle-side firmware `scan_area(radius)` | Not started | Needs a full walk pattern inside Lua firmware |
