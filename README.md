# Ultron Control

A ComputerCraft turtle control system with real-time command execution and comprehensive API.

## 🚀 Quick Start

### Start the Server
```bash
go run main.go
# Server runs on http://localhost:3300
```

### Execute Commands (Synchronous - Default)
```powershell
# Get turtle status
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method GET

# Execute command with immediate result (default behavior)
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'return turtle.getFuelLevel()'

# Queue command for background execution (legacy mode)
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Headers @{"X-Execution-Mode"="async"} -Body 'print("Background task")'
```

## 🎯 Key Features

### ⚡ Real-Time Command Execution (Default)
- **Synchronous execution**: Commands execute immediately with real-time results
- **30-second timeout**: Automatic timeout protection
- **Error handling**: Immediate error feedback and debugging
- **Interactive**: Perfect for data retrieval and interactive operations

### 📋 Queue-Based Execution (Legacy)
- **Asynchronous execution**: Commands queue for background processing
- **Batch operations**: Multiple commands in sequence
- **Long-running tasks**: No timeout limitations
- **Fire-and-forget**: Ideal for automation and background tasks

### 🔗 WebSocket Connection Tracking
- **Enhanced connection management**: Real-time turtle connection monitoring
- **Automatic cleanup**: Proper disconnection handling
- **Debug logging**: Comprehensive WebSocket event logging
- **Turtle ID 0 support**: Full support for all turtle IDs including 0

## 📖 API Documentation

### Execution Modes

| Mode | Default | Header Required | Use Case | Response Time |
|------|---------|----------------|----------|---------------|
| **Synchronous** | ✅ Yes | None (or `X-Execution-Mode: sync`) | Data retrieval, interactive commands | Immediate |
| **Asynchronous** | ❌ No | `X-Execution-Mode: async` | Batch operations, long tasks | Queue confirmation |

### Basic API Endpoints

- `GET /api/turtle` - List all turtles
- `GET /api/turtle/{id}` - Get turtle data
- `POST /api/turtle/{id}` - Execute command (sync by default)
- `GET /api/turtle/usage` - API usage documentation

### Command Examples

#### Synchronous Commands (Default)
```powershell
# Get fuel level (immediate result)
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'return turtle.getFuelLevel()'

# Get inventory (immediate result)
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'return turtle.list()'

# Move turtle (immediate success/failure)
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'return turtle.forward()'
```

#### Asynchronous Commands (Legacy)
```powershell
# Queue multiple commands
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "application/json" -Headers @{"X-Execution-Mode"="async"} -Body '["turtle.forward()", "turtle.turnRight()", "turtle.forward()"]'

# Background task
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Headers @{"X-Execution-Mode"="async"} -Body 'for i=1,10 do turtle.forward() end'
```

## 🏗️ Architecture

### Server Components
- **Go HTTP Server**: RESTful API with Gorilla mux
- **WebSocket Handler**: Real-time turtle communication
- **Connection Tracking**: Active connection management per turtle
- **Command Execution**: Hybrid sync/async execution engine

### Client Components  
- **ComputerCraft Lua**: Enhanced turtle scripts
- **WebSocket Client**: Real-time server communication
- **Command Processing**: Structured command handling
- **Result Formatting**: Comprehensive execution metadata

### Communication Protocol
- **Legacy Messages**: JSON turtle data updates
- **Structured Messages**: Type-based message handling
- **Command Responses**: Request/response correlation
- **Connection Management**: Registration and cleanup

## 📁 Project Structure

```
├── main.go                 # Application entry point
├── config.toml            # Configuration file
├── api/
│   ├── init.go           # Connection management & sync execution
│   ├── turtle.go         # Main turtle API handlers
│   ├── global.go         # Shared data structures
│   └── setup.go          # API route setup
├── static/
│   ├── turtle.lua        # Main turtle program
│   ├── ultron.lua        # Enhanced turtle library
│   ├── pastebin.lua      # Auto-update utility
│   └── startup.lua       # Turtle startup script
└── .copilot/
    ├── api-documentation.md    # Complete API documentation
    ├── test-commands.md        # Test commands and results
    └── session-notes.md        # Development notes
```

## 🧪 Testing

### Automated Tests
```powershell
# Test basic turtle status
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle" -Method GET

# Test default sync execution
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Body 'return "test"'

# Test legacy async execution  
Invoke-RestMethod -Uri "http://localhost:3300/api/turtle/0" -Method POST -ContentType "text/plain" -Headers @{"X-Execution-Mode"="async"} -Body 'print("test")'
```

### Development Tools
- **Air**: Live reload development server
- **PowerShell**: Windows-compatible testing commands
- **VS Code**: Integrated development environment
- **ComputerCraft**: Minecraft mod for turtle execution

## 🔧 Configuration

### Server Configuration (`config.toml`)
```toml
[server]
port = 3300
static_dir = "static"
work_dir = "workdir"
```

### Environment Setup
1. Install Go 1.19+
2. Install ComputerCraft mod
3. Configure turtle scripts
4. Start development server with Air

## 🚀 Deployment

### Development
```bash
# Install Air for live reload
go install github.com/cosmtrek/air@latest

# Start development server
air
```

### Production
```bash
# Build binary
go build -o ultron-control

# Run server
./ultron-control
```

## 🎮 ComputerCraft Setup

### Turtle Installation
```lua
-- Run this in ComputerCraft turtle to install Ultron scripts
shell.run("rm", "startup.lua")
shell.run("pastebin", "get", "XWDutV7S", "startup.lua")
shell.run("startup.lua")
```

### Turtle Requirements
- ComputerCraft mod installed
- Turtle with networking capability
- Connection to server (localhost:3300)

## 📝 Recent Updates

### v2.0 - Synchronous Default
- **⚡ Default Mode Change**: Synchronous execution is now default
- **🔗 Enhanced WebSocket**: Improved connection tracking and cleanup  
- **🐛 Turtle ID 0 Fix**: Full support for turtle ID 0
- **📋 Better Logging**: Comprehensive debug output
- **📚 Documentation**: Complete API documentation and examples

### v1.0 - Foundation
- **🏗️ Core API**: Basic turtle control API
- **🔄 Async Commands**: Queue-based command execution
- **📡 WebSocket Communication**: Real-time turtle communication
- **🎮 ComputerCraft Integration**: Lua script integration

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙋‍♂️ Support

For questions, issues, or contributions:
- Check the [API Documentation](.copilot/api-documentation.md)
- Review [Test Commands](.copilot/test-commands.md)
- Open an issue for bugs or feature requests
