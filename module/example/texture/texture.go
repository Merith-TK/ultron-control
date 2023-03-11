package main

import (
	"log"
	"net/http"
	"ultron/config"

	"github.com/gorilla/mux"
)

// Name is the name of the module, this is used for the module name in the API
// api/<name>
// IF YOU ARE NOT THE ORIGINAL CREATOR OF THE MODULE, PLEASE DO NOT CHANGE THIS
// AS IT WILL BREAK THINGS UNLESS YOU MODIFY ALL ASSOCIATED LUA FILES
func Name() string    { return `texture` } //OP
func Version() string { return `0.0.1` }

// Desc is the description of the module, this is used for the module description when loading
func Desc() string { return `Provides Textures from API` } //OP
// Usage is used to tell the users/developers how this particiular plugin works
func Usage() string {
	return `
/api/texture
	GET: Returns resourcepack.zip
/api/texture/modid
	GET: Returns Error
/api/texture/modid/texture
	GET: Returns texture
`
}
func Init(m *mux.Router) {
	// initial setup for the plugin
	log.Println("[Texture] Loading Texture Module")
	m.HandleFunc("/api/texture", Handle)
	m.HandleFunc("/api/texture/{asset}/{modid}/{texture}", Handle)

	extractResources()

	log.Println("[Texture] Loaded Texture Module")
}

var workdir = config.GetConfig().UltronData

func Handle(w http.ResponseWriter, r *http.Request) {

	// USAGE: /api/texture/{asset}/modid/texture
	// asset: block, item
	// modid: modid
	// texture: texture name

	// register vars
	vars := mux.Vars(r)
	asset := vars["asset"]
	modid := vars["modid"]
	texture := vars["texture"]

	// print request path
	log.Println("[Texture] Requested", r.URL.Path)

	// serve texture
	http.ServeFile(w, r, workdir+"/resourcepack/assets/"+modid+"/textures/"+asset+"/"+texture+".png")
}
