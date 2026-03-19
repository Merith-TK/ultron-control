package api

import (
	"log"

	"github.com/gorilla/mux"
)

// Just a place to easily load the Modules

func InitModules(m *mux.Router) {
	// Setup Turtle API
	//create api for /api/turtle with argument for id
	m.HandleFunc("/api/turtle", TurtleHandle)
	m.HandleFunc("/api/turtle/{id}", TurtleHandle)
	m.HandleFunc("/api/turtle/{id}/{action}", TurtleHandle)
	m.HandleFunc("/api/turtle/{id}/{action}/{action2}", TurtleHandle)
	log.Println("[Module] Loaded Turtle")

	// Setup Texture Provider
	log.Println("[Texture] Loading Texture Module")
	m.HandleFunc("/api/texture", TextureHandle)
	m.HandleFunc("/api/texture/{asset}/{modid}/{texture}", TextureHandle)
	ExtractResources()
	log.Println("[Module] Loaded Textures")

	// MCP server (streamable HTTP transport, JSON-RPC 2.0)
	m.Handle("/mcp", MCPHandler)
	log.Println("[Module] Loaded MCP")

	// Ensure doc repos are cloned in background
	go EnsureDocRepos()

	// Add Modules Below this Line
}
