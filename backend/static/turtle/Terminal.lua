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
	fuel = {
		current = 0,
		max = 0,
	},
	selectedSlot = 0,
	inventory = {
	},
	cmdResult = nil,
	cmdQueue = {},
	miscData = {},
}

-- function to recieve commands from websocket

_G.skyrtle = require("skyrtle")

skyrtle.hijack()
local init = require("../init")
local config = init.config

init.debugPrint()


init.debugPrint("ApiDelay: " .. config.apiDelay)
init.debugPrint("Websocket URL: " .. config.ws.turtle)
init.debugPrint("Websocket Header: " .. textutils.serialize(wsHeader))


-- local ws, wsErr = http.websocket(config.ws.turtle, wsHeader )
-- while not ws and not debug do
-- 	print("Error: Unable to connect to websocket, trying again")
-- 	sleep(1)
-- 	ws, wsErr = http.websocket(config.ws.turtle, wsHeader )
-- end

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

	turtleData.fuel.current = turtle.getFuelLevel()
	turtleData.fuel.max = turtle.getFuelLimit()

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
	init.ws("send",TurtleData)
end

-- process cmdQueue as function
local function processCmdQueue()
	init.debugPrint("Processing cmdQueue")
	init.debugPrint("cmdQueue: " .. textutils.serialize(turtleData.cmdQueue))
	-- check if cmdQueue is empty
	if #turtleData.cmdQueue ~= 0 then
		turtleData.cmdResult = nil
		init.debugPrint("Executing cmdQueue")
		local cmd = table.remove(turtleData.cmdQueue, 1)
		if cmd then
			-- remove [" and "] from beginning and end of cmd
			cmd = cmd:sub(3, -3)

			init.debugPrint("cmd: " .. cmd)

			init.debugPrint("Processing command: " .. cmd)
			local cmdExec, err = loadstring(cmd)
			if cmdExec then
				local success, result = pcall(cmdExec)
				if success then
					--init.debugPrint("[CMD] " .. cmd .. ": " .. result)
					turtleData.cmdResult = result
				else
					--init.debugPrint("[CMD] " .. cmd .. ": " .. result)
					turtleData.cmdResult = result
				end
			else
				init.debugPrint("[CMD] " .. cmd .. ": " .. err)
				turtleData.cmdResult = nil
			end
		end
	end
end



local function recieveOrders()
	while true do
		init.debugPrint("Waiting for orders")
		local event, url, data = os.pullEvent("websocket_message")
		if data then
			init.debugPrint("Order Recieved: " .. data)
			table.insert(turtleData.cmdQueue, data)
			init.debugPrint("cmdQueue: " .. textutils.serialize(turtleData.cmdQueue))
		else
			init.debugPrint("No data recieved")
		end

		-- save cmdQueue to file /cmdQueue.json
		local file = fs.open("/cmdQueue.json", "w")
		file.write(textutils.serializeJSON(turtleData.cmdQueue))
		file.close()

		processCmdQueue()
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
	local cmdQueue = textutils.unserialize(file.readAll())
	file.close()
	if cmdQueue then
		turtleData.cmdQueue = cmdQueue
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