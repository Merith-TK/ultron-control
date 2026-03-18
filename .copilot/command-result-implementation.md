# Command Result Return Implementation Plan

**Date**: August 8, 2025  
**Goal**: Return command execution results in POST response

## Current Flow Analysis

### Existing Process:
1. **Client POSTs** command to `/api/turtle/{id}`
2. **Server** adds command to `Turtles[pos].CmdQueue`
3. **Turtle** receives command via websocket in `ultron.recieveOrders()`
4. **Turtle** processes command in `processCmdQueue()` using `ultron.processCmd()`
5. **Turtle** sets `ultron.data.cmdResult` with execution result
6. **Turtle** sends updated data back to server via websocket
7. **Server** updates `Turtles[pos].CmdResult` when turtle data received

### Current Problems:
- POST request returns immediately after adding to queue
- No correlation between specific command and its result
- Client must poll `/api/turtle/{id}/cmdResult` to get results
- No way to know which result belongs to which command

## Implementation Strategy

### ⚠️ **Critical Constraint: Minimal Persistent Data**
- API server must be stateless and restartable
- Turtles must be stateless and restartable  
- No persistent storage requirements
- Clients handle their own data persistence

### Revised Approach: Synchronous Command Execution

#### Option 1: Blocking POST with Immediate Response
Instead of tracking commands across server restarts, we execute commands synchronously within the websocket connection:

```go
func TurtleHandle(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        // Find active websocket connection for this turtle
        turtleConn := findActiveTurtleConnection(turtleID)
        if turtleConn == nil {
            ReturnError(w, http.StatusServiceUnavailable, "Turtle not connected")
            return
        }
        
        // Generate unique command ID for this request only
        cmdID := generateCommandID()
        
        // Send command directly via websocket
        commandMsg := map[string]interface{}{
            "id": cmdID,
            "command": commandString,
        }
        
        err := turtleConn.WriteJSON(commandMsg)
        if err != nil {
            ReturnError(w, http.StatusInternalServerError, "Failed to send command")
            return
        }
        
        // Wait for immediate response with timeout
        result := waitForImmediateResult(turtleConn, cmdID, 30*time.Second)
        ReturnData(w, result)
    }
}
```

#### Option 2: Fire-and-Forget with Status Endpoint
Simpler approach - POST returns immediately, client polls a separate status endpoint:

```go
// POST /api/turtle/{id} - Fire and forget
func TurtleHandle(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        // Add command to turtle's queue (existing behavior)
        Turtles[pos].CmdQueue = append(Turtles[pos].CmdQueue, commandString)
        
        // Return success with suggestion to check status
        response := map[string]interface{}{
            "status": "queued",
            "message": "Command queued successfully",
            "checkResultAt": fmt.Sprintf("/api/turtle/%d/cmdResult", turtleID),
        }
        ReturnData(w, response)
    }
}
```

### Phase 1: Command Tracking System

#### Recommended: Option 3 - Collaborative Structure Design

Both sides contribute to structure definition while maintaining intentional flexibility:

**Server Side (Core Structure + Flexible Fields):**
```go
type Turtle struct {
    Name         string        `json:"name"`         // Server-defined core field
    ID           int           `json:"id"`           // Server-defined core field
    Pos          Position      `json:"pos"`          // Server-defined structure
    Fuel         Fuel          `json:"fuel"`         // Server-defined structure
    SelectedSlot int           `json:"selectedSlot"` // Server-defined core field
    Inventory    []interface{} `json:"inventory"`    // Server-defined list, flexible content
    
    // Flexible fields - content defined by turtle
    Sight     interface{} `json:"sight"`     // Turtle defines sight data format
    CmdResult interface{} `json:"cmdResult"` // Turtle defines command result format
    Misc      interface{} `json:"misc"`      // Turtle defines misc data format
    HeartBeat int         `json:"heartbeat"` // Server-defined for connection tracking
}

// Core position structure defined by server
type Position struct {
    X     int    `json:"x"`
    Y     int    `json:"y"`
    Z     int    `json:"z"`
    R     int    `json:"r"`
    Rname string `json:"rname"`
}

// Command flow structure - server defines the protocol
type CommandRequest struct {
    Command   string `json:"command"`   // Raw command string
    RequestID string `json:"requestId"` // For synchronous response tracking
}

type CommandResponse struct {
    RequestID string      `json:"requestId"` // Match with request
    Result    interface{} `json:"result"`    // Turtle-defined result format
    Success   bool        `json:"success"`   // Server-defined success indicator
}
```

