package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func getServerRoot(filepaths filePaths) (bool, string) {
	/* Try to find index.(html|php). Select the shorter path. */
	var search_paths *[]string
	if len(filepaths.php) > 0 {
		search_paths = &filepaths.php
	} else {
		search_paths = &filepaths.html
	}

	root_file := ""
	found_index := false

	for _, path := range *search_paths {
		base := filepath.Base(path)
		match, _ := filepath.Match("*index*", base)
		if match {
			found_index = true
			if root_file == "" || len(path) < len(root_file) {
				root_file = path
			}
		}
	}
	return found_index, filepath.Dir(root_file)
}

func runDevSever(filepaths filePaths, port uint) {
	printHeading("Running development server")
	found_index, root_folder := getServerRoot(filepaths)

	if found_index {
		printSuccess("Homepage found")
	} else {
		printWarning("No homepage found")
	}

	printInfo("Server root directory: " + root_folder)

	if len(filepaths.php) > 0 {
		if !checkCommandAvailable("php") {
			printFailure("php not available, starting HTTP server")
		} else {
			os.Chdir(root_folder)
			cmd := exec.Command("php", "-S", fmt.Sprint("localhost:", port))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Println("Error running php ", err)
			}
			return
		}

	}

	printSuccess("Starting HTTP server on http://localhost:" + fmt.Sprint(port))
	http.Handle("/", http.FileServer(http.Dir(root_folder)))
	err := http.ListenAndServe(fmt.Sprint(":", port), nil)
	if err != nil {
		printFailure(err.Error())
	}
}
