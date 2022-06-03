local computerData = {
	name = "",
	id = 0,
	cmdResult = nil,
	cmdQueue = {},
	miscData = {},
}

local init = require("../init")
local config = init.config

init.ws("open")
init.debugPrint()

init.debugPrint("ApiDelay: " .. config.apiDelay)
init.debugPrint("Websocket URL: " .. config.ws.current)
init.debugPrint("Websocket Header: " .. textutils.serialize(config.wsHeader.current))

-- function to send turtle data to websocket
local function updateControl()
	while true do
		computerData.cmdResult = init.cmdResult
		computerData.cmdQueue = init.cmdQueue
		computerData.id = os.getComputerID()
		local label = os.getComputerLabel()
		if label and not label == "" then
			computerData.name = label
		else
			os.setComputerLabel(tostring(computerData.id))
			computerData.name = tostring(computerData.id)
		end

		init.ws("send",computerData)
		init.waitForDelay()
	end
end


--------------------------------------------------------------------------------
-- main loop

local function recieveOrders()
	computerData.cmdQueue = init.recieveOrders(computerData.cmdQueue)
end
local function processCmdQueue()
	computerData.cmdResult = init.processCmdQueue(computerData.cmdQueue)
end
local function apiLoop()
	while true do
		updateControl()
		init.waitForDelay()
	end
end

local function main()
	parallel.waitForAll(apiLoop, recieveOrders, processCmdQueue)

end

while true do
	local succ, err = pcall(main)
	if not succ then
		print("[Error] " .. err)
		break
	end
	
end
init.ws("close")