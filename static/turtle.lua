assert(ultron)

_G.pastebin = require("pastebin")

if not fs.exists("/skyrtle.lua") then
    local localfile = fs.open("skyrtle.lua", "w")
    local dl = http.get(
                   "https://raw.githubusercontent.com/SkyTheCodeMaster/SkyDocs/main/src/main/misc/skyrtle.lua")
    if dl then
        localfile.write(dl.readAll())
    else
        print("[Error]: Unable to download " .. "skyrtle.lua")
    end
    localfile.close()
end

_G.skyrtle = require("skyrtle")
skyrtle.hijack()

ultron.config.api.ws = ultron.config.api.host:gsub("http", "ws") .. "/turtle/ws"
ultron.ws("open", ultron.config.api.ws)
ultron.debugPrint()

ultron.debugPrint("ApiDelay: " .. ultron.config.api.delay)
ultron.debugPrint("Websocket URL: " .. "\n" .. ultron.config.api.ws)
-- ultron.debugPrint("Websocket Header: " .. textutils.serialize(wsHeader))

ultron.data = {
    name = "",
    id = 0,
    pos = {x = 0, y = 0, z = 0, r = 0, rname = ""},
    fuel = {current = 0, max = 0},
    sight = {up = {}, down = {}, front = {}},
    selectedSlot = 0,
    inventory = {},
    cmdResult = {},
    misc = {},
    heartbeat = 0
}

local function inspectWorld()
    local sight = {}
    local up, upName = turtle.inspectUp()
    local down, downName = turtle.inspectDown()
    local front, frontName = turtle.inspect()
    if up then
        sight.up = upName
    else
        sight.up = {}
    end
    if down then
        sight.down = downName
    else
        sight.down = {}
    end
    if front then
        sight.front = frontName
    else
        sight.front = {}
    end
    return sight
end

-- function to send turtle data to websocket
local function updateControl()
    ultron.data.id = os.getComputerID()
    local label = os.getComputerLabel()
    if label and not label == "" then
        ultron.data.name = label
    else
        os.setComputerLabel(tostring(ultron.data.id))
        ultron.data.name = tostring(ultron.data.id)
    end

    -- TODO: Move Execution thread to this function to prevent API desync

    local x, y, z = skyrtle.getPosition()
    local r, rname = skyrtle.getFacing()
    if gps.locate() then
        x, y, z = gps.locate()
        -- TODO: set skyrtle position to gps position
    end
    ultron.data.pos.x = x
    ultron.data.pos.y = y
    ultron.data.pos.z = z
    ultron.data.pos.r = r
    ultron.data.pos.rname = rname

    ultron.data.fuel.current = turtle.getFuelLevel()
    ultron.data.fuel.max = turtle.getFuelLimit()

    ultron.data.selectedSlot = turtle.getSelectedSlot()

    for i = 1, 16 do
        local item = turtle.getItemDetail(i, true)
        if item then
            ultron.data.inventory[i] = item
        else
            ultron.data.inventory[i] = {}
        end
    end

    ultron.data.sight = inspectWorld()
    ultron.data.heartbeat = os.epoch("utc")

    turtle.select(ultron.data.selectedSlot)

    local ultronData = ultron.data
    local TurtleData = textutils.serializeJSON(ultronData)
    ultron.ws("send", TurtleData)
    if ultron.config.debug then
        local packetFile = fs.open("/lastPacket.json", "w")
        packetFile.write(TurtleData)
        packetFile.close()
    elseif fs.exists("/lastPacket.json") then
        fs.delete("/lastPacket.json")
    end
    local miscDataFile = fs.open("/miscData.json", "w")
    miscDataFile.write(textutils.serializeJSON(ultron.data.misc))
    miscDataFile.close()
end

-- process cmdQueue as functionlocal function recieveOrders()
local function processCmdQueue()
    while true do
        sleep()
        local result = ultron.processCmd(ultron.cmd)
        if result then
            ultron.data.cmdResult = textutils.unserialise(result)
        end
    end
end

local function waitForDelay() sleep(ultron.config.api.delay) end

local function event_TurtleInventory() os.pullEvent("turtle_inventory") end
local function apiLoop()
    while true do
        updateControl()
        parallel.waitForAny(waitForDelay, event_TurtleInventory)
    end
end

local function main()
    parallel.waitForAll(apiLoop, ultron.recieveOrders, processCmdQueue)
end

-- load cmdQueue from file /cmdQueue.json
local cmdQueueJson = fs.open("/cmdQueue.json", "r")
if cmdQueueJson then
    local cmdQueue = textutils.unserializeJSON(cmdQueueJson.readAll())
    cmdQueueJson.close()
    if cmdQueue then ultron.data.cmdQueue = cmdQueue end
end
-- load miscData from /miscData.json
local miscDataJson = fs.open("/miscData.json", "r")
if miscDataJson then
    local miscData = textutils.unserializeJSON(miscDataJson.readAll())
    miscDataJson.close()
    if miscData then ultron.data.misc = miscData end
end

print("------------ Ultron Turtle ------------")
print("--------------- Started ---------------")

local function startupLoop()
    if ultron.data.misc.startup then
        if ultron.data.misc.startup.pastebin then
            for _, entry in ipairs(ultron.data.misc.startup.pastebin) do
                print("running pastebin " .. entry)
                ultron.processCmd('shell.run("pastebin run ' .. entry .. '")')
            end
        end
    end
end

local function mainLoop()
    while true do
        local succ, err = pcall(main)
        if not succ then
            print("[Error] " .. err)
            break
        end
    end
end

parallel.waitForAll(mainLoop, startupLoop)
ultron.ws("close")
print("--------------- Exited  ---------------")
