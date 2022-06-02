local init = {}
local cID = os.getComputerID()
init.url = "http://localhost:3300/"
init.config = {
	debug = false,
	ws = {
		turtle = init.url .. "turtlews",
		pocket = init.url .. "pocketws",
		pocket = init.url .. "computerws",
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
		computer = init.url .. "api/v1/computer/" .. cID.. "/",
		turtle =   init.url .. "api/v1/turtle/" .. cID.. "/",
		pocket =   init.url .. "api/v1/pocket/" .. cID.. "/",
		world  =   init.url .. "api/v1/world/" .. cID.. "/",
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

-- replace all http with ws in init.config.ws
for i, v in pairs(init.config.ws) do
	init.config.ws[i] = v:gsub("http", "ws")
end

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
		print("Websocket opened")
		return true
	else
		print("Websocket failed to open")
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
	--pcall(websocket.close)
	--openWebsocket()
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
init.ws("open")



return init

-- wget run http://localhost:3300/static/init.lua