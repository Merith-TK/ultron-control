package api

import (
	"io/fs"
	"log"
	"path/filepath"

	"embed"

	"github.com/gorilla/mux"
)

// Just a place to easily load the Modules

func InitModules(m *mux.Router, dataDir string, docsFS embed.FS) {
	// World map SQLite database
	if err := InitWorldMap(filepath.Join(dataDir, "worldmap.db")); err != nil {
		log.Println("[WorldMap] Failed to initialize:", err)
	}

	// Setup Turtle API
	//create api for /api/turtle with argument for id
	m.HandleFunc("/api/turtle", TurtleHandle)
	m.HandleFunc("/api/turtle/{id}", TurtleHandle)
	m.HandleFunc("/api/turtle/{id}/{action}", TurtleHandle)
	m.HandleFunc("/api/turtle/{id}/{action}/{action2}", TurtleHandle)
	log.Println("[Module] Loaded Turtle")

	// Setup Texture Provider
	log.Println("[Texture] Loading Texture Module")
	TextureWorkdir = dataDir
	m.HandleFunc("/api/texture", TextureHandle)
	m.HandleFunc("/api/texture/{asset}/{modid}/{texture}", TextureHandle)
	ExtractResources()
	log.Println("[Module] Loaded Textures")

	// Extract embedded docs to dataDir/docs/ (skip existing files).
	DocsDir = filepath.Join(dataDir, "docs")
	subFS, err := fs.Sub(docsFS, "mcp/docs")
	if err != nil {
		log.Println("[Docs] Failed to sub embedded FS:", err)
	} else if err := extractEmbedded(subFS, ".", DocsDir); err != nil {
		log.Println("[Docs] Failed to extract embedded docs:", err)
	}

	// MCP server (streamable HTTP transport, JSON-RPC 2.0)
	m.Handle("/mcp", MCPHandler)
	log.Println("[Module] Loaded MCP")

	// Ensure doc repos are cloned in background
	go EnsureDocRepos()

	// Add Modules Below this Line
}
