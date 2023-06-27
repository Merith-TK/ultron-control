-- create config structure
local ultron = {}
local cID = os.getComputerID()
-- ultron.config is the main config that the host of the API can use
-- you should not need to change more than ultron.config.host
ultron.cmd = ""
ultron.config = {
    debug = false,
    version = "0.0.1",
    api = {
        host = "http://localhost:3300/api",
        delay = 0.5,
        timeout = 5,
        retries = 3
    }
}
ultron.module = ""
ultron.data = {}

-- If cfg/debug exists, enable debug
if fs.exists("/cfg/debug") then ultron.config.debug = true end

--------------------------------------------------------------------------------
-- debugPrint(string)
-- Prints a string to the terminal if debug is enabled
--------------------------------------------------------------------------------
function ultron.debugPrint(str)
    if ultron.config.debug then if str then print("[Debug] " .. str) end end
end

-- get list of modiles from api
function ultron.get_modules()
    local response = http.get(ultron.config.api.host .. "/modules")
    if response then
        local data = textutils.unserializeJSON(response.readAll())
        if data then ultron.modules = data end
    end
    return ultron.modules
end
ultron.get_modules()

--------------------------------------------------------------------------------
-- Remote Target Functions
--------------------------------------------------------------------------------

-- open websocket
ultron.websocket = {}
local wsError = false
local function openWebsocket(url)
    if not url then url = ultron.config.api.ws end
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
    -- ultron.debugPrint("Attempting to reconnect...")
    sleep(ultron.config.apiDelay)
    openWebsocket()
end

function ultron.ws(connectionType, data)
    if connectionType == "open" then
        if not openWebsocket(data) then websocketError() end
    elseif connectionType == "send" then
        local err, result = pcall(ultron.websocket.send, data)
        if not err then websocketError(result) end
        -- ultron.debugPrint("Websocket sent: " .. data)
    elseif connectionType == "receive" then
        local err, result = pcall(ultron.websocket.receive, 1)
        if not err then 
            websocketError(result)
        else
            if result then
                ultron.debugPrint("Websocket received: " .. result)
                ultron.cmd = result
                return result
            else
                return nil
            end
        end
    elseif type == "close" then
        local err, result = pcall(ultron.websocket.close)
        if not err then websocketError(result) end
        ultron.debugPrint("Websocket closed")
    end
end



--------------------------------------------------------------------------------
-- processCmd(cmd)
-- Processes a single command
function ultron.processCmd(cmdQueue)
    if not cmdQueue or cmdQueue == "" then
        return false, "No command queue given"
    end
    if cmdQueue ~= "" then
        ultron.debugPrint("Processing cmd: " .. cmdQueue)
        local cmdExec, err = load(cmdQueue, nil, "t", _ENV)
        if cmdExec then
            print("[cmd:in] = " .. cmdQueue)
            local result = {pcall(cmdExec)}
            cmdResult = result
            if result then
                result = textutils.serialize(cmdResult, {compact = true})
            end
            print("[cmd:out] = " .. tostring(result))
            ultron.data.cmdResult = result
            ultron.cmd = "" -- Clearing ultron.cmd after processing the command
            return result
        end
	else
        sleep(ultron.config.api.delay)
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
                data = textutils.unserializeJSON(data)
                if data then
                    ultron.debugPrint("Received orders: " .. textutils.serialize(data))
                    ultron.cmd = data
                end
            end
        end
    end
end

-- waitForDelay()
-- Waits for the api delay
function ultron.waitForDelay()
    -- ultron.debugPrint("Waiting for apiDelay")
    sleep(ultron.config.apiDelay)
end

--------------------------------------------------------------------------------
-- download module
--------------------------------------------------------------------------------
function ultron.download_module(moduleName)
    -- download files using http
    ultron.debugPrint("Downloading Module: " .. moduleName)
    local url = ultron.config.api.host .. "/" .. moduleName .. "/fs/module.lua" -- .. file
    ultron.debugPrint(url)
    local localfile = fs.open("module.lua", "w")
    local dl = http.get(url)
    if dl then
        localfile.write(dl.readAll())
    else
        print("[Error]: Unable to download module")
    end
    localfile.close()
end

function ultron.wget(file, url)
    local localfile = fs.open(file, "w")
    if fs.exists(file) then fs.copy(file, file .. ".bak") end
    local dl = http.get(url)
    local didError = false
    if dl then
        local data = dl.readAll()
        if data ~= "" then
            localfile.write(data)
        else
            print("[Err] Could not download '" .. file .. "' recieved No Data")
            didError = true
        end
        dl.close()
    else
        print("[Error]: Unable to download " .. file)
        didError = true
    end
    localfile.close()
    if didError then
        if fs.exists(file .. ".bak") then
            fs.delete(file)
            fs.move(file .. ".bak", file)
        end
    else
        if fs.exists(file .. ".bak") then fs.delete(file .. ".bak") end
    end
end

-- for remote install, use `wget run <ultron.config.api.host>/api/static/ultron.lua`
if shell.getRunningProgram() == "rom/programs/http/wget.lua" then
    ultron.wget("startup.lua", ultron.config.api.host .. "/static/startup.lua")
    ultron.wget("ultron.lua", ultron.config.api.host .. "/static/ultron.lua")
end

return ultron
