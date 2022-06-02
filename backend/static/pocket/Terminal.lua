local pocketData = {
	name = "",
	id = 0,
	pos = {
		x = 0,
		y = 0,
		z = 0,
		r = 0,
		rname = "",
	},
	miscData = {},
}
local turtleData = {}
local init = require("../init")
local config = init.config

init.debugPrint()

init.debugPrint("ApiDelay: " .. config.apiDelay)
init.debugPrint("Websocket URL: " .. config.ws.current)
init.debugPrint("Websocket Header: " .. textutils.serialize(wsHeader))

local function fetchTurtleData()
	init.ws("send", "turtle")
	local event, request, data = os.pullEvent("websocket_message")
	if data then
		turtleData = textutils.unserialize(data)
		init.debugPrint("Turtle data received: " .. data)
	else
		init.debugPrint("Turtle data not received")
	end
end

local function fetchTurtle(name, id)
	-- find turtle id in turtleData
	local turtle = nil
	for i,v in pairs(turtleData) do
		if v.id == id then
			turtle = v
			return v
		end
	end
	if not turtle then
		init.debugPrint("Turtle not found")
		return nil
	end
end
fetchTurtleData()
print(fetchTurtle(1))