**Turtle Side (Flexible Content Within Structure):**
```lua
-- Turtle defines content for flexible fields while respecting server structure

-- Sight data format - turtle's choice
ultron.data.sight = {
    up = {name = "minecraft:stone", state = {...}},
    down = {name = "minecraft:dirt", hardness = 0.5},
    front = {} -- empty when no block
}

-- Command result format - turtle's choice but within server's response structure
function ultron.processCmd(command)
    local cmdResult = {
        -- Turtle-defined detailed result structure
        command = command,
        output = capturedOutput,
        returnValue = actualResult,
        executionTime = executionTime,
        success = success,
        errorMessage = errorMsg,
        timestamp = os.epoch("utc"),
        -- Any other data turtle finds useful
        debugInfo = {...}
    }
    
    -- Fits into server's CmdResult interface{} field
    ultron.data.cmdResult = cmdResult
    
    -- For synchronous requests, send structured response
    if synchronousRequest then
        local response = {
            requestId = requestId,
            result = cmdResult,  -- Turtle's flexible content
            success = success    -- Server's required field
        }
        ultron.ws("send", textutils.serializeJSON(response))
    end
    
    return cmdResult
end

-- Misc data - completely turtle-defined
ultron.data.misc = {
    startup = {pastebin = {...}},
    customFeatures = {...},
    userPreferences = {...},
    -- Whatever turtle wants to track
}
```

**Server Side (Minimal Structure):**
```go
type ActiveConnections struct {
    turtleConnections map[int]*websocket.Conn // turtle ID -> connection
    mutex             sync.RWMutex
}

var connections = &ActiveConnections{
    turtleConnections: make(map[int]*websocket.Conn),
}

func TurtleHandle(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        // Check if turtle is currently connected
        connections.mutex.RLock()
        turtleConn, connected := connections.turtleConnections[turtleID]
        connections.mutex.RUnlock()
        
        if !connected {
            ReturnError(w, http.StatusServiceUnavailable, "Turtle not connected")
            return
        }
        
        // Let turtle handle command structure - just pass raw command
        result, err := executeCommandSync(turtleConn, commandString, 30*time.Second)
        if err != nil {
            ReturnError(w, http.StatusInternalServerError, err.Error())
            return
        }
        
        ReturnData(w, result)
    }
}

func executeCommandSync(conn *websocket.Conn, command string, timeout time.Duration) (interface{}, error) {
    // Generate simple request ID
    requestID := generateRequestID()
    
    // Create response channel for this request only
    responseChan := make(chan interface{}, 1)
    
    // Store in temporary map (request scope only)
    requestChannels := map[string]chan interface{}{
        requestID: responseChan,
    }
    
    // Send raw command - let turtle structure it
    err := conn.WriteMessage(websocket.TextMessage, []byte(command))
    if err != nil {
        return nil, fmt.Errorf("failed to send command: %v", err)
    }
    
    // Wait for turtle's response
    select {
    case result := <-responseChan:
        return result, nil
    case <-time.After(timeout):
        return nil, fmt.Errorf("command timeout")
    }
}
```

