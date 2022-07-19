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

local init = require("../init")
local config = init.config
local turtleData = init.turtleData

init.debugPrint()
init.debugPrint("ApiDelay: " .. config.apiDelay)
init.debugPrint("Websocket URL: " .. config.ws.current)
init.debugPrint("Websocket Header: " .. textutils.serialize(wsHeader))

init.ws("open")
local function fetchTurtleData()
	init.ws("send", "turtle")
	local event, request, data = os.pullEvent("websocket_message")
	if data then
		turtleData = textutils.unserializeJSON(data)
		init.debugPrint("Turtle data received: " .. data)
	else
		init.debugPrint("Turtle data not received")
	end
	init.ws("close")
end

local function fetchTurtle(id, name)
	-- find turtle id in turtleData
	local turtle = nil
	for i,v in ipairs(turtleData) do -- THIS IS WHERE ITS COMPLAINING
		if v.id == id then
			turtle = v
			return v
		end
	end
	if not turtle then
		init.debugPrint("Turtle not found")
		return nil
	else
		init.debugPrint("Turtle found")
		return turtle
	end
end

init.ws("close")