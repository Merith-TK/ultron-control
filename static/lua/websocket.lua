local turtlePost = {
	name = "",
	id = 0,
	pos = {
		x = 0,
		y = 0,
		z = 0,
		r = 0
	},
	fuel = 0,
	maxFuel = 0,
	selectedSlot = 0,
	inventory = {
	},
	cmdResult = nil,
	cmdQueue = {},
}


-- function to recieve commands from websocket

_G.skyrtle = require("skyrtle")
local init = require("init")


local wsHeader = { "data", tostring(os.getComputerID()), } -- header for websocket
local ws, wsErr = http.websocket(init.config.ws, wsHeader )
while not ws do
	print("Error: Unable to connect to websocket, trying again")
	sleep(1)
	ws, wsErr = http.websocket(init.config.url, wsHeader )
end
print(init.config.ws)

-- function to send turtle data to websocket
local function updateControl()
	turtlePost.id = os.getComputerID()
	local label = os.getComputerLabel()
	if label and not label == "" then
		turtlePost.name = label
	else
		os.setComputerLabel(tostring(turtlePost.id))
		turtlePost.name = tostring(turtlePost.id)
	end

	local x,y,z = skyrtle.getPosition()
	local r = skyrtle.getFacing()
	turtlePost.pos.x = x
	turtlePost.pos.y = y
	turtlePost.pos.z = z
	turtlePost.pos.r = r

	turtlePost.fuel = turtle.getFuelLevel()
	turtlePost.maxFuel = turtle.getFuelLimit()

	turtlePost.selectedSlot = turtle.getSelectedSlot()

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
		turtlePost.inventory[i] = item
	end
	turtle.select(turtlePost.selectedSlot)

	local TurtleData =  textutils.serializeJSON(turtlePost)
	ws.send(TurtleData)
end


local function recieveOrders()
	local message, binary = ws.receive()
	if message then
		-- if message is a command then add to turtlePost.cmdQueue
		if message:sub(1,1) == "{" then
			local cmd = textutils.unserialize(message)
			if cmd.cmd then
				table.insert(turtlePost.cmdQueue, cmd)
			end
		end
	end
end

-- process cmdQueue as function
local function processCmdQueue()
	if #turtlePost.cmdQueue > 0 then
		local cmd = table.remove(turtlePost.cmdQueue, 1)
		if cmd then
			if init.config.debug then
				print("[Debug] Processing: " .. cmd)
			end
			local cmd, err = loadstring(cmd)
			if cmd then
				setfenv(cmd, getfenv())
				
				--append result to cmdResult
				local result = cmd()
				if result then
					turtlePost.cmdResult = result
				else
					turtlePost.cmdResult = err
				end
				
				--pcall(cmd)
				if init.config.debug then
					print("[Debug] Result: " .. tostring(turtlePost.cmdResult))
				end
			else
				print("error in loadstring(" .. cmd .. ")");
				turtlePost.cmdResult = err
			end
		end
	end
end

local function waitForDelay()
	sleep(init.config.apiDelay)
end


local function main()
	while true do
		parallel.waitForAny(recieveOrders, processCmdQueue, waitForDelay)
		updateControl()
	end
end

main()
ws.close()