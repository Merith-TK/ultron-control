package api

import (
	"log"
	"net/http"

	"github.com/merith-tk/ultron-control/config"

	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

// Name is the name of the module, this is used for the module name in the API
// api/<name>
// IF YOU ARE NOT THE ORIGINAL CREATOR OF THE MODULE, PLEASE DO NOT CHANGE THIS
// AS IT WILL BREAK THINGS UNLESS YOU MODIFY ALL ASSOCIATED LUA FILES
// func Name() string    { return `texture` } //OP
// func Version() string { return `0.0.1` }

// // Desc is the description of the module, this is used for the module description when loading
// func Desc() string { return `Provides Textures from API` } //OP
// // Usage is used to tell the users/developers how this particiular plugin works
// func Usage() string {
// 	return `
// /api/texture
// 	GET: Returns resourcepack.zip
// /api/texture/modid
// 	GET: Returns Error
// /api/texture/modid/texture
// 	GET: Returns texture
// `
// }

var TextureWorkdir = config.GetConfig().UltronData

func TextureHandle(w http.ResponseWriter, r *http.Request) {

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
	http.ServeFile(w, r, TextureWorkdir+"/resourcepack/assets/"+modid+"/textures/"+asset+"/"+texture+".png")
}

func ExtractResources() {
	log.Println("[Texture] Extracting Resources")
	if _, err := os.Stat(TextureWorkdir + "/resourcepack"); os.IsNotExist(err) {
		os.Mkdir(filepath.Join(TextureWorkdir, "resourcepack"), 0777)
		if _, err := os.Stat(TextureWorkdir + "/resourcepack.zip"); os.IsNotExist(err) {
			log.Println("[Texture] No resourcepack.zip found")
		} else {
			err := unzip(TextureWorkdir+"/resourcepack.zip", TextureWorkdir+"/resourcepack")
			if err != nil {
				log.Println("[Texture] Error extracting resourcepack.zip")
				return
			}
		}
	}
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()
	os.MkdirAll(dest, 0755)
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()
		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()
			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}
	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}
	return nil
}
