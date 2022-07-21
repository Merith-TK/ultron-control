-- create config structure
local ultron = {}
local cID = os.getComputerID()
ultron.config = {
	debug = true,
	version = "0.0.1",
	api = {
		host = "http://ultron-api:3300/api",
		delay = 0.5,
		timeout = 5,
		retries = 3,
	},
}
ultron.module = ""
ultron.data = {}

--------------------------------------------------------------------------------
-- debugPrint(string)
-- Prints a string to the terminal if debug is enabled
--------------------------------------------------------------------------------
function ultron.debugPrint(str)
	if ultron.config.debug then
		if str then
			print("[Debug] "..str)
		end
	end
end


-- get list of modiles from api
function ultron.get_modules()
	local response = http.get(ultron.config.api.host .. "/modules")
	if response then
		local data = textutils.unserializeJSON(response.readAll())
		if data then
			ultron.modules = data
		end
	end
	return ultron.modules
end
ultron.get_modules()


--------------------------------------------------------------------------------
-- extra functions for the api
--------------------------------------------------------------------------------


-- open websocket
ultron.websocket = {}
local wsError = false
local function openWebsocket(url)
	if not url then
		url = ultron.config.api.ws
	end
	local ws = http.websocket(url)
	if ws then
		ultron.websocket = ws
		ultron.debugPrint("Websocket opened")
		wsError = false
		return true
	else
		return false
	end
end

local function websocketError(data)
	-- attempt to reconnect to websocket
	if not wsError then
		wsError = true
		if data then
			ultron.debugPrint("Websocket error: " .. data)
		else
			ultron.debugPrint("Websocket error")
		end
		ultron.debugPrint("Attempting to reconnect to websocket")
		if openWebsocket() then
			wsError = false
			ultron.debugPrint("Websocket reconnected")
		end
	end
	--ultron.debugPrint("Attempting to reconnect...")
	sleep(ultron.config.apiDelay)
	openWebsocket()
end

function ultron.ws(connectionType, data)
	if connectionType == "open" then
		if not openWebsocket(data) then
			websocketError()
		end
	elseif connectionType == "send" then
		local err, result = pcall(ultron.websocket.send, data)
		if not err then websocketError(result) end
		--ultron.debugPrint("Websocket sent: " .. data)
	elseif connectionType == "receive" then
		local err, result = pcall(ultron.websocket.receive, 1)
		if not err then websocketError(result) end
		if result then
			ultron.debugPrint("Websocket received: " .. result)
			return result
		else return nil end
	elseif type == "close" then
		local err, result = pcall(ultron.websocket.close)
		if not err then websocketError(result) end
		ultron.debugPrint("Websocket closed")
	end
end

--------------------------------------------------------------------------------
-- loadCommandQueue()
-- Loads the command queue from file
function ultron.loadCommandQueue()
	local cmdQueue = {}
	local file = fs.open("/cmdQueue.json", "r")
	if file then
		local data = file.readAll()
		file.close()
		if data then
			cmdQueue = textutils.unserializeJSON(data)
			ultron.debugPrint("cmdQueue: " .. textutils.serialize(cmdQueue))
			return cmdQueue
		else
			ultron.debugPrint("cmdQueue: " .. textutils.serialize(cmdQueue))
			return cmdQueue
		end
	else
		ultron.debugPrint("cmdQueue: " .. textutils.serialize(cmdQueue))
		ultron.data.cmdQueue = cmdQueue
	end
end

--------------------------------------------------------------------------------
-- saveCommandQueue()
-- Saves the command queue to file
function ultron.saveCommandQueue()
	local file = fs.open("/cmdQueue.json", "w")
	if file then
		file.write(textutils.serializeJSON(ultron.data.cmdQueue))
		file.close()
	end
end



--------------------------------------------------------------------------------
-- processCmd(cmd)
-- Processes a single command
function ultron.processCmd(cmd)
	if not cmd then
		return false, "No command given"
	end
	local cmdExec, err = load(cmd, nil, "t", _ENV)
	if cmdExec then
		print("[cmd:in] = " .. cmd)
		local result = {pcall(cmdExec)}
		cmdResult = result
		if result then
			result = textutils.serialize(cmdResult)
		end
		print("[cmd:out] = " .. tostring(result))
	end
