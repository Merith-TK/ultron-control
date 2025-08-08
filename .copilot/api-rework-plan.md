# API System Analysis & Rework Plan

**Date**: August 8, 2025  
**Focus**: Websocket communication improvement for real-time data relay

## Current System Analysis

### Current Architecture Problems

#### 1. **Monolithic Loop Structure in turtle.lua**
```lua
local function main()
    parallel.waitForAll(apiLoop, ultron.recieveOrders, processCmdQueue)
end
```

**Issues:**
- Single `apiLoop()` controls all communication timing
- Data updates are tied to `ultron.config.api.delay` (0.5s default)
- No event-driven communication - only periodic polling
- GUI clients can't get real-time turtle data updates

#### 2. **Server-Side Websocket Limitations**
```go
func TurtleHandleWs(w http.ResponseWriter, r *http.Request) {
    // Only handles turtle->server communication
    // No mechanism to broadcast to GUI clients
    // Single websocket connection per turtle
}
```

**Issues:**
- Server only maintains one websocket per turtle
- No client registry or broadcasting system
- GUI clients can only get data via HTTP polling
- No real-time data relay capability

#### 3. **Data Flow Problems**
```
Current: Turtle -> Server (via WS) -> GUI Client (via HTTP GET polling)
Needed:  Turtle -> Server (via WS) -> GUI Clients (via WS broadcast)
```

## Proposed Rework Architecture

### 1. **Event-Driven Turtle Client (Lua)**

#### New Structure:
```lua
-- Separate concerns into independent coroutines
local function main()
    parallel.waitForAll(
        dataCollector,    -- Collect turtle state on events
        wsCommListener,   -- Listen for server commands
        wsSender,         -- Send data when needed
        cmdProcessor      -- Process command queue
    )
end
```

#### Key Changes:
- **Event-driven data collection**: Trigger on inventory changes, movement, etc.
- **Immediate data transmission**: Send updates when state changes, not on timer
- **Separate command processing**: Independent of data collection
- **Heartbeat system**: Periodic keepalive separate from data updates

### 2. **Multi-Client Websocket Server (Go)**

#### New Components:
```go
type ClientManager struct {
    turtleClients map[int]*websocket.Conn    // Turtle connections
    guiClients    map[string]*websocket.Conn // GUI client connections
    broadcast     chan BroadcastMessage      // Channel for broadcasting
}

type BroadcastMessage struct {
    Type     string      `json:"type"`     // "turtle_update", "command_result", etc.
    TurtleID int         `json:"turtleId"`
    Data     interface{} `json:"data"`
}
```

#### Key Features:
- **Client registration**: Separate endpoints for turtles vs GUI clients
- **Real-time broadcasting**: When turtle sends data, immediately broadcast to all GUI clients
- **Message typing**: Different message types for different events
- **Connection management**: Handle client disconnections gracefully

### 3. **New API Endpoints Structure**

```
/api/turtle/ws           # Turtle websocket endpoint (existing)
/api/gui/ws             # GUI client websocket endpoint (NEW)
/api/turtle/{id}/cmd    # Send command to specific turtle (existing but improved)
```

## Implementation Plan

### Phase 1: Server-Side Improvements

1. **Create Client Manager**
   - Implement connection registry
   - Add broadcasting system
   - Handle multiple client types

2. **Add GUI Websocket Endpoint**
   - New endpoint for GUI clients
   - Client authentication/identification
   - Subscribe to specific turtle updates

3. **Modify Turtle Websocket Handler**
   - Add broadcasting when turtle data received
   - Maintain backwards compatibility

### Phase 2: Lua Client Refactor

1. **Event-Driven Data Collection**
   ```lua
   local function dataCollector()
       while true do
           local event, data = os.pullEvent()
           if event == "turtle_inventory" or 
              event == "turtle_moved" or 
              event == "gps_located" then
               updateTurtleData()
               triggerDataSend()
           end
       end
   end
   ```

2. **Separate Communication Threads**
   ```lua
   local function wsSender()
       while true do
           local shouldSend = os.pullEvent("data_ready")
           sendTurtleData()
       end
   end
   ```

3. **Command Processing Independence**
   ```lua
   local function cmdProcessor()
       while true do
           if #ultron.cmdQueue > 0 then
               processCommand(table.remove(ultron.cmdQueue, 1))
           end
           sleep(0.1)
       end
   end
   ```

### Phase 3: Message Protocol

#### Message Types:
```json
{
  "type": "turtle_update",
  "turtleId": 123,
  "data": { /* full turtle data */ }
}

{
  "type": "turtle_connected",
  "turtleId": 123,
  "data": { "name": "Miner-01" }
}

{
  "type": "command_result",
  "turtleId": 123,
  "data": { "success": true, "result": "..." }
}
```

## Benefits of New Architecture

1. **Real-time Updates**: GUI clients get immediate turtle data
2. **Better Separation of Concerns**: Each component has single responsibility
3. **Scalability**: Multiple GUI clients can connect simultaneously
4. **Event-driven**: More responsive and efficient
5. **Maintainability**: Clearer code structure and data flows

## Migration Strategy

1. **Backwards Compatibility**: Keep existing HTTP API working
2. **Gradual Migration**: Implement new features alongside old
3. **Feature Flags**: Allow switching between old/new behavior
4. **Testing**: Maintain existing functionality while adding new

Would you like me to start implementing any specific part of this rework plan?
