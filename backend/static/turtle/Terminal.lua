local turtleData = {
	name = "",
	id = 0,
	pos = {
		x = 0,
		y = 0,
		z = 0,
		r = 0,
		rname = "",
	},
	fuel = 0,
	maxFuel = 0,
	selectedSlot = 0,
	inventory = {
	},
	cmdResult = nil,
	miscData = {},
}


local cmdQueue = {}

-- function to recieve commands from websocket

_G.skyrtle = require("skyrtle")

skyrtle.hijack()
local init = require("../init")
local config = init.config



local wsHeader = { "data", tostring(os.getComputerID()), } -- header for websocket


init.debugPrint("ApiDelay: " .. config.apiDelay)
init.debugPrint("Websocket URL: " .. config.ws.turtle)
init.debugPrint("Websocket Header: " .. textutils.serialize(wsHeader))


local ws, wsErr = http.websocket(config.ws.turtle, wsHeader )
while not ws and not debug do
	print("Error: Unable to connect to websocket, trying again")
	sleep(1)
	ws, wsErr = http.websocket(config.ws.turtle, wsHeader )
end


-- function to send turtle data to websocket
local function updateControl()
	turtleData.id = os.getComputerID()
	local label = os.getComputerLabel()
	if label and not label == "" then
		turtleData.name = label
	else
		os.setComputerLabel(tostring(turtleData.id))
		turtleData.name = tostring(turtleData.id)
	end

	local x,y,z = skyrtle.getPosition()
	local r, rname = skyrtle.getFacing()
	turtleData.pos.x = x
	turtleData.pos.y = y
	turtleData.pos.z = z
	turtleData.pos.r = r
	turtleData.pos.rname = rname

	turtleData.fuel = turtle.getFuelLevel()
	turtleData.maxFuel = turtle.getFuelLimit()

	turtleData.selectedSlot = turtle.getSelectedSlot()

	for i = 1, 16 do
		turtle.select(i)
		local item = turtle.getItemDetail(i, true)
		if not item then
			item = {
				name = "",
				count = 0,
				slot = i,
			}
		end
		turtleData.inventory[i] = item
	end
	turtle.select(turtleData.selectedSlot)

	local TurtleData =  textutils.serializeJSON(turtleData)
	ws.send(TurtleData)
end

-- process cmdQueue as function
local function processCmdQueue()
	if #cmdQueue > 0 then
		local cmd = table.remove(cmdQueue, 1)
		if cmd then
			init.debugPrint("Processing command: " .. cmd)
			local result = skyrtle.processCmd(cmd)
			turtleData.cmdResult = result
			init.debugPrint("Command result: " .. result)
		end
	end
end

local function recieveOrders()
	while true do
		local data, err = ws.receive()
		if data then
			init.debugPrint("Order Recieved: " .. data)
			if not data == "null" then
				-- add data to cmdQueue
				table.insert(cmdQueue, data)
			end
		end
		processCmdQueue()
	end
end


local function waitForDelay()
	sleep(config.apiDelay)
end


local function main()
	while true do
		parallel.waitForAny(waitForDelay, recieveOrders)
		updateControl()
	end
end

main()
ws.close()