**Turtle Side (Full Control):**
```lua
-- Turtle manages its own command structure
ultron.pendingCommands = {}  -- Track commands turtle is processing

-- Enhanced command processing with turtle-defined structure
function ultron.processCmd(command)
    if not command or command == "" then
        return {success = false, error = "No command provided"}
    end
    
    -- Generate turtle-side command ID
    local cmdId = "cmd_" .. os.epoch("utc") .. "_" .. math.random(1000, 9999)
    
    ultron.debugPrint("Processing cmd [" .. cmdId .. "]: " .. command)
    
    -- Create turtle's command structure
    local cmdInfo = {
        id = cmdId,
        command = command,
        startTime = os.epoch("utc"),
        status = "executing"
    }
    
    ultron.pendingCommands[cmdId] = cmdInfo
    
    -- Execute command
    local cmdExec, err = load(command, nil, "t", _ENV)
    if cmdExec then
        print("[cmd:in] = " .. command)
        local success, result = pcall(cmdExec)
        
        -- Create turtle's result structure
        local cmdResult = {
            id = cmdId,
            command = command,
            success = success,
            result = result,
            error = success and nil or result,
            startTime = cmdInfo.startTime,
            endTime = os.epoch("utc"),
            executionTime = os.epoch("utc") - cmdInfo.startTime
        }
        
        print("[cmd:out] = " .. textutils.serialize(cmdResult.result))
        
        -- Clean up
        ultron.pendingCommands[cmdId] = nil
        
        return cmdResult
    else
        local cmdResult = {
            id = cmdId,
            command = command,
            success = false,
            error = "Failed to compile command: " .. (err or "unknown error"),
            startTime = cmdInfo.startTime,
            endTime = os.epoch("utc")
        }
        
        ultron.pendingCommands[cmdId] = nil
        return cmdResult
    end
end

-- Modified command queue processing
local function processCmdQueue()
    while true do
        sleep()
        if ultron.cmd and ultron.cmd ~= "" then
            local result = ultron.processCmd(ultron.cmd)
            
            -- Store result in turtle's data structure
            ultron.data.cmdResult = result
            
            -- Send result back to server immediately
            local resultMessage = {
                type = "command_result", 
                data = result,
                turtle_id = os.getComputerID()
            }
            
            ultron.ws("send", textutils.serializeJSON(resultMessage))
            
            -- Clear processed command
            ultron.cmd = ""
        end
    end
end
```

### Benefits of Collaborative Structure Design:

1. **✅ Clear separation of concerns**: Server handles protocol, turtle handles content
2. **✅ Consistent with existing pattern**: Follows current sight/cmdResult/misc flexibility
3. **✅ Structured flexibility**: Core fields ensure compatibility, flexible fields enable evolution
4. **✅ Backwards compatible**: Existing flexible field usage preserved
5. **✅ Extensible**: Both sides can evolve within their domains
6. **✅ Type safety where needed**: Core fields are typed, flexible fields are open

### Design Principles:

1. **Server defines**: Communication protocol, core required fields, connection management
2. **Turtle defines**: Content format for flexible fields, execution details, custom features
3. **Shared understanding**: Both sides agree on field purposes and basic structure
4. **Intentional flexibility**: Specific fields (sight, cmdResult, misc) designed to be adaptable

### Command Flow with Collaborative Design:

```
1. Server receives POST /api/turtle/{id} with command
2. Server creates CommandRequest with requestId
3. Server sends to turtle via websocket
4. Turtle processes command and creates turtle-defined result
5. Turtle sends CommandResponse with server-required fields + turtle-defined content
6. Server extracts success/failure for HTTP response
7. Server returns turtle's full result content to client
```

This maintains the existing balance where:
- **Server ensures**: Reliable communication and basic compatibility
- **Turtle controls**: Rich data content and feature evolution

### Tradeoffs:

1. **📝 Turtle must be connected**: Commands fail if turtle offline
2. **📝 Single command at a time**: No queuing (but simpler and more reliable)
3. **📝 Request timeout**: Long-running commands may timeout

### Enhanced Websocket Handler (Server):

