-- agent.lua — MCP Agent Helper Library
-- Downloaded to turtle on startup. Provides high-level helpers that remote
-- agents can call via run_command without knowing the firmware internals.
--
-- All persistent state lives under ultron.data.misc.agentmcp so it is
-- serialized in every heartbeat and visible in get_turtle output.
--
-- Available globals after load:
--   agentmcp.*   — state management (waypoints, tasks, explored map)
--   moveTo(x,y,z,maxSteps) — navigate toward coordinates
--   goTo(name, maxSteps)   — navigate to named waypoint

_G.agentmcp = {}

-- Ensure agentmcp namespace exists in misc
local function init()
    if not ultron.data.misc.agentmcp then
        ultron.data.misc.agentmcp = {}
    end
    local a = ultron.data.misc.agentmcp
    if not a.waypoints then a.waypoints = {} end
    if not a.explored  then a.explored  = {} end
end

-- Chunk key for a given x,z position (16x16 grid)
local function chunkKey(x, z)
    return math.floor(x / 16) .. "," .. math.floor(z / 16)
end

------------------------------------------------------------------------
-- Waypoints
------------------------------------------------------------------------

-- Save current turtle position as a named waypoint.
-- Returns a confirmation string.
function agentmcp.setWaypoint(name)
    init()
    local x, y, z = skyrtle.getPosition()
    local r, _ = skyrtle.getFacing()
    ultron.data.misc.agentmcp.waypoints[name] = { x = x, y = y, z = z, r = r }
    return "waypoint '" .. name .. "' saved at " .. x .. "," .. y .. "," .. z
end

-- Get all saved waypoints as a JSON string.
function agentmcp.getWaypoints()
    init()
    return textutils.serializeJSON(ultron.data.misc.agentmcp.waypoints)
end

-- Get a single waypoint table, or nil.
function agentmcp.getWaypoint(name)
    init()
    local wp = ultron.data.misc.agentmcp.waypoints[name]
    if wp then return textutils.serializeJSON(wp) end
    return nil
end

------------------------------------------------------------------------
-- Task tracking
------------------------------------------------------------------------

-- Set the current task. Overwrites any previous task.
function agentmcp.setTask(name, phase)
    init()
    ultron.data.misc.agentmcp.task = { name = name, phase = phase or "init" }
    return "task: " .. name .. " [" .. (phase or "init") .. "]"
end

-- Update the phase of the current task. Pass extra={key=val} for
-- additional fields (e.g. error message, target coords).
function agentmcp.updateTask(phase, extra)
    init()
    local t = ultron.data.misc.agentmcp.task
    if not t then
        ultron.data.misc.agentmcp.task = { name = "unknown", phase = phase }
    else
        t.phase = phase
        if extra then
            for k, v in pairs(extra) do t[k] = v end
        end
    end
    return "phase -> " .. phase
end

-- Clear the current task (call on clean completion).
function agentmcp.clearTask()
    init()
    ultron.data.misc.agentmcp.task = nil
    return "task cleared"
end

------------------------------------------------------------------------
-- Area / explored map
------------------------------------------------------------------------

-- Mark the current chunk as explored with a list of notable resource
-- item IDs (e.g. {"minecraft:spruce_log", "minecraft:coal_ore"}).
-- Call with no args to mark as explored-but-empty.
function agentmcp.recordExplored(resources)
    init()
    local x, y, z = skyrtle.getPosition()
    local key = chunkKey(x, z)
    ultron.data.misc.agentmcp.explored[key] = {
        scanned_at = os.epoch("utc"),
        resources  = resources or {},
    }
    return "chunk " .. key .. " recorded (" .. #(resources or {}) .. " resources)"
end

-- Get the full explored map as JSON.
function agentmcp.getExplored()
    init()
    return textutils.serializeJSON(ultron.data.misc.agentmcp.explored)
end

-- Find all explored chunks that contain a resource matching itemName
-- (substring match, so "spruce_log" matches "minecraft:spruce_log").
-- Returns JSON of matching chunk keys → chunk data.
function agentmcp.findResource(itemName)
    init()
    local results = {}
    for key, data in pairs(ultron.data.misc.agentmcp.explored) do
        if data.resources then
            for _, res in ipairs(data.resources) do
                if res:find(itemName, 1, true) then
                    results[key] = data
                    break
                end
            end
        end
    end
    return textutils.serializeJSON(results)
end

