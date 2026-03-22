-- create config structure
local ultron = {}
local cID = os.getComputerID()
-- ultron.config is the main config that the host of the API can use
-- you should not need to change more than ultron.config.host
ultron.cmd = ""
ultron.config = {
    debug = true,
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
    sleep(ultron.config.api.delay)
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
    elseif connectionType == "close" then
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
        return {
            success = false,
            error = "No command queue given",
            command = cmdQueue or "",
            timestamp = os.epoch("utc")
        }
    end
    
    if cmdQueue ~= "" then
        ultron.debugPrint("Processing cmd: " .. cmdQueue)
        local startTime = os.epoch("utc")
        
        local cmdExec, compileErr = load(cmdQueue, nil, "t", _ENV)
        if cmdExec then
            print("[cmd:in] = " .. cmdQueue)
            
            -- Execute command and capture result
            local success, result = pcall(cmdExec)
            local endTime = os.epoch("utc")
            
            -- Create detailed result structure
            local cmdResult = {
                success = success,
                command = cmdQueue,
                result = success and result or nil,
                error = (not success) and tostring(result) or nil,
                startTime = startTime,
                endTime = endTime,
                executionTime = endTime - startTime,
                timestamp = endTime
            }
            
            print("[cmd:out] = " .. (success and tostring(result) or ("ERROR: " .. tostring(result))))
            
            -- Store in turtle data for legacy compatibility
            ultron.data.cmdResult = cmdResult
            ultron.cmd = "" -- Clearing ultron.cmd after processing the command
            
            return cmdResult
        else
            -- Compilation error
            local endTime = os.epoch("utc")
            local cmdResult = {
                success = false,
                command = cmdQueue,
                error = "Compilation error: " .. (compileErr or "unknown error"),
                startTime = startTime,
                endTime = endTime,
                executionTime = endTime - startTime,
                timestamp = endTime
            }
            
            print("[cmd:err] = " .. cmdResult.error)
            ultron.data.cmdResult = cmdResult
            ultron.cmd = ""
            
            return cmdResult
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
        if wsError then
            if openWebsocket() then
                wsError = false
                ultron.debugPrint("Websocket reconnected")
            end
        else
            local data = ultron.ws("receive")
            if data then
                -- Try to parse as JSON first
                local parsedData = textutils.unserializeJSON(data)

                if parsedData then
                    -- Check if it's a structured command request
                    if parsedData.command and parsedData.requestId then
                        ultron.debugPrint("Received sync command: " .. parsedData.requestId)
                        -- Handle synchronous command
                        ultron.handleSyncCommand(parsedData)
                    else
                        -- Handle legacy data format
                        ultron.debugPrint("Received orders: " .. textutils.serialize(parsedData))
                        ultron.cmd = parsedData
                    end
                else
                    -- Handle raw string commands (legacy)
                    ultron.debugPrint("Received raw command: " .. data)
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

-- Handle synchronous command execution
function ultron.handleSyncCommand(cmdRequest)
    local requestId = cmdRequest.requestId
    local command = cmdRequest.command
    
    ultron.debugPrint("Executing sync command [" .. requestId .. "]: " .. command)
    
    -- Execute the command and capture result
    local result = ultron.processCmd(command)
    
    -- Create response structure
    local response = {
        requestId = requestId,
        result = result,
        success = (result ~= nil and result ~= false)
    }
    
    -- Send response back to server
    local responseJson = textutils.serializeJSON(response)
    ultron.ws("send", responseJson)
    
    ultron.debugPrint("Sent sync response [" .. requestId .. "]")
end

function ultron.wget(file, url)
    -- First, attempt to download the data before touching the local file
    local dl = http.get(url)
    local didError = false
    local data = nil

    if dl then
        data = dl.readAll()
        dl.close()
        if data == "" or data == nil then
            print("[Err] Could not download '" .. file .. "' received No Data")
            didError = true
        end
    else
        print("[Error]: Unable to download " .. file .. " - server unreachable")
        didError = true
    end

    -- Only proceed with file operations if download was successful
    if not didError then
        -- Create backup before overwriting
        if fs.exists(file) then fs.copy(file, file .. ".bak") end

        -- Now write the downloaded data
        local localfile = fs.open(file, "w")
        localfile.write(data)
        localfile.close()

        -- Clean up backup on success
        if fs.exists(file .. ".bak") then fs.delete(file .. ".bak") end

        ultron.debugPrint("Successfully updated " .. file)
    else
        ultron.debugPrint("Preserving existing " .. file .. " - server unreachable")
    end
end

-- for remote install, use `wget run <ultron.config.api.host>/api/static/ultron.lua`
if shell.getRunningProgram() == "rom/programs/http/wget.lua" then
    ultron.wget("startup.lua", ultron.config.api.host .. "/static/startup.lua")
    ultron.wget("ultron.lua", ultron.config.api.host .. "/static/ultron.lua")
end

return ultron
