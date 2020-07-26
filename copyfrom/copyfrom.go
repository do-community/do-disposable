package main

import (
	"bytes"
	"encoding/gob"
	"github.com/jakemakesstuff/structuredhttp"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

// Shows the command usage.
func usage()  {
	println("copyfrom - copy a file/folder from the host")
	println("usage: copyfrom <host file/folder path> [droplet save location]")
	os.Exit(0)
}

// The main function.
func main() {
	// Check the arg count.
	if len(os.Args) == 1 {
		usage()
	}

	// Get the host path.
	hostRelPath := os.Args[1]
	if hostRelPath == "-h" {
		usage()
	}

	// Get the droplet path.
	var dropletPath string
	var err error
	if len(os.Args) > 2 {
		dropletPath, err = filepath.Abs(os.Args[2])
		if err != nil {
			panic(err)
		}
	} else {
		if filepath.Base(hostRelPath) == "." {
			dropletPath, _ = syscall.FullPath(".")
		} else {
			dropletPath, _ = syscall.FullPath(filepath.Base(hostRelPath))
		}
	}

	// Handle files or folders.
	var handleFileFolder func(hostRelPath, dropletPath string)
	handleFileFolder = func(hostRelPath, dropletPath string) {
		resp, err := structuredhttp.GET("http://127.0.0.1:8190/v1/GetHost").Query("path", hostRelPath).Run()
		if err != nil {
			panic(err)
		}
		if resp.RaiseForStatus() != nil {
			t, _ := resp.Text()
			println(t)
			os.Exit(1)
		}
		if resp.RawResponse.Header.Get("Is-Folder") == "true" {
			// Get the folder information.
			type folderInfo struct {
				Perm os.FileMode
				Contents []string
			}
			b, err := resp.Bytes()
			if err != nil {
				panic(err)
			}
			var info folderInfo
			err = gob.NewDecoder(bytes.NewReader(b)).Decode(&info)
			if err != nil {
				panic(err)
			}

			// Ensure the folder doesn't exist.
			if _, err := os.Stat(dropletPath); !os.IsNotExist(err) {
				println("folder already exists")
				os.Exit(1)
			}

			// Handle making the directory if it doesn't exist.
			_ = os.MkdirAll(dropletPath, info.Perm)

			// Handle each file in the folder.
			for _, c := range info.Contents {
				handleFileFolder(filepath.Join(hostRelPath, c), filepath.Join(dropletPath, c))
			}
		} else {
			// Get the file perms.
			x, err := strconv.ParseUint(resp.RawResponse.Header.Get("Perm"), 10, 64)
			if err != nil {
				panic(err)
			}
			perms := os.FileMode(x)

			// Ensure the file doesn't exist.
			if s, err := os.Stat(dropletPath); !os.IsNotExist(err) {
				if s.IsDir() {
					dropletPath = filepath.Join(dropletPath, filepath.Base(hostRelPath))
				} else {
					println("file already exists")
					os.Exit(1)
				}
			}

			// Write the file.
			w, err := os.Create(dropletPath)
			if err != nil {
				panic(err)
			}
			defer w.Close()
			defer resp.RawResponse.Body.Close()
			_, err = io.Copy(w, resp.RawResponse.Body)
			if err != nil {
				panic(err)
			}
			_ = os.Chmod(dropletPath, perms)
		}
	}
	handleFileFolder(hostRelPath, dropletPath)
}
