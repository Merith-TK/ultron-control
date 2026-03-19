# Ultron Control — Agent Guide

**READ THIS BEFORE TOUCHING ANYTHING.**

---

## What Is This

Ultron Control is a CC:Tweaked (ComputerCraft) turtle management system embedded in Minecraft.
Turtles are programmable robots that can move, dig, place blocks, craft items, manage inventory, and more.

This MCP server gives you real-time control over connected turtles via Lua code execution.

---

## Tools Available

| Tool | Description |
|------|-------------|
| `get_instructions` | Returns this guide. Call it first. |
| `list_turtles` | All connected turtles with full state (pos, fuel, inventory, sight) |
| `get_turtle(id)` | Single turtle state — use this to read position, don't query it via run_command |
| `run_command(id, command)` | Execute Lua on a turtle, synchronously, 30s timeout |
| `list_docs` | List available documentation sets |
| `get_doc(name, file)` | Read a doc file — use this for full CC:Tweaked API reference |

---

## run_command: Critical Rules

`run_command` executes any valid Lua chunk on the turtle. It is NOT a one-liner REPL.

### You MUST use `return` to get data back

```lua
-- WRONG: returns nothing
turtle.getFuelLevel()

-- RIGHT
return turtle.getFuelLevel()
```

### You CAN write full multi-line scripts

```lua
local x, y, z = skyrtle.getPosition()
local r, rname = skyrtle.getFacing()
local fuel = turtle.getFuelLevel()
return "pos=" .. x .. "," .. y .. "," .. z .. " facing=" .. rname .. " fuel=" .. fuel
```

### Multi-return functions WILL bite you

CC:Tweaked functions like `skyrtle.getPosition()` return multiple values.
`run_command` only captures the first return value. Always unpack manually:

```lua
-- WRONG: only gets x
return skyrtle.getPosition()

-- WRONG: textutils.serializeJSON chokes on integer-keyed tables
return {skyrtle.getPosition()}

-- RIGHT
local x, y, z = skyrtle.getPosition()
return x .. "," .. y .. "," .. z
```

### 30-second timeout is real

Each `run_command` call has a hard 30s timeout. For long operations (deep mining, long travel), split into multiple calls. A movement loop covering 100 blocks will timeout.

**Safe loop sizing:** Keep loops to **5–8 move+scan steps** per command. 15-step loops with per-step inspection reliably time out. Break long travel into multiple `run_command` calls.

---

## Reading Turtle State

**Use `get_turtle(id)` — it's free, always fresh, and has everything:**

```json
{
  "pos": { "x": 12, "y": 72, "z": -15, "r": 3, "rname": "west" },
  "fuel": { "current": 20460, "max": 100000 },
  "inventory": [ ... 16 slots ... ],
  "sight": { "up": {}, "down": { "name": "minecraft:stone" }, "front": {} }
}
```

- `pos.r`: 0=north, 1=east, 2=south, 3=west
- `sight.*`: empty `{}` means air/nothing; populated means a block is there with its data
- `inventory`: 16 slots (1-indexed in Lua, 0-indexed in JSON array). Slot 0 in JSON = slot 1 in Lua.

---

## Position Tracking

Turtles use `skyrtle` — a dead reckoning library that intercepts `turtle.forward()`, `turtle.turnLeft()`, etc. and tracks position internally. It syncs from GPS when available.

Position is automatically updated in every heartbeat. Just call `get_turtle`.

---

## Crafting

Turtle inventory is a **4×4 grid** (16 slots). The crafting area is the **top-left 3×3**:

```
[ 1][ 2][ 3][ 4]
[ 5][ 6][ 7][ 8]
[ 9][10][11][12]
[13][14][15][16]
```

**CRITICAL RULE: ALL 16 slots must contain ONLY recipe items or be empty.**
Any item in any slot that is not part of the current recipe will cause `No matching recipes`.
This includes slots 13–16. There is NO safe "extra item" slot.

**Rules:**
1. Arrange recipe items in the correct slots (e.g. slots 1,2,3,5,6,7,9,10,11 for a 3×3 recipe)
2. Every other slot must be completely empty
3. Call `turtle.craft()`

