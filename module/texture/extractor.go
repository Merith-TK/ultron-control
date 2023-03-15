package main

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
)

func extractResources() {
	log.Println("[Texture] Extracting Resources")
	if _, err := os.Stat(workdir + "/resourcepack"); os.IsNotExist(err) {
		os.Mkdir(filepath.Join(workdir, "resourcepack"), 0777)
		if _, err := os.Stat(workdir + "/resourcepack.zip"); os.IsNotExist(err) {
			log.Println("[Texture] No resourcepack.zip found")
		} else {
			err := unzip(workdir+"/resourcepack.zip", workdir+"/resourcepack")
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