```go
func TurtleHandleWs(w http.ResponseWriter, r *http.Request) {
    c, err := Upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Print("upgrade:", err)
        return
    }
    defer c.Close()
    
    var turtleID int
    
    for {
        mt, message, err := c.ReadMessage()
        if err != nil {
            log.Println("read:", err)
            break
        }
        
        // Try to parse as structured message first
        var messageObj struct {
            Type     string      `json:"type"`
            Data     interface{} `json:"data"`
            TurtleID int         `json:"turtle_id"`
        }
        
        if json.Unmarshal(message, &messageObj) == nil && messageObj.Type != "" {
            // Handle structured messages
            switch messageObj.Type {
            case "command_result":
                handleCommandResult(messageObj.Data, messageObj.TurtleID)
            case "turtle_update":
                handleTurtleUpdate(messageObj.Data, messageObj.TurtleID)
            default:
                log.Printf("Unknown message type: %s", messageObj.Type)
            }
        } else {
            // Handle legacy turtle data (existing behavior)
            handleLegacyTurtleData(message, &turtleID)
        }
        
        // Register/update turtle connection
        if turtleID > 0 {
            connections.mutex.Lock()
            connections.turtleConnections[turtleID] = c
            connections.mutex.Unlock()
        }
    }
    
    // Clean up connection on disconnect
    if turtleID > 0 {
        connections.mutex.Lock()
        delete(connections.turtleConnections, turtleID)
        connections.mutex.Unlock()
    }
}

func handleCommandResult(data interface{}, turtleID int) {
    // Extract result and notify any waiting POST requests
    // The turtle has already structured the result data
    log.Printf("Command result from turtle %d: %+v", turtleID, data)
    
    // Notify waiting synchronous requests
    notifyWaitingRequest(data)
}
```
```go
type Command struct {
    ID      string `json:"id"`      // Unique command identifier
    Command string `json:"command"` // The actual command to execute
    Result  string `json:"result"`  // Command execution result
    Status  string `json:"status"`  // "pending", "executing", "completed", "failed"
    Created int64  `json:"created"` // Timestamp when command was created
}

type Turtle struct {
    // ... existing fields ...
    CmdQueue    []Command `json:"cmdQueue"`    // Change from []string to []Command
    PendingCmds map[string]*Command `json:"-"` // Track pending commands by ID
}
```

#### 1.2 Modify POST Handler
```go
func TurtleHandle(w http.ResponseWriter, r *http.Request) {
    // ... existing code ...
    if r.Method == "POST" {
        // Generate unique command ID
        cmdID := generateCommandID()
        
        // Create command object
        cmd := Command{
            ID:      cmdID,
            Command: commandString,
            Status:  "pending",
            Created: time.Now().Unix(),
        }
        
        // Add to queue and pending map
        Turtles[pos].CmdQueue = append(Turtles[pos].CmdQueue, cmd)
        Turtles[pos].PendingCmds[cmdID] = &cmd
        
        // Wait for result with timeout
        result := waitForCommandResult(cmdID, 30*time.Second)
        
        // Return result in response
        ReturnData(w, result)
    }
}
```

### Phase 2: Lua-Side Command Processing Updates

#### 2.1 Update Command Reception
```lua
-- In ultron.recieveOrders()
function ultron.recieveOrders()
    while true do
        sleep()
        if not wsError then
            local data = ultron.ws("receive")
            if data then
                local cmdObj = textutils.unserializeJSON(data)
                if cmdObj and cmdObj.id and cmdObj.command then
                    ultron.debugPrint("Received command: " .. cmdObj.id)
                    -- Add to local command queue with ID tracking
                    table.insert(ultron.cmdQueue, cmdObj)
                end
            end
        end
    end
end
```