**Technique — storing items while crafting:**
Since no extra items can be in inventory, store them outside the turtle temporarily:
- `turtle.place()` / `turtle.placeUp()` — place a block in the world, dig it back after
- Only works for block-placeable items (logs, planks, cobblestone, etc.)
- After crafting, use `turtle.select(target_slot)` BEFORE `turtle.dig()` — the mined item goes into the selected slot if it's empty. Use this to place items directly into specific recipe slots.
- **WARNING:** A single plank in the crafting grid crafts into a spruce button, not planks. Don't accidentally call `craft()` with planks in inventory unless you intend the chest recipe.
- **WARNING:** When placing blocks for temp storage, always leave at least one direction open. Placing blocks in all surrounding directions traps the turtle.
- **DO NOT use dropDown/suckDown for temp item staging.** Dropped items on grass are unreliable to retrieve — they may fall to an unreachable position or not be picked up by `suckDown`. Use place-as-block instead.

**Multi-step crafting (e.g. 2 logs → 8 planks → chest):**
```
1. Place log2 as block in a known direction (e.g. south)
2. Craft log1 → 4 planks [slot 1 only, all else empty]
3. Place those 4 planks as blocks in 4 known directions (up, north, east, west)
4. Dig log2 back with select(1) → log2 in slot 1
5. Craft log2 → 4 planks [slot 1 only, all else empty]
6. Distribute 4 planks to chest recipe slots 2,3,5,7 via transferTo
7. Dig the 4 plank blocks back using select(target) before each dig:
   select(1) + digUp()  → plank in slot 1
   select(9) + dig north → plank in slot 9
   select(10) + dig east → plank in slot 10
   select(11) + dig west → plank in slot 11
8. craft() → chest ✓
```

**Example — Planks from 1 log (slot 1 only, all others empty):**
```
[LOG][  ][  ][  ]
[  ][  ][  ][  ]
[  ][  ][  ][  ]
[  ][  ][  ][  ]
```

**Example — Furnace (8 cobblestone ring, empty center):**
```
[CB][CB][CB][  ]
[CB][  ][CB][  ]
[CB][CB][CB][  ]
[  ][  ][  ][  ]
```
Slots 1,2,3,5,7,9,10,11 = cobblestone. Slot 6 = empty. Slots 4,8,12,13-16 = empty.

---

## Agent Helper Library

The turtle firmware includes `agent.lua`, downloaded on startup. It provides high-level helpers available in every `run_command` call — no setup required.

### State management (`agentmcp.*`)

```lua
-- Waypoints
agentmcp.setWaypoint("home")          -- save current pos as "home"
agentmcp.getWaypoints()               -- JSON of all saved waypoints
agentmcp.getWaypoint("home")          -- JSON of one waypoint

-- Task tracking
agentmcp.setTask("mining", "init")    -- set current task + phase
agentmcp.updateTask("descending")     -- update phase
agentmcp.updateTask("failed", {error="out of fuel at 4,65,-8"})
agentmcp.clearTask()                  -- mark task complete

-- Area / explored map
agentmcp.recordExplored({"minecraft:spruce_log", "minecraft:coal_ore"})
agentmcp.getExplored()                -- JSON of all explored chunks
agentmcp.findResource("spruce_log")   -- JSON of chunks containing that resource

-- Scanning (inspects all sides, filters boring background blocks)
agentmcp.scanSurroundings()           -- returns JSON list of notable blocks
agentmcp.scanSurroundings(true)       -- same, and records into explored map
```

### Navigation

```lua
-- Move toward absolute coordinates (up to maxSteps per call)
moveTo(tx, ty, tz)           -- default 16 steps
moveTo(tx, ty, tz, 8)        -- limit to 8 steps

-- Move to a named waypoint
goTo("home")
goTo("mine_entrance", 8)
```

`moveTo` / `goTo` return values:
- `"done"` — arrived at target
- `"partial x,y,z"` — maxSteps used up, call again to continue
- `"blocked:axis:err"` — hit bedrock, world border, or ran out of fuel

Movement order: Y (vertical) first, then X, then Z. Digs diggable blocks automatically. Use maxSteps ≤ 16 to stay inside the 30s timeout.

**Pattern — travel with resume:**
```lua
-- Call repeatedly until done or blocked
local result = moveTo(2, 73, -12)
return result  -- "done", "partial ...", or "blocked:..."
```

---

## Persistent State — ultron.data.misc

Turtles expose an arbitrary Lua table via `ultron.data.misc`. This is sent in every heartbeat and visible in `get_turtle(id)` under the `misc` field.

All agent data lives under `ultron.data.misc.agentmcp` — never write directly to the misc root, as other systems may use it.

**Writing to misc (persists across commands, sent in heartbeat):**
```lua
-- Initialize namespace if needed
if not ultron.data.misc.agentmcp then
    ultron.data.misc.agentmcp = {}
end

-- Store a waypoint
ultron.data.misc.agentmcp.waypoints = ultron.data.misc.agentmcp.waypoints or {}
ultron.data.misc.agentmcp.waypoints.home = {x=7, y=73, z=-12, r=0}

-- Record current task
ultron.data.misc.agentmcp.task = { name="mining", phase="descending" }
return "saved"
```

