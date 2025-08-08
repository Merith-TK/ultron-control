# Ultron Control API Documentation

## Overview

The Ultron Control API provides both synchronous and asynchronous command execution for ComputerCraft turtles. The API supports real-time command execution with immediate results (synchronous mode) and traditional queue-based execution (asynchronous mode).

**Default Execution Mode**: Synchronous (real-time execution)

## Base URL
```
http://localhost:3300/api/turtle
```

## Authentication
No authentication required for local development.

## Execution Modes

### Synchronous Mode (Default)
- **Description**: Commands execute immediately and return results in real-time
- **Timeout**: 30 seconds
- **Use Case**: Interactive commands, data retrieval, immediate feedback
- **Response**: Complete command result with execution metadata

### Asynchronous Mode (Legacy)
- **Description**: Commands are queued and executed when turtle polls for new commands
- **Timeout**: No timeout (queue-based)
- **Use Case**: Batch operations, long-running tasks
- **Response**: Queue confirmation only

## API Endpoints

### GET /api/turtle
Returns data for all connected turtles.

**Response Example**:
```json
[
  {
    "name": "0",
    "id": 0,
    "inventory": [...],
    "selectedSlot": 1,
    "pos": {"y": 0, "x": 0, "z": 0, "r": 0, "rname": "north"},
    "fuel": {"current": 0, "max": 100000},
    "sight": {"up": "", "down": "", "front": ""},
    "cmdResult": {...},
    "cmdQueue": [],
    "misc": "",
    "heartbeat": 1754670195506
  }
]
```

### GET /api/turtle/{id}
Returns data for a specific turtle.

**Parameters**:
- `id` (path): Turtle ID number, or "debug" for debug turtle

**Response**: Single turtle object (same structure as array item above)

### POST /api/turtle/{id}
Execute commands on a specific turtle.

**Parameters**:
- `id` (path): Turtle ID number

**Headers**:
- `Content-Type`: `application/json` or `text/plain`
- `X-Execution-Mode` (optional): `sync` (default) or `async`

**Request Body**:
- JSON Array: `["command1", "command2"]` for multiple commands
- Plain Text: `command` for single command

**Response Format**:

#### Synchronous Mode (Default)
```json
{
  "success": true,
  "result": {
    "command": "return turtle.getFuelLevel()",
    "startTime": 1754670217197,
    "endTime": 1754670217197,
    "executionTime": 0,
    "success": true,
    "result": 0,
    "error": "0"
  },
  "mode": "synchronous"
}
```

#### Asynchronous Mode
```json
{
  "success": true,
  "message": "Commands queued successfully",
  "mode": "asynchronous",
  "count": 1
}
```

### WebSocket Endpoint
- **Path**: `/api/turtle/ws`
- **Purpose**: Internal turtle communication
- **Usage**: Not for external clients

## Command Examples

### PowerShell Examples

#### Basic Synchronous Command (Default)
```powershell
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" `
  -Method POST `
  -ContentType "text/plain" `
  -Body 'return turtle.getFuelLevel()'
```

#### Explicit Synchronous Mode
```powershell
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" `
  -Method POST `
  -ContentType "text/plain" `
  -Headers @{"X-Execution-Mode"="sync"} `
  -Body 'return "Hello World"'
```

#### Asynchronous Mode (Legacy)
```powershell
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" `
  -Method POST `
  -ContentType "text/plain" `
  -Headers @{"X-Execution-Mode"="async"} `
  -Body 'print("Background task")'
```

#### Multiple Commands (Async Only)
```powershell
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" `
  -Method POST `
  -ContentType "application/json" `
  -Headers @{"X-Execution-Mode"="async"} `
  -Body '["turtle.forward()", "turtle.turnRight()", "turtle.forward()"]'
```

### cURL Examples

#### Basic Synchronous Command (Default)
```bash
curl -X POST http://localhost:3300/api/turtle/0 \
  -H "Content-Type: text/plain" \
  -d "return turtle.getFuelLevel()"
```

#### Asynchronous Mode
```bash
curl -X POST http://localhost:3300/api/turtle/0 \
  -H "Content-Type: text/plain" \
  -H "X-Execution-Mode: async" \
  -d "print('Background task')"
```

## Error Handling

### Common Error Responses

#### Turtle Not Connected
```json
{
  "error": {
    "code": "503",
    "message": "Turtle not connected"
  }
}
```

#### Turtle Not Found
```json
{
  "error": {
    "code": "503", 
    "message": "Turtle has not been added yet"
  }
}
```

#### Invalid Command Syntax
```json
{
  "success": true,
  "result": {
    "command": "invalid_function()",
    "success": false,
    "error": "[string \"invalid_function()\"]:1: attempt to call global 'invalid_function' (a nil value)",
    "executionTime": 0
  },
  "mode": "synchronous"
}
```

#### Multiple Commands in Sync Mode
```json
{
  "error": {
    "code": "400",
    "message": "Synchronous mode only supports single command"
  }
}
```

## Best Practices

### When to Use Synchronous Mode (Default)
- ✅ Data retrieval commands (`return turtle.getFuelLevel()`)
- ✅ Status checks (`return turtle.getSelectedSlot()`)
- ✅ Interactive operations requiring immediate feedback
- ✅ Single command execution
- ✅ Error handling where you need immediate results

### When to Use Asynchronous Mode
- ✅ Long-running operations (mining, building)
- ✅ Multiple command sequences
- ✅ Background tasks that don't require immediate results
- ✅ Batch operations
- ✅ Fire-and-forget commands

### Performance Considerations
- **Synchronous**: 30-second timeout limit per command
- **Asynchronous**: No timeout, commands executed in sequence
- **Connection**: Requires active WebSocket connection for sync mode
- **Throughput**: Async mode better for bulk operations

## Implementation Details

### Architecture
- **Server**: Go with Gorilla WebSocket
- **Client**: ComputerCraft Lua scripts
- **Protocol**: Hybrid sync/async with structured messaging
- **Connection**: WebSocket-based real-time communication

### Connection Tracking
- Active connections tracked per turtle ID
- Automatic cleanup on disconnect
- Enhanced debugging and logging
- Support for turtle ID 0 and higher

### Backwards Compatibility
- Legacy async behavior preserved
- Existing scripts continue to work
- Gradual migration path to sync mode
- No breaking changes to existing APIs

## Troubleshooting

### Sync Commands Fail
1. Verify turtle is connected: `GET /api/turtle/{id}`
2. Check WebSocket connection in server logs
3. Ensure turtle is running updated scripts
4. Try async mode as fallback

### Connection Issues
1. Restart turtle program in ComputerCraft
2. Check server is running on port 3300
3. Verify network connectivity
4. Check server logs for WebSocket errors

### Command Timeouts
1. Use async mode for long operations
2. Break complex commands into smaller parts
3. Check for infinite loops in command
4. Verify turtle has sufficient resources
