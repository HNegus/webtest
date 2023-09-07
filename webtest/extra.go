package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var epoch = time.Unix(0, 0).Format(time.RFC1123)

var noCacheHeaders = map[string]string{
    "Expires":         epoch,
    "Cache-Control":   "no-cache, private, max-age=0",
    "Pragma":          "no-cache",
    "X-Accel-Expires": "0",
}

var etagHeaders = []string{
    "ETag",
    "If-Modified-Since",
    "If-Match",
    "If-None-Match",
    "If-Range",
    "If-Unmodified-Since",
}

func NoCache(h http.Handler) http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
        // Delete any ETag headers that may have been set
        for _, v := range etagHeaders {
            if r.Header.Get(v) != "" {
                r.Header.Del(v)
            }
        }

        // Set our NoCache headers
        for k, v := range noCacheHeaders {
            w.Header().Set(k, v)
        }

        h.ServeHTTP(w, r)
    }

    return http.HandlerFunc(fn)
}

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

	// TODO: simplify
	// TODO: add support for killing server
	printInfo("Server root directory: " + root_folder)

	url := fmt.Sprint("localhost:", port)

	if len(filepaths.php) > 0 {
		if !checkCommandAvailable("php") {
			printFailure("php not available, starting HTTP server")
		} else {
			cmd := exec.Command("php", "-S", url)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Dir = root_folder
			if err := cmd.Start(); err != nil {
				log.Println("Error running php ", err)
			}

			url = fmt.Sprint("http://", url)

			switch runtime.GOOS {
			case "linux":
				exec.Command("xdg-open", url).Start()
			case "windows":
				exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
			case "darwin":
				exec.Command("open", url).Start()
			default:
			}
			cmd.Wait()

			return
		}

	}
	http.Handle("/", NoCache(http.FileServer(http.Dir(root_folder))))
	channel := make(chan int)

	go func(done chan int) {
		err := http.ListenAndServe(fmt.Sprint(":", port), nil)
		if err != nil {
			printFailure(err.Error())
			done <- 0
		}
	}(channel)

	url = fmt.Sprint("http://", url)

	printSuccess("Starting HTTP server on " + url)

	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
	}
	<-channel

}
