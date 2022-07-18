package api

import (
	"io/ioutil"
	"os"
	"plugin"
	"regexp"

	"github.com/gorilla/mux"
)

func loadModules(r *mux.Router, moduleDir string) *mux.Router {
	// check for module directory, if not found, create it
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		os.Mkdir(moduleDir, 0755)
		println("Module directory created: " + moduleDir)
	} else {
		println("Module directory found: " + moduleDir)
	}

	// load modules
	files, err := listFiles(moduleDir, "")
	if err != nil {
		panic(err)
	} else {
		println("Found " + string(len(files)) + " modules")
	}
	for _, file := range files {
		// if file is not a .so file, skip it
		if file.IsDir() {
			continue
		}
		println("Loading module: " + file.Name())
		module, err := plugin.Open(moduleDir + "/" + file.Name())
		if err != nil {
			panic(err)
		}
		initFunc, err := module.Lookup("Init")
		if err != nil {
			panic(err)
		}
		initFunc.(func(*mux.Router))(r)
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
		println("Found file: " + file.Name())
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
