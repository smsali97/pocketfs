package main

import (
	"../server/services"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)


func setupRoutes() {
	http.HandleFunc("/directory", services.MainDirectoryService)
	http.HandleFunc("/download", services.DownloadFile)
	http.HandleFunc("/upload", services.UploadFile)
	http.ListenAndServe(":8080", nil)
}

func walkDirectories() {
	// handler function for each file or dir
	var ff = func(pathX string, infoX os.FileInfo, errX error) error {

		// first thing to do, check error. and decide what to do about it
		if errX != nil {
			fmt.Printf("error 「%v」 at a path 「%q」\n", errX, pathX)
			return errX
		}

		fmt.Printf("pathX: %v\n", pathX)

		// find out if it's a dir or file, if file, print info
		if infoX.IsDir() {
			fmt.Printf("is dir.\n")
		} else {
			fmt.Printf("  dir: 「%v」\n", filepath.Dir(pathX))
			fmt.Printf("  file name 「%v」\n", infoX.Name())
			fmt.Printf("  extenion: 「%v」\n", filepath.Ext(pathX))
		}

		return nil
	}
	err := filepath.Walk("file-server", ff)
	if err != nil {
		fmt.Printf("error walking the path: %v\n", err)
		return
	}
}

func main() {
	//walkDirectories()
	setupRoutes()
}
