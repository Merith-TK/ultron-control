local init = {}
local cID = os.getComputerID()
init.url = "http://ultron-api:3300/api/"
init.config = {
	debug = true,
	ws = {
		turtle = init.url .. "turtle/ws",
		computer = init.url .. "computer/ws",
	},
	wsHeader = {
		turtle = {
			"turtle",
			tostring(cID),
		},
		computer = {
			"computer",
			tostring(cID),
		},
	},
	api = {
		current = nil,
		computer = init.url .. "computer/" .. cID.. "/",
		turtle =   init.url .. "turtle/" .. cID.. "/",
	},
	luaUrl = init.url .. "static/",
	files = {
		all = {
			"init.lua",
			"startup.lua",
		},
		computer = {
			"computer/Terminal.lua"
		},
		turtle = {
			"turtle/skyrtle.lua",
			"turtle/Terminal.lua",
			"world-post.lua",
		},

	},
	turtle = {
		fuels = {
			"minecraft:coal",
			"minecraft:coal_block",
			"minecraft:charcoal",
		},
	},
	downloadDelay = 0.25,
	apiDelay = 1,
}

init.currentData = {}
init.turtleData = {
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
	inventory = {},
	cmdResult = {},
	cmdQueue = {},
	miscData = {},
}
init.computerData = {
	name = "",
	id = 0,
	cmdResult = {},
	cmdQueue = {},
	miscData = {},
}


-- replace all http with ws in init.config.ws
for i, v in pairs(init.config.ws) do
	init.config.ws[i] = v:gsub("http", "ws")
end



--------------------------------------------------------------------------------
-- debugPrint(string)
-- Prints a string to the terminal if debug is enabled
--------------------------------------------------------------------------------
function init.debugPrint(str)
	if init.config.debug then
		if str then
			print("[Debug] "..str)
		end
	end
end

init.debugPrint("Enabled")

--------------------------------------------------------------------------------
-- system setup
--------------------------------------------------------------------------------

local function downloadFiles(files)
	for _, file in ipairs(files) do
		-- download files using http
		init.debugPrint("Downloading file: " .. file)
 		sleep(init.config.downloadDelay)
 		local url = init.config.luaUrl .. file
		init.debugPrint(url)
 		local localfile = fs.open(file, "w")
		local dl = http.get(url)
 		if dl then
 			localfile.write(dl.readAll())
 		else
 			print("[Error]: Unable to download " .. file)
 		end
 		localfile.close()
 	end
end

-- create websocket based off of what type of computer we are running on
if turtle then
	init.currentData = init.turtleData
	init.config.ws.current = init.config.ws.turtle
	init.config.wsHeader.current = init.config.wsHeader.turtle
elseif pocket then
	init.currentData = init.pocketData
	init.config.ws.current = init.config.ws.pocket
	init.config.wsHeader.current = init.config.wsHeader.pocket
else
	init.currentData = init.computerData
	init.config.ws.current = init.config.ws.computer
	init.config.wsHeader.current = init.config.wsHeader.computer
end

-- check if running as sub-shell
if shell.getRunningProgram() == "rom/programs/http/wget.lua" then
	-- check if startup exists
	local startup = false
	if fs.exists("/startup.lua") then startup = true end

	shell.run("set motd.enable false")
	term.clear()
	term.setCursorPos(1,1)

	print("[Updater]: Auto-updating...")
	downloadFiles(init.config.files.all)
	if turtle then
		downloadFiles(init.config.files.turtle)
	elseif pocket then
		downloadFiles(init.config.files.pocket)
	elseif commands then
		print("[Error]: This program is not compatible with the Command Computer for security reasons.")
		return
	else
		downloadFiles(init.config.files.computer)
	end
	print("[Updater] Update complete")

	if not startup then
		os.reboot()
	end
end



-- open websocket
init.websocket = {}
local function openWebsocket()
	local ws = http.websocket(init.config.ws.current, init.config.wsHeader.current)
	if ws then
		init.websocket = ws
		init.debugPrint("Websocket opened")
		return true
	else
		return false
	end
end

