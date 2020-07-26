package main

import (
	"bytes"
	"encoding/gob"
	"github.com/jakemakesstuff/structuredhttp"
	"io"
	"os"
	"path/filepath"
)

// Used to gracefully error the application.
func gracefulError(message string) {
	println(message)
	os.Exit(1)
}

// Used to handle the transfer to the host.
func transferToHost(hostRelPath string, r io.Reader, len uint, perm os.FileMode) {
	// Initialise the transfer.
	type transferInit struct {
		LocalPath string
		TotalBytes uint
		Perm os.FileMode
	}
	buf := &bytes.Buffer{}
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(&transferInit{
		LocalPath:  hostRelPath,
		TotalBytes: len,
		Perm:       perm,
	})
	if err != nil {
		panic(err)
	}
	resp, err := structuredhttp.POST("http://127.0.0.1:8190/v1/StartTransferSession").Reader(buf).Run()
	if err != nil {
		panic(err)
	}
	err = resp.RaiseForStatus()
	if err != nil {
		gracefulError(func() (text string) { text, _ = resp.Text(); return }())
	}
	transferId, err := resp.Text()
	if err != nil {
		panic(err)
	}

	// Chunk the transfer into 1MB blocks.
	block := make([]byte, 1000000)
	for {
		// Read 1MB maximum.
		n, err := r.Read(block)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		resizedBlock := block[:n]

		// Upload the chunk.
		resp, err := structuredhttp.POST("http://127.0.0.1:8190/v1/HandleFragment").Bytes(resizedBlock).Header("Transfer-ID", transferId).Run()
		if err != nil {
			panic(err)
		}
		err = resp.RaiseForStatus()
		if err != nil {
			gracefulError(func() (text string) { text, _ = resp.Text(); return }())
		}
	}
}

// Shows the command usage.
func usage()  {
	println("copyback - copy a file/folder back from the droplet")
	println("usage: copyback <droplet file/folder path> [host save location]")
	os.Exit(0)
}

// Used to process a file.
func processFile(s os.FileInfo, dropletAbsPath, hostRelPath string) {
	// Get the size for later.
	size := s.Size()

	// Create a reader for the file.
	r, err := os.Open(dropletAbsPath)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	// Upload the file.
	transferToHost(hostRelPath, r, uint(size), s.Mode())
}

// The main function.
func main() {
	// Check the arg count.
	if len(os.Args) == 1 {
		usage()
	}

	// Get the droplet path.
	dropletPath := os.Args[1]
	if dropletPath == "-h" {
		usage()
	}

	// Stat the droplet file.
	dropletAbsPath, err := filepath.Abs(dropletPath)
	if err != nil {
		gracefulError(err.Error())
	}

	// Check if the file exists.
	s, err := os.Stat(dropletAbsPath)
	if err != nil {
		gracefulError(err.Error())
		return
	}

	// Get the host relative path.
	var hostRelPath string
	if len(os.Args) > 2 {
		hostRelPath = os.Args[2]
	} else {
		if filepath.Base(dropletAbsPath) == "." {
			hostRelPath = "."
		} else {
			hostRelPath = "./"+filepath.Base(dropletAbsPath)
		}
	}

	// Check if this is a folder.
	if s.IsDir() {
		err := filepath.Walk(dropletAbsPath, func(path string, s os.FileInfo, err error) error {
			if s.IsDir() {
				return nil
			}
			diff, err := filepath.Rel(dropletAbsPath, path)
			if err != nil {
				return err
			}
			processFile(s, path, filepath.Join(hostRelPath, diff))
			return nil
		})
		if err != nil {
			panic(err)
		}
	} else {
		processFile(s, dropletAbsPath, hostRelPath)
	}
}
