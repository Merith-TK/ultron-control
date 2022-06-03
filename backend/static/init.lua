local init = {}
local cID = os.getComputerID()
init.url = "http://localhost:3300/"
init.config = {
	debug = true,
	ws = {
		turtle = init.url .. "turtlews",
		pocket = init.url .. "pocketws",
		computer = init.url .. "computerws",
	},
	wsHeader = {
		current = nil,
		turtle = {
			"turtle",
			tostring(cID),
		},
		pocket = {
			"pocket",
			tostring(cID),
		},
		computer = {
			"computer",
			tostring(cID),
		},
	},
	api = {
		current = nil,
		computer = init.url .. "api/computer/" .. cID.. "/",
		turtle =   init.url .. "api/turtle/" .. cID.. "/",
		pocket =   init.url .. "api/pocket/" .. cID.. "/",
		world  =   init.url .. "api/world/",
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
		},
		pocket = {
			"pocket/Terminal.lua",
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

init.cmdResult = false
init.cmdQueue = {}

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

-- open websocket
init.websocket = {}
local function openWebsocket()
	local ws = http.websocket(init.config.ws.current, init.config.wsHeader.current)
	if ws then
		init.websocket = ws
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
		init.debugPrint("Websocket opened")
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
-- processQueue(queue)
-- Processes the queue

function init.processCmdQueue(cmdQueue)
	while true do
		-- check if cmdQueue is empty
		if #init.cmdQueue ~= 0 then
			while #init.cmdQueue ~= 0 do
				cmdResult = nil
				init.debugPrint("Executing init.cmdQueue")
				local cmd = table.remove(init.cmdQueue, 1)
				print("Executing cmd: " .. cmd)
				local file = fs.open("/cmdQueue.json", "w")
				file.write(textutils.serializeJSON(init.cmdQueue))
				file.close()
				if cmd then
					init.debugPrint("cmd: " .. cmd)
					init.debugPrint("Processing command: " .. cmd)
					local cmdExec, err = loadstring(cmd)
					if cmdExec then
						print("Executing command: " .. cmd)
						setfenv(cmdExec, getfenv())
						local success, result = pcall(cmdExec)
							cmdResult = result
					else
						cmdResult = nil
					end
				end
				print("Commands left: " .. #init.cmdQueue)
			end
			init.cmdResult = cmdResult
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
			-- append data table contents to init.cmdQueue
			for i = 1, #data do
				if data[i] == "ultron.break()" then
					-- clear init.cmdQueue
					init.cmdQueue = {}
					os.reboot()
				end
				table.insert(init.cmdQueue, data[i])
			end
			init.debugPrint("cmdQueue: " .. textutils.serialize(init.cmdQueue))
		else
			init.debugPrint("No data recieved")
		end
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
			init.debugPrint("cmdQueue: " .. data)
			cmdQueue = textutils.unserializeJSON(data)
			init.debugPrint("cmdQueue: " .. textutils.serialize(cmdQueue))
			return cmdQueue
		else
			init.debugPrint("cmdQueue: " .. textutils.serialize(cmdQueue))
			return cmdQueue
		end
	else
		init.debugPrint("cmdQueue: " .. textutils.serialize(cmdQueue))
		init.cmdQueue = cmdQueue
	end
end

--------------------------------------------------------------------------------
-- waitForDelay()
-- Waits for the api delay
function init.waitForDelay()
	--init.debugPrint("Waiting for apiDelay")
	sleep(init.config.apiDelay)
end


--------------------------------------------------------------------------------
-- setup system
--------------------------------------------------------------------------------
-- download files
local function downloadFiles(files)
	for _, file in ipairs(files) do
		-- download files using http
		print("[Downloading]: " .. file)
 		sleep(init.config.downloadDelay)
 		local url = init.config.luaUrl .. file
 		local file = fs.open(file, "w")
 		if http.get(url).readAll() then
 			file.write(http.get(url).readAll())
 		else
 			print("[Error]: Unable to download " .. file)
 		end
 		file.close()
 	end
end

-- create websocket based off of what type of computer we are running on
if turtle then
	init.config.ws.current = init.config.ws.turtle
	init.config.wsHeader.current = init.config.wsHeader.turtle
elseif pocket then
	init.config.ws.current = init.config.ws.pocket
	init.config.wsHeader.current = init.config.wsHeader.pocket
else
	init.config.ws.current = init.config.ws.computer
	init.config.wsHeader.current = init.config.wsHeader.computer
end

-- check if running as sub-shell
if shell.getRunningProgram() == "rom/programs/http/wget.lua" then
	shell.run("set motd.enable false")
	term.clear()
	term.setCursorPos(1,1)

	print("[Updating]: Auto-updating...")
	downloadFiles(init.config.files.all)
	if turtle then
		downloadFiles(init.config.files.turtle)
		local skyrtle = require("/turtle/skyrtle")
		-- get skyrtle position
		local x,y,z = skyrtle.getPosition()
		local facing = skyrtle.getFacing()
		if x == 0 and y == 0 and z == 0 and facing == 0 then
			-- get user input for position
			print("Please enter turtle position")
			print("Press f3 and look at turtle to get position")
			print("X Y Z Facing")
			print("Example: 0 0 0 north")
			local userPos = read()
			-- convert userpos to table seperated by spaces
			local pos = {}
			for i in string.gmatch(userPos, "%S+") do
				table.insert(pos, i)
			end
			-- create skyrtle file
			skyrtle.setPosition(tonumber(pos[1]), tonumber(pos[2]), tonumber(pos[3]))
			skyrtle.setFacing(pos[4])
		end
		print("[Skyrtle]: Location ", skyrtle.getPosition())
		print("[Skyrtle]: Facing ", skyrtle.getFacing())
	elseif pocket then
		downloadFiles(init.config.files.pocket)
	elseif commands then
		print("[Error]: This program is not compatible with the Command Computer for security reasons.")
		return
	else
		downloadFiles(init.config.files.computer)
	end
	print("[Update] Update complete")
end


return init

-- wget run http://localhost:3300/static/init.lua