local function websocketError(data)
	-- attempt to reconnect to websocket
	if data then
		init.debugPrint("Websocket error: " .. data)
	else
		init.debugPrint("Websocket error not sent")
	end
	init.debugPrint("Attempting to reconnect...")
	sleep(init.config.apiDelay)
	openWebsocket()
end

function init.ws(connectionType, data)
	if connectionType == "open" then
		openWebsocket()
	elseif connectionType == "send" then
		local err, result = pcall(init.websocket.send, data)
		if not err then websocketError(result) end
		--init.debugPrint("Websocket sent: " .. data)
	elseif connectionType == "receive" then
		local err, result = pcall(init.websocket.receive, 1)
		if not err then websocketError(result) end
		init.debugPrint("Websocket received: " .. result)
		return result
	elseif type == "close" then
		local err, result = pcall(init.websocket.close)
		if not err then websocketError(result) end
		init.debugPrint("Websocket closed")
	end
end

--------------------------------------------------------------------------------
-- loadCommandQueue()
-- Loads the command queue from file
function init.loadCommandQueue()
	local cmdQueue = {}
	local file = fs.open("/cmdQueue.json", "r")
	if file then
		local data = file.readAll()
		file.close()
		if data then
			cmdQueue = textutils.unserializeJSON(data)
			init.debugPrint("cmdQueue: " .. textutils.serialize(cmdQueue))
			return cmdQueue
		else
			init.debugPrint("cmdQueue: " .. textutils.serialize(cmdQueue))
			return cmdQueue
		end
	else
		init.debugPrint("cmdQueue: " .. textutils.serialize(cmdQueue))
		init.currentData.cmdQueue = cmdQueue
	end
end

--------------------------------------------------------------------------------
-- saveCommandQueue()
-- Saves the command queue to file
function init.saveCommandQueue()
	local file = fs.open("/cmdQueue.json", "w")
	if file then
		file.write(textutils.serializeJSON(init.currentData.cmdQueue))
		file.close()
	end
end

--------------------------------------------------------------------------------
-- processQueue(queue)
-- Processes the queue

function init.processCmdQueue(cmdQueue)
	while true do
		-- check if cmdQueue is empty
		if #init.currentData.cmdQueue ~= 0 then
			while #init.currentData.cmdQueue ~= 0 do
				cmdResult = nil
				local cmd = table.remove(init.currentData.cmdQueue, 1)
				local file = fs.open("/cmdQueue.json", "w")
				file.write(textutils.serializeJSON(init.currentData.cmdQueue))
				file.close()
				if cmd then
					local cmdExec, err = loadstring(cmd)
					if cmdExec then
						print("[cmd:in] = " .. cmd)
						setfenv(cmdExec, getfenv())
						local result = {pcall(cmdExec)}
						cmdResult = result
						if result then
							result = textutils.serialize(cmdResult)
						end
						print("[cmd:out] = " .. tostring(result))
						init.currentData.cmdResult = cmdResult
					-- else
					-- 	cmdResult = {}
					end
				end
				--init.currentData.cmdResult = cmdResult
			end
		else
			sleep()
		end
	end
end

--------------------------------------------------------------------------------
-- waitForOrders(queue)
-- Waits for orders from the websocket
function init.recieveOrders()
	while true do
		init.debugPrint("Waiting for orders")
		local event, url, data = os.pullEvent("websocket_message")
		if data then
			init.debugPrint("Order Recieved: " .. data)
			data = textutils.unserializeJSON(data)
			-- append data table contents to init.currentData.cmdQueue
			for i = 1, #data do
				if data[i] == "ultron.break()" then
					init.currentData.cmdQueue = {}
					init.saveCommandQueue()
					
					os.reboot()
				end
				table.insert(init.currentData.cmdQueue, data[i])
			end
			init.debugPrint("cmdQueue: " .. textutils.serialize(init.currentData.cmdQueue))
		else
			init.debugPrint("No data recieved")
		end
	end
end



--------------------------------------------------------------------------------
-- waitForDelay()
-- Waits for the api delay
function init.waitForDelay()
	--init.debugPrint("Waiting for apiDelay")
	sleep(init.config.apiDelay)
end

return init

-- wget run http://localhost:3300/static/init.lua