package module

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"plugin"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

func LoadModules(r *mux.Router, moduleDir string) *mux.Router {
	// check for module directory, if not found, create it
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		os.Mkdir(moduleDir, 0755)
		fmt.Println("Module directory created: " + moduleDir)
	} else {
		fmt.Println("Module directory found: " + moduleDir)
	}

	// load modules
	files, err := listFiles(moduleDir, "")
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Found", len(files), "modules")
	}
	for _, file := range files {
		// if file is not a .so file, skip it
		if !strings.HasSuffix(file.Name(), "ult.so") || file.IsDir() {
			continue
		}
		fmt.Println("Loading module: " + file.Name())
		module, err := plugin.Open(moduleDir + "/" + file.Name())
		if err != nil {
			panic(err)
		}

		// get the module name
		moduleName, err := module.Lookup("Name")
		if err != nil {
			panic(err)
		}
		name, _ := moduleName.(func() string)
		moduleVersion, _ := module.Lookup("Version")
		fmt.Println("[Module] ["+name()+"]", "Version: "+moduleVersion.(func() string)())
		moduleDesc, _ := module.Lookup("Desc")
		fmt.Println("[Module] ["+name()+"]", "Description: "+moduleDesc.(func() string)())
		moduleUsage, _ := module.Lookup("Usage")
		fmt.Println("[Module] ["+name()+"]", "Usage: "+moduleUsage.(func() string)())
		moduleInit, _ := module.Lookup("Init")
		moduleInit.(func())()
		moduleHandleWs, err := module.Lookup("HandleWs")
		if err != nil {
			fmt.Println("Module websocket handler not found")
		} else {
			r.HandleFunc("/api/"+name()+"ws", moduleHandleWs.(func(http.ResponseWriter, *http.Request)))
		}
		moduleHandle, _ := module.Lookup("Handle")
		r.HandleFunc("/api/"+name(), moduleHandle.(func(http.ResponseWriter, *http.Request)))
	}
	return r
}
func listFiles(dir, pattern string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	filteredFiles := []os.FileInfo{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fmt.Println("Found file: " + file.Name())
		matched, err := regexp.MatchString(pattern, file.Name())
		if err != nil {
			return nil, err
		}
		if matched {
			filteredFiles = append(filteredFiles, file)
		}
	}
	return filteredFiles, nil
}
