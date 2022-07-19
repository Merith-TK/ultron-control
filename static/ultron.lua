-- create config structure
local ultron = {}
local cID = os.getComputerID()
ultron.config = {
	debug = true,
	version = "0.0.1",
	api = {
		--host = "http://ultron-api:3300/api",
		host = "https://3300-merithtk-ultroncontrol-2hqr82hg1bo.ws-us54.gitpod.io/api",
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
local function openWebsocket(url)
	ultron.debugPrint("Opening websocket connection ".. ultron.config.api.ws)
	if not url then
		url = ultron.config.api.ws
	end
	local ws = http.websocket(url)
	if ws then
		ultron.websocket = ws
		ultron.debugPrint("Websocket opened")
		return true
	else
		return false
	end
end

local function websocketError(data)
	-- attempt to reconnect to websocket
	if data then
		ultron.debugPrint("Websocket error: " .. data)
	else
		ultron.debugPrint("Websocket error not sent")
	end
	ultron.debugPrint("Attempting to reconnect...")
	sleep(ultron.config.apiDelay)
	openWebsocket()
end

function ultron.ws(connectionType, data)
	if connectionType == "open" then
		openWebsocket(data)
	elseif connectionType == "send" then
		local err, result = pcall(ultron.websocket.send, data)
		if not err then websocketError(result) end
		--ultron.debugPrint("Websocket sent: " .. data)
	elseif connectionType == "receive" then
		local err, result = pcall(ultron.websocket.receive, 1)
		if not err then websocketError(result) end
		ultron.debugPrint("Websocket received: " .. result)
		return result
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
-- processQueue(queue)
-- Processes the queue

function ultron.processCmdQueue(cmdQueue)
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
						ultron.data.cmdResult = cmdResult
					-- else
					-- 	cmdResult = {}
					end
				end
				--ultron.data.cmdResult = cmdResult
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
		ultron.debugPrint("Waiting for orders")
		local event, url, data = os.pullEvent("websocket_message")
		if data then
			ultron.debugPrint("Order Recieved: " .. data)
			data = textutils.unserializeJSON(data)
			-- append data table contents to ultron.data.cmdQueue
			for i = 1, #data do
				if data[i] == "ultron.break()" then
					ultron.data.cmdQueue = {}
					ultron.saveCommandQueue()
					
					os.reboot()
				end
				table.insert(ultron.data.cmdQueue, data[i])
			end
			ultron.debugPrint("cmdQueue: " .. textutils.serialize(ultron.data.cmdQueue))
		else
			ultron.debugPrint("No data recieved")
		end
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