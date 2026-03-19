package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
)

// DocEntry mirrors one entry in mcp/docs/manifest.json
type DocEntry struct {
	Name  string `json:"name"`
	Repo  string `json:"repo"`
	Root  string `json:"root"`
	Local bool   `json:"local,omitempty"`
}

// DocManifest is the top-level structure of mcp/docs/manifest.json
type DocManifest struct {
	Docs []DocEntry `json:"docs"`
}

// ReadManifest reads and parses mcp/docs/manifest.json.
func ReadManifest() (*DocManifest, error) {
	data, err := os.ReadFile(filepath.Join("mcp", "docs", "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m DocManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// EnsureDocRepos checks each entry in the manifest and clones the repo if
// the target directory does not already contain a git repository.
// Runs each clone sequentially; call in a goroutine to avoid blocking startup.
func EnsureDocRepos() {
	manifest, err := ReadManifest()
	if err != nil {
		log.Printf("[Docs] Cannot read manifest: %v", err)
		return
	}

	for _, entry := range manifest.Docs {
		if entry.Local || entry.Repo == "" {
			log.Printf("[Docs] %s is local, skipping clone", entry.Name)
			continue
		}

		dest := filepath.Join("mcp", "docs", entry.Name)

		_, err := git.PlainOpen(dest)
		if err == nil {
			log.Printf("[Docs] %s already cloned, skipping", entry.Name)
			continue
		}
		if err != git.ErrRepositoryNotExists {
			log.Printf("[Docs] Error opening %s: %v", entry.Name, err)
			continue
		}

		log.Printf("[Docs] Cloning %s from %s ...", entry.Name, entry.Repo)
		_, err = git.PlainClone(dest, false, &git.CloneOptions{
			URL:          entry.Repo,
			Depth:        1,
			SingleBranch: true,
			Progress:     os.Stdout,
		})
		if err != nil {
			log.Printf("[Docs] Failed to clone %s: %v", entry.Name, err)
		} else {
			log.Printf("[Docs] Cloned %s OK", entry.Name)
		}
	}
}

// HTTP handlers (used by the REST API endpoints in init.go)

func GetManifest(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(filepath.Join("mcp", "docs", "manifest.json"))
	if err != nil {
		http.Error(w, "Manifest not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func ListDocs(w http.ResponseWriter, r *http.Request) {
	manifest, err := ReadManifest()
	if err != nil {
		http.Error(w, "Failed to read manifest", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(manifest)
}

func GetDocs(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	file := r.URL.Query().Get("file")
	if name == "" || file == "" {
		http.Error(w, "Missing name or file parameter", http.StatusBadRequest)
		return
	}

	manifest, err := ReadManifest()
	if err != nil {
		http.Error(w, "Failed to read manifest", http.StatusInternalServerError)
		return
	}

	var root string
	for _, entry := range manifest.Docs {
		if entry.Name == name {
			root = entry.Root
			break
		}
	}
	if root == "" {
		http.Error(w, "Unknown doc set: "+name, http.StatusNotFound)
		return
	}

	filePath := filepath.Join("mcp", "docs", name, root, file)
	data, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	w.Write(data)
}
