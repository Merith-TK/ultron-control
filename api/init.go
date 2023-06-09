package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func CreateApiServer(domain string, port int, luaFiles string, dataDir string) {
	// // create webserver on port 3300
	r := mux.NewRouter()
	r.Headers("Content-Type", "application/json")

	// handle /
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// return all api routes
		w.Write([]byte("Welcome to the Ultron API!"))
	})

	InitModules(r)

	// Serve Turtle Files
	r.PathPrefix("/api/static/").Handler(http.StripPrefix("/api/static/", http.FileServer(http.Dir(luaFiles))))

	//handle global api on /api/v1
	r.HandleFunc("/api", handleGlobalApi)

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