------------------------------------------------------------------------
-- Handler registry
-- Handlers are called when a block is inspected, to collect extra data
-- beyond what inspect() provides (e.g. peripheral inventory, fluid level).
--
-- Usage:
--   agentmcp.registerHandler("myhandler", "chest|barrel", function(wrapDir, name, state)
--       local p = peripheral.wrap(wrapDir)
--       if p and p.list then return p.list() end
--   end)
--
-- wrapDir is the peripheral.wrap() direction string:
--   "top", "bottom", "front" (also used after turning for left/back/right)
------------------------------------------------------------------------

local _handlers = {}

-- Register a new handler. namePattern is matched against block names with
-- string.find (plain substring match). fn returns a table or nil.
function agentmcp.registerHandler(name, namePattern, fn)
    _handlers[#_handlers + 1] = { name = name, pattern = namePattern, fn = fn }
end

-- Run all registered handlers against a block. Returns merged extra table,
-- or nil if no handler produced data.
function agentmcp.runHandlers(wrapDir, blockName, blockState)
    if not blockName then return nil end
    local extra = {}
    local hasData = false
    for _, h in ipairs(_handlers) do
        if blockName:find(h.pattern, 1, true) then
            local ok, result = pcall(h.fn, wrapDir, blockName, blockState)
            if ok and result ~= nil then
                extra[h.name] = result
                hasData = true
            end
        end
    end
    return hasData and extra or nil
end

-- Built-in: inventory handler for storage blocks.
-- Reads item list via peripheral API when adjacent to a chest, barrel, etc.
agentmcp.registerHandler("inventory",
    "chest|barrel|hopper|shulker|dropper|dispenser|furnace",
    function(wrapDir, blockName, blockState)
        local p = peripheral.wrap(wrapDir)
        if not p or not p.list then return nil end
        local raw = p.list()
        local result = {}
        for slot, item in pairs(raw) do
            result[#result + 1] = { slot = slot, name = item.name, count = item.count }
        end
        return #result > 0 and result or nil
    end
)

-- Full inspection: inspect all available sides, run handlers on each block found.
-- Returns a table of side → {name, state, extra}.
-- Call agentmcp.fullInspect() via run_command to get rich block data with
-- handler results included (e.g. chest inventories).
function agentmcp.fullInspect()
    local function side(inspectFn, wrapDir)
        local ok, data = inspectFn()
        if not ok then return {} end
        local result = { name = data.name, state = data.state }
        local extra = agentmcp.runHandlers(wrapDir, data.name, data.state)
        if extra then result.extra = extra end
        return result
    end

    local sides = {}
    sides.up    = side(turtle.inspectUp,   "top")
    sides.down  = side(turtle.inspectDown, "bottom")
    sides.front = side(turtle.inspect,     "front")

    if fs.exists("cfg/inspectAll") then
        turtle.turnLeft()
        sides.left  = side(turtle.inspect, "front") -- peripheral dir is always "front" after turning
        turtle.turnLeft()
        sides.back  = side(turtle.inspect, "front")
        turtle.turnLeft()
        sides.right = side(turtle.inspect, "front")
        turtle.turnLeft() -- restore
    end

    return textutils.serializeJSON(sides)
end

------------------------------------------------------------------------
-- Scanning
------------------------------------------------------------------------

-- Inspect all four horizontal sides plus up/down. Returns a list of
-- notable block names (filters out stone, dirt, grass, deepslate,
-- gravel, sand — the boring background noise).
-- If record=true, saves the findings into the current chunk entry.
-- Restores original facing after scan.
function agentmcp.scanSurroundings(record)
    local boring = {
        ["minecraft:stone"]       = true,
        ["minecraft:dirt"]        = true,
        ["minecraft:grass_block"] = true,
        ["minecraft:deepslate"]   = true,
        ["minecraft:cobblestone"] = true,
        ["minecraft:gravel"]      = true,
        ["minecraft:sand"]        = true,
        ["minecraft:air"]         = true,
    }

    local found = {}
    local seen  = {}

    local function check(inspectFn)
        local ok, data = inspectFn()
        if ok and data and data.name and not boring[data.name] and not seen[data.name] then
            found[#found + 1] = data.name
            seen[data.name] = true
        end
    end

    -- up / down (no turn needed)
    check(turtle.inspectUp)
    check(turtle.inspectDown)
    -- front, left, right, then restore
    check(turtle.inspect)
    turtle.turnLeft()
    check(turtle.inspect)
    turtle.turnRight()
    turtle.turnRight()
    check(turtle.inspect)
    turtle.turnLeft() -- back to original

    if record then
        agentmcp.recordExplored(found)
    end
    return textutils.serializeJSON(found)
end

------------------------------------------------------------------------
-- Navigation
------------------------------------------------------------------------

-- Turn to face a target direction (0=north,1=east,2=south,3=west).
local function faceDir(target)
    local r, _ = skyrtle.getFacing()
    local diff = (target - r) % 4
    if diff == 1 then
        turtle.turnRight()
    elseif diff == 2 then
        turtle.turnRight()
        turtle.turnRight()
    elseif diff == 3 then
        turtle.turnLeft()
    end
    -- diff == 0: already facing correct direction
end

-- Move forward one step, digging if a diggable block is in the way.
local function stepForward()
    if turtle.detect() then turtle.dig() end
    return turtle.forward()
end

local function stepUp()
    if turtle.detectUp() then turtle.digUp() end
    return turtle.up()
end

local function stepDown()
    if turtle.detectDown() then turtle.digDown() end
    return turtle.down()
end

-- Navigate toward (tx, ty, tz), taking up to maxSteps move steps.
-- Movement order: Y (vertical) first, then X, then Z.
-- Returns:
--   "done"               — arrived at target
--   "partial x,y,z"      — maxSteps exhausted, call again to continue
--   "blocked:axis:err"   — movement failed (bedrock, world border, fuel)
--
-- Keep maxSteps at 16 or fewer to avoid the 30s run_command timeout.
-- WARNING: moveTo digs through any block in the path, including placed
-- blocks like furnaces. Use moveToSafe when near placed blocks.
function moveTo(tx, ty, tz, maxSteps)
    maxSteps = maxSteps or 16
    local steps = 0

    while steps < maxSteps do
        local x, y, z = skyrtle.getPosition()
        if x == tx and y == ty and z == tz then return "done" end

        local ok, err
        if y < ty then
            ok, err = stepUp()
            if not ok then return "blocked:up:" .. tostring(err) end
        elseif y > ty then
            ok, err = stepDown()
            if not ok then return "blocked:down:" .. tostring(err) end
        elseif x < tx then
            faceDir(1) -- east  (x+)
            ok, err = stepForward()
            if not ok then return "blocked:east:" .. tostring(err) end
        elseif x > tx then
            faceDir(3) -- west  (x-)
            ok, err = stepForward()
            if not ok then return "blocked:west:" .. tostring(err) end
        elseif z < tz then
            faceDir(2) -- south (z+)
            ok, err = stepForward()
            if not ok then return "blocked:south:" .. tostring(err) end
        elseif z > tz then
            faceDir(0) -- north (z-)
            ok, err = stepForward()
            if not ok then return "blocked:north:" .. tostring(err) end
        end

        steps = steps + 1
    end

    local x, y, z = skyrtle.getPosition()
    return "partial " .. x .. "," .. y .. "," .. z
end

-- Navigate to a named waypoint. Same return values as moveTo.
function goTo(name, maxSteps)
    init()
    local wp = ultron.data.misc.agentmcp.waypoints[name]
    if not wp then return "waypoint '" .. tostring(name) .. "' not found" end
    return moveTo(wp.x, wp.y, wp.z, maxSteps)
end

-- moveToSafe: like moveTo but NEVER digs. Returns "blocked:*" if any block
-- is in the way. Use near placed blocks (furnaces, chests, etc.).
function moveToSafe(tx, ty, tz, maxSteps)
    maxSteps = maxSteps or 16
    local steps = 0
    while steps < maxSteps do
        local x, y, z = skyrtle.getPosition()
        if x == tx and y == ty and z == tz then return "done" end
        local ok, err
        if y < ty then
            if turtle.detectUp() then return "blocked:up:block" end
            ok, err = turtle.up()
            if not ok then return "blocked:up:" .. tostring(err) end
        elseif y > ty then
            if turtle.detectDown() then return "blocked:down:block" end
            ok, err = turtle.down()
            if not ok then return "blocked:down:" .. tostring(err) end
        elseif x < tx then
            faceDir(1)
            if turtle.detect() then return "blocked:east:block" end
            ok, err = turtle.forward()
            if not ok then return "blocked:east:" .. tostring(err) end
        elseif x > tx then
            faceDir(3)
            if turtle.detect() then return "blocked:west:block" end
            ok, err = turtle.forward()
            if not ok then return "blocked:west:" .. tostring(err) end
        elseif z < tz then
            faceDir(2)
            if turtle.detect() then return "blocked:south:block" end
            ok, err = turtle.forward()
            if not ok then return "blocked:south:" .. tostring(err) end
        elseif z > tz then
            faceDir(0)
            if turtle.detect() then return "blocked:north:block" end
            ok, err = turtle.forward()
            if not ok then return "blocked:north:" .. tostring(err) end
        end
        steps = steps + 1
    end
    local x, y, z = skyrtle.getPosition()
    return "partial " .. x .. "," .. y .. "," .. z
end

return agentmcp
