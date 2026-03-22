package api

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"embed"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// extractEmbedded walks srcDir inside src and copies files to destDir,
// skipping any file that already exists on disk (preserves customisation).
func extractEmbedded(src fs.FS, srcDir, destDir string) error {
	return fs.WalkDir(src, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(srcDir, path)
		dest := filepath.Join(destDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		if _, err := os.Stat(dest); err == nil {
			return nil // already exists, skip
		}
		data, err := fs.ReadFile(src, path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0644)
	})
}

func CreateApiServer(domain string, port int, dataDir string, staticFS embed.FS, docsFS embed.FS) {
	// // create webserver on port 3300
	r := mux.NewRouter()

	// handle /
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// return all api routes
		w.Write([]byte("Welcome to the Ultron API!"))
	})

	InitModules(r, dataDir, docsFS)

	// Extract embedded Lua files to dataDir/lua/ (skip existing) then serve from disk.
	luaDir := filepath.Join(dataDir, "lua")
	if err := extractEmbedded(staticFS, "static", luaDir); err != nil {
		log.Println("[Lua] Failed to extract embedded files:", err)
	}
	r.PathPrefix("/api/static/").Handler(http.StripPrefix("/api/static/", http.FileServer(http.Dir(luaDir))))

	//handle global api on /api/v1
	r.HandleFunc("/api", handleGlobalApi)

	// Documentation endpoints
	r.HandleFunc("/api/docs/list", ListDocs)
	r.HandleFunc("/api/docs/get", GetDocs)
	r.HandleFunc("/api/docs/manifest", GetManifest)

	// if page not found, return server error
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ReturnError(w, http.StatusNotImplemented, "Server Error: Check for trailing / in url, or verify against documentation of API")
	})

	// start webserver on config.Port
	portstr := strconv.Itoa(port)
	http.ListenAndServe(domain+":"+portstr, r)
}

// ReturnError returns an error to the client with the specified status code and message
func ReturnError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	ReturnData(w, map[string]interface{}{"error": map[string]interface{}{"code": strconv.Itoa(code), "message": message}})
}

// ReturnData returns data as json to the client
func ReturnData(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func ReturnDataRaw(w http.ResponseWriter, data []byte, headers map[string]string) {
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.Write(data)
}

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Connection tracking for synchronous command execution
type ActiveConnections struct {
	turtleConnections map[int]*websocket.Conn // turtle ID -> websocket connection
	mutex             sync.RWMutex            // Protect concurrent access
}

// Per-turtle command serialization
type TurtleCommandMutex struct {
	mutex sync.Mutex
}

// Global connection manager
var connections = &ActiveConnections{
	turtleConnections: make(map[int]*websocket.Conn),
}

// Global command mutex manager for per-turtle serialization
var turtleCommandMutexes = make(map[int]*TurtleCommandMutex)
var commandMutexLock sync.RWMutex

// Get or create a command mutex for a specific turtle
func getTurtleCommandMutex(turtleID int) *TurtleCommandMutex {
	commandMutexLock.RLock()
	mutex, exists := turtleCommandMutexes[turtleID]
	commandMutexLock.RUnlock()

	if !exists {
		commandMutexLock.Lock()
		// Check again in case another goroutine created it
		mutex, exists = turtleCommandMutexes[turtleID]
		if !exists {
			mutex = &TurtleCommandMutex{}
			turtleCommandMutexes[turtleID] = mutex
			log.Printf("[Command Mutex] Created command serialization for turtle %d", turtleID)
		}
		commandMutexLock.Unlock()
	}

	return mutex
}

// Command request/response structures
type CommandRequest struct {
	Command   string `json:"command"`
	RequestID string `json:"requestId"`
}

type CommandResponse struct {
	RequestID string      `json:"requestId"`
	Result    interface{} `json:"result"`
	Success   bool        `json:"success"`
}

// Pending request tracking for synchronous execution
var pendingRequests = make(map[string]chan CommandResponse)
var pendingMutex sync.RWMutex

// Connection management functions
func registerTurtleConnection(turtleID int, conn *websocket.Conn) {
	connections.mutex.Lock()
	defer connections.mutex.Unlock()
	connections.turtleConnections[turtleID] = conn
}

func unregisterTurtleConnection(turtleID int) {
	connections.mutex.Lock()
	defer connections.mutex.Unlock()
	delete(connections.turtleConnections, turtleID)
}

func getTurtleConnection(turtleID int) (*websocket.Conn, bool) {
	connections.mutex.RLock()
	defer connections.mutex.RUnlock()
	conn, exists := connections.turtleConnections[turtleID]
	return conn, exists
}

// Generate unique request ID
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// Execute command synchronously with timeout and per-turtle serialization
func executeCommandSync(turtleID int, command string, timeout time.Duration) (interface{}, error) {
	// Get per-turtle command mutex to ensure FIFO execution
	commandMutex := getTurtleCommandMutex(turtleID)
	commandMutex.mutex.Lock()
	defer commandMutex.mutex.Unlock()

	log.Printf("[Command Serialization] Acquired lock for turtle %d", turtleID)

	// Get turtle connection
	conn, exists := getTurtleConnection(turtleID)
	if !exists {
		return nil, &CommandError{Message: "Turtle not connected"}
	}

	// Generate request ID
	requestID := generateRequestID()

	// Create response channel
	responseChan := make(chan CommandResponse, 1)

	// Register pending request
	pendingMutex.Lock()
	pendingRequests[requestID] = responseChan
	pendingMutex.Unlock()

	// Clean up on exit
	defer func() {
		pendingMutex.Lock()
		delete(pendingRequests, requestID)
		pendingMutex.Unlock()
		close(responseChan)
		log.Printf("[Command Serialization] Released lock for turtle %d", turtleID)
	}()

	// Send command request
	request := CommandRequest{
		Command:   command,
		RequestID: requestID,
	}

	err := conn.WriteJSON(request)
	if err != nil {
		return nil, &CommandError{Message: "Failed to send command: " + err.Error()}
	}

	// Wait for response or timeout
	select {
	case response := <-responseChan:
		if response.Success {
			return response.Result, nil
		}
		return response.Result, &CommandError{Message: "Command execution failed"}
	case <-time.After(timeout):
		return nil, &CommandError{Message: "Command timeout"}
	}
}

// Handle command response from turtle
func handleCommandResponse(response CommandResponse) {
	pendingMutex.RLock()
	responseChan, exists := pendingRequests[response.RequestID]
	pendingMutex.RUnlock()

	if exists {
		select {
		case responseChan <- response:
			// Response delivered
		default:
			// Channel full or closed, ignore
		}
	}
}

// Custom error type for command execution
type CommandError struct {
	Message string
}

func (e *CommandError) Error() string {
	return e.Message
}