end
--------------------------------------------------------------------------------
-- processQueue(queue)
-- Processes the queue
function ultron.processCmdQueue(cmdQueue)
	ultron.data.cmdQueue = cmdQueue
	while true do
		-- check if cmdQueue is empty
		if #ultron.data.cmdQueue ~= 0 then
			while #ultron.data.cmdQueue ~= 0 do
				cmdResult = nil
				local cmd = table.remove(ultron.data.cmdQueue, 1)
				local file = fs.open("/cmdQueue.json", "w")
				file.write(textutils.serializeJSON(ultron.data.cmdQueue))
				file.close()
				if cmd then
					ultron.debugPrint("Processing cmd: " .. cmd)
					cmdResult = ultron.processCmd(cmd)
					if cmdResult then
						ultron.debugPrint("cmdResult: " .. cmdResult)
					end
				end
			end
		else
			sleep()
		end
	end
end

--------------------------------------------------------------------------------
-- waitForOrders(queue)
-- Waits for orders from the websocket
function ultron.recieveOrders()
	while true do
		sleep()
		if not wsError then
			local data = ultron.ws("receive")
			if data then
				local data = textutils.unserializeJSON(data)
				if data then
					ultron.debugPrint("Received orders: " .. textutils.serialize(data))
					if data.type == "orders" then
						ultron.data.orders = data.orders
						ultron.debugPrint("Orders: " .. textutils.serialize(ultron.data.orders))
						ultron.saveOrders()
					end
				end
			end
		end
		-- ultron.debugPrint("Waiting for orders")
		-- local event, url, data = os.pullEvent("websocket_message")
		-- if data then
		-- 	ultron.debugPrint("Order Recieved: " .. data)
		-- 	data = textutils.unserializeJSON(data)
		-- 	-- append data table contents to ultron.data.cmdQueue
		-- 	for i = 1, #data do
		-- 		if data[i] == "ultron.break()" then
		-- 			ultron.data.cmdQueue = {}
		-- 			ultron.saveCommandQueue()
					
		-- 			os.reboot()
		-- 		end
		-- 		table.insert(ultron.data.cmdQueue, data[i])
		-- 	end
		-- 	ultron.debugPrint("cmdQueue: " .. textutils.serialize(ultron.data.cmdQueue))
		-- else
		-- 	ultron.debugPrint("No data recieved")
		-- end
	end
end



--------------------------------------------------------------------------------
-- waitForDelay()
-- Waits for the api delay
function ultron.waitForDelay()
	--ultron.debugPrint("Waiting for apiDelay")
	sleep(ultron.config.apiDelay)
end

--------------------------------------------------------------------------------
-- download module
--------------------------------------------------------------------------------
function ultron.download_module(moduleName)
		-- download files using http
		local file = "module.lua"
		ultron.debugPrint("Downloading Module: " .. moduleName)
 		local url = ultron.config.api.host .. "/" .. moduleName .. "/fs/" .. file
		ultron.debugPrint(url)
 		local localfile = fs.open(file, "w")
		local dl = http.get(url)
 		if dl then
 			localfile.write(dl.readAll())
 		else
 			print("[Error]: Unable to download " .. file)
 		end
 		localfile.close()
end

function ultron.wget(file, url)
	local localfile = fs.open(file, "w")
	local dl = http.get(url)
	if dl then
		local data = dl.readAll()
		if data ~= "" then
			localfile.write(data)
		else
			print("[Err] Could not download '".. file.. "' recieved No Data")
		end
	else
		print("[Error]: Unable to download "..file)
	end
	localfile.close()
	dl.close()
end
if shell.getRunningProgram() == "rom/programs/http/wget.lua" then
	ultron.wget("startup.lua", ultron.config.api.host .. "/static/startup.lua")
	ultron.wget("ultron.lua", ultron.config.api.host .. "/static/ultron.lua")
end


return ultron