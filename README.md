Here's the refined documentation incorporating all the detailed response format and additional endpoint information:

---

# Ultron Control
**Complete API Documentation for ComputerCraft Turtles**

## Response Format (GET `/api/turtle/<id>`)
```json
{
  "name": "TurtleName",
  "id": 0,
  "inventory": [
    {
      "count": 52,
      "displayName": "Cobblestone",
      "name": "minecraft:cobblestone",
      "maxCount": 64,
      "tags": {
        "c:cobblestones": true,
        "minecraft:stone_tool_materials": true
      },
      "itemGroups": [
        {
          "displayName": "Building Blocks",
          "id": "minecraft:building_blocks"
        }
      ]
    },
    /* ...15 more slots... */
  ],
  "selectedSlot": 2,
  "pos": {
    "x": -55,
    "y": 87,
    "z": -127,
    "r": 0,
    "rname": "north"
  },
  "fuel": {
    "current": 31549,
    "max": 100000
  },
  "sight": {
    "up": { /* turtle.inspectUp() result */ },
    "down": { /* turtle.inspectDown() result */ },
    "front": { /* turtle.inspect() result */ }
  },
  "cmdResult": [true, "optional return data"],
  "cmdQueue": [],
  "misc": { /* custom data storage */ },
  "heartbeat": 1743842934360
}
```

### Field Details
| Field | Description |
|-------|-------------|
| `heartbeat` | Last communication timestamp (UTC milliseconds) |
| `misc` | Custom data storage (modify via `ultron.data.misc` in Lua) |
| `cmdQueue` | Pending command queue (FIFO) |
| `cmdResult` | [0]: bool success, [1+]: returned Lua values |
| `sight` | Results of `inspect()`, `inspectUp()`, `inspectDown()` |
| `fuel` | `current`/`max` fuel levels |
| `pos` | `x,y,z` + `r` (0-3) and `rname` (north/east/south/west) |
| `id` | ComputerCraft's turtle ID (matches API endpoint) |
| `name` | Custom turtle label |
| `inventory` | 16 slots (empty slots shown as `{}`), arranged as shown below

```m
[1] [2] [3] [4]
[5] [6] [7] [8]
[9] [10][11][12]
[13][14][15][16]
```

---

## Core Endpoints

### 1. Get All Turtles
**`GET /api/turtles`**
*For discovering turtle IDs - use sparingly*

*Example response is minified, expect full turtle data tables*
Response:
```json
[
  {"id": 0, "name": "Miner1", "pos": {"x": 10, "y": 64, "z": -30}},
  {"id": 1, "name": "Builder", "pos": {"x": 5, "y": 70, "z": 12}}
]
```

### 2. Get Turtle Status
**`GET /api/turtle/<id>`**
Returns full status snapshot (see response format above)

### 3. Send Commands
**`POST /api/turtle/<id>`**
**Content-Type:** `text/plain`

Executes Lua code on the turtle. Code runs in global environment with access to:
- All `turtle.*` API functions
- `ultron.data.misc` for persistent storage
- Return values populate `cmdResult`

#### Example Commands:
```lua
-- Basic movement with error handling
if not turtle.forward() then
  turtle.dig()
  return "Had to dig through obstacle"
end

-- Inventory management
turtle.select(1)
local success, data = turtle.inspectDown()
return success, data

-- Persistent data storage
ultron.data.misc.last_action = "Mined at "..os.time()
```

#### Command Tips:
1. Chain actions with error checking:
   ```lua
   turtle.turnRight() or turtle.attack()
   ```
2. Return inspection data:
   ```lua
   return turtle.inspect()
   ```
3. Store state between commands:
   ```lua
   ultron.data.misc.counter = (ultron.data.misc.counter or 0) + 1
   ```

---

## Practical Usage Guide

### Monitoring Workflow
1. Poll `/api/turtle/<id>` every 1-5 seconds
2. Check `cmdResult` for command feedback
3. Monitor `fuel.current` for refuel needs
4. Use `sight` data for obstacle detection

### Control Patterns
**Basic Mining:**
```lua
while turtle.dig() do
  turtle.forward()
  if turtle.detectDown() then
    turtle.digDown()
  end
end
return "Reached obstacle"
```

**Item Sorting:**
```lua
for slot=1,16 do
  turtle.select(slot)
  local item = turtle.getItemDetail()
  if item and item.name:find("ore") then
    turtle.drop()
  end
end
return "Inventory sorted"
```

**Emergency Recall:**
```lua
while turtle.getFuelLevel() > 1000 do
  if not turtle.up() then
    turtle.digUp()
  end
end
return "Returned to surface"
```

---

## Best Practices
1. **Queue Management**: Limit to 3-5 commands per request
2. **Error Handling**: Always check command success
3. **Fuel Safety**: Verify `fuel.current` before long movements
4. **State Storage**: Use `ultron.data.misc` for persistent data
5. **Return Values**: Structure returns as `return success, data1, data2`

---

## Troubleshooting
| Symptom | Solution |
|---------|----------|
| No `cmdResult` | Ensure your code uses `return` |
| Empty inventory slots | Normal - slots show as `{}` |
| Stale heartbeat | Check turtle connectivity |
| Command timeout | Verify no infinite loops in code |