#### 2.2 Update Command Processing
```lua
-- In processCmdQueue()
local function processCmdQueue()
    while true do
        sleep()
        if #ultron.cmdQueue > 0 then
            local cmdObj = table.remove(ultron.cmdQueue, 1)
            local result = ultron.processCmd(cmdObj.command)
            
            -- Create result object with command ID
            local cmdResult = {
                id = cmdObj.id,
                command = cmdObj.command,
                result = result,
                status = result and "completed" or "failed",
                timestamp = os.epoch("utc")
            }
            
            -- Send result back immediately
            ultron.ws("send", textutils.serializeJSON({
                type = "command_result",
                data = cmdResult
            }))
        end
    end
end
```

### Phase 3: Server-Side Result Handling

#### 3.1 Result Processing in Websocket Handler
```go
func TurtleHandleWs(w http.ResponseWriter, r *http.Request) {
    // ... existing code ...
    for {
        mt, message, err := c.ReadMessage()
        // ... error handling ...
        
        var messageObj struct {
            Type string      `json:"type"`
            Data interface{} `json:"data"`
        }
        
        if json.Unmarshal(message, &messageObj) == nil {
            switch messageObj.Type {
            case "command_result":
                handleCommandResult(messageObj.Data)
            case "turtle_update":
                handleTurtleUpdate(messageObj.Data)
            default:
                // Handle as legacy turtle data
                handleLegacyTurtleData(message)
            }
        }
    }
}

func handleCommandResult(data interface{}) {
    // Extract command ID and result
    // Notify waiting POST handler
    // Remove from pending commands
}
```

#### 3.2 Result Waiting Mechanism
```go
var commandResults = make(map[string]chan Command)
var commandMutex sync.RWMutex

func waitForCommandResult(cmdID string, timeout time.Duration) Command {
    // Create result channel
    resultChan := make(chan Command, 1)
    
    commandMutex.Lock()
    commandResults[cmdID] = resultChan
    commandMutex.Unlock()
    
    // Wait for result or timeout
    select {
    case result := <-resultChan:
        return result
    case <-time.After(timeout):
        return Command{ID: cmdID, Status: "timeout"}
    }
}

func notifyCommandResult(result Command) {
    commandMutex.Lock()
    defer commandMutex.Unlock()
    
    if resultChan, exists := commandResults[result.ID]; exists {
        resultChan <- result
        delete(commandResults, result.ID)
        close(resultChan)
    }
}
```

## Implementation Order (Revised)

1. **Add connection tracking** - Track active turtle websocket connections
2. **Implement synchronous command execution** - Direct websocket communication
3. **Update websocket handler** - Handle immediate command responses
4. **Update Lua command processing** - Support immediate response mode
5. **Add timeout and error handling** - Graceful failure modes
6. **Test thoroughly** - Verify server restart resilience
7. **Maintain backwards compatibility** - Keep existing queue-based system as fallback

## Benefits (Revised)

1. **✅ Stateless design** - Survives server restarts
2. **✅ Immediate feedback** - POST requests return actual command results
3. **✅ Simple architecture** - No complex persistence requirements
4. **✅ Fast failure** - Immediate error if turtle disconnected
5. **✅ Foundation for websockets** - Direct connection management helps with GUI clients

## Alternative: Hybrid Approach

For maximum compatibility, we could implement both modes:

```go
// Header: X-Execution-Mode: sync|async
if r.Header.Get("X-Execution-Mode") == "sync" {
    // Use synchronous execution (new behavior)
    result := executeCommandSync(turtleConn, command, timeout)
    ReturnData(w, result)
} else {
    // Use queue-based execution (existing behavior)
    Turtles[pos].CmdQueue = append(Turtles[pos].CmdQueue, command)
    ReturnData(w, map[string]string{"status": "queued"})
}
```

This preserves existing client behavior while enabling new synchronous mode for clients that need immediate results.

## Testing Strategy

1. Test with simple commands (`print("hello")`)
2. Test with commands that return values (`turtle.forward()`)
3. Test timeout scenarios (turtle disconnected)
4. Test concurrent commands from multiple clients
5. Verify existing HTTP GET endpoints still work