**Reading from misc:**
```lua
local a = ultron.data.misc.agentmcp
if a and a.waypoints and a.waypoints.home then
    local h = a.waypoints.home
    return h.x .. "," .. h.y .. "," .. h.z
else
    return "no waypoint set"
end
```

**Read via get_turtle:** After a heartbeat, `misc` appears in the JSON state under `misc.agentmcp`. No need to run a command to read it.

**Use cases:**
- Storing home/base waypoints (`agentmcp.waypoints.home`)
- Tracking current task phase (`agentmcp.task`)
- Caching explored chunks and resource locations found in them (`agentmcp.explored`)
- Storing a "return path" before a long trip

---

## Fuel

- Check: `turtle.getFuelLevel()` and `turtle.getFuelLimit()`
- Refuel: `turtle.select(slot); turtle.refuel()` — consumes ALL items in that slot
- Coal = 80 fuel per item, coal blocks = 800
- A turtle at 0 fuel cannot move

---

## Rebooting

- Sending `reboot()` will ALWAYS timeout — the connection dies before a response can come back. That's expected, not an error.
- After reboot, wait for the turtle to reconnect (watch `list_turtles` heartbeat update)
- Lua file changes (ultron.lua, turtle.lua, etc.) require a reboot to take effect. The turtle auto-downloads updated files on startup.

---

## CC:Tweaked API Quick Reference

```lua
-- Movement (each returns true/false, optional error string)
turtle.forward()  turtle.back()
turtle.up()       turtle.down()
turtle.turnLeft() turtle.turnRight()

-- Digging
turtle.dig()      turtle.digUp()    turtle.digDown()

-- Placing
turtle.place()    turtle.placeUp()  turtle.placeDown()

-- Inspection (returns: bool, table|string)
turtle.detect()     turtle.detectUp()     turtle.detectDown()
turtle.inspect()    turtle.inspectUp()    turtle.inspectDown()

-- Inventory
turtle.select(slot)               -- slots are 1-16
turtle.getItemDetail(slot, true)  -- true = detailed NBT info
turtle.getItemCount(slot)
turtle.transferTo(slot, count)
turtle.drop(count)   turtle.dropUp()   turtle.dropDown()
turtle.suck(count)   turtle.suckUp()   turtle.suckDown()

-- Fuel
turtle.getFuelLevel()
turtle.getFuelLimit()
turtle.refuel(count)  -- count optional, omit to use whole stack

-- Crafting
turtle.craft(count)  -- returns true/false, error string

-- skyrtle (position tracking, injected into turtle.*)
skyrtle.getPosition()  -- returns x, y, z
skyrtle.getFacing()    -- returns r (int), rname (string)
```

For full documentation: `get_doc(name="computercraft", file="...")`
Start with `list_docs` to see what's available, then explore from there.

---

## Common Patterns

**Scan all 6 sides:**
```lua
local result = {}
result.front = {turtle.inspect()}
result.up    = {turtle.inspectUp()}
result.down  = {turtle.inspectDown()}
turtle.turnLeft()
result.left  = {turtle.inspect()}
turtle.turnRight() turtle.turnRight()
result.right = {turtle.inspect()}
turtle.turnLeft()
-- Note: no back inspection without moving
return textutils.serializeJSON(result)
```

**Mine and collect:**
```lua
local collected = 0
while turtle.detect() do
    turtle.dig()
    turtle.forward()
    collected = collected + 1
    if collected >= 10 then break end
end
return collected
```

**Safe movement (check before move):**
```lua
if turtle.detect() then
    turtle.dig()
end
local ok, err = turtle.forward()
return tostring(ok) .. " " .. tostring(err)
```

---

## Movement Failures

`turtle.forward()` (and `back`, `up`, `down`) return `false, errorString` when blocked. Four causes:

| Error | Cause |
|-------|-------|
| block in the way | Dig first, then move |
| unbreakable block | Bedrock — choose a different path |
| no block but still blocked | World border — turn around |
| cannot move | Out of fuel — refuel before moving |

Always check the return value and handle accordingly:
```lua
local ok, err = turtle.forward()
if not ok then
    if turtle.detect() then
        turtle.dig()
        turtle.forward()
    else
        -- world border or out of fuel
        return "blocked: " .. tostring(err)
    end
end
```
