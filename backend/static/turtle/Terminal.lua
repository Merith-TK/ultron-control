
-- function to recieve commands from websocket

_G.skyrtle = require("skyrtle")

skyrtle.hijack()
local init = require("../init")
local config = init.config

init.ws("open")
init.debugPrint()


init.debugPrint("ApiDelay: " .. config.apiDelay)
init.debugPrint("Websocket URL: " .. config.ws.turtle)
init.debugPrint("Websocket Header: " .. textutils.serialize(wsHeader))

-- function to send turtle data to websocket
local function updateControl()
	init.currentData.id = os.getComputerID()
	local label = os.getComputerLabel()
	if label and not label == "" then
		init.currentData.name = label
	else
		os.setComputerLabel(tostring(init.currentData.id))
		init.currentData.name = tostring(init.currentData.id)
	end

	local x,y,z = skyrtle.getPosition()
	local r, rname = skyrtle.getFacing()
	init.currentData.pos.x = x
	init.currentData.pos.y = y
	init.currentData.pos.z = z
	init.currentData.pos.r = r
	init.currentData.pos.rname = rname

	init.currentData.fuel.current = turtle.getFuelLevel()
	init.currentData.fuel.max = turtle.getFuelLimit()

	init.currentData.selectedSlot = turtle.getSelectedSlot()

	for i = 1, 16 do
		local item = turtle.getItemDetail(i, true)
		if item then
			init.currentData.inventory[i] = item
		else
			init.currentData.inventory[i] = {}
		end
	end
	turtle.select(init.currentData.selectedSlot)

	local TurtleData =  textutils.serializeJSON(init.currentData)
	init.ws("send",TurtleData)
end

-- process cmdQueue as functionlocal function recieveOrders()
local function recieveOrders()
	init.currentData.cmdQueue = init.recieveOrders(init.currentData.cmdQueue)
end
local function processCmdQueue()
	local result = init.processCmdQueue(init.currentData.cmdQueue)
	if result then
		init.currentData.cmdResult = result
	end
end


local function waitForDelay()
	sleep(config.apiDelay)
end

local function event_TurtleInventory()
	os.pullEvent("turtle_inventory")
end
local function apiLoop()
	while true do
		updateControl()
		parallel.waitForAny(waitForDelay,  event_TurtleInventory)
	end
end

local function main()
	parallel.waitForAll(apiLoop, recieveOrders, processCmdQueue)
end

-- load cmdQueue from file /cmdQueue.json
local file = fs.open("/cmdQueue.json", "r")
if file then
	local cmdQueue = textutils.unserializeJSON(file.readAll())
	file.close()
	if cmdQueue then
		init.currentData.cmdQueue = cmdQueue
	end
end

while true do
	local succ, err = pcall(main)
	if not succ then
		print("[Error] " .. err)
		break
	end
end
init.ws("close")