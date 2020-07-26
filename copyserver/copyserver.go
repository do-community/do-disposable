package copyserver

import (
	"bytes"
	"encoding/gob"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// Defines information about the transfer.
type transferInformation struct {
	totalBytes uint
	writtenBytes uint
	writer io.WriteCloser
}

// The data used to initialise a transfer.
type transferInit struct {
	LocalPath string
	TotalBytes uint
	Perm os.FileMode
}

// Copyserver is used to handle copying between the droplet and host.
func Copyserver(ln net.Listener) error {
	// Defines the router.
	router := httprouter.New()

	// Defines a map of transfer ID > transfer information.
	transferMap := map[string]*transferInformation{}

	// Create the transfer session.
	router.POST("/v1/StartTransferSession", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// Get the initialisation data.
		var data transferInit
		defer r.Body.Close()
		err := gob.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// Get the full path to the item.
		fullPath, err := filepath.Abs(data.LocalPath)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// Check if the path exists.
		if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("file already exists"))
			return
		}

		// Make the directory if it doesn't exist.
		dir, _ := filepath.Split(fullPath)
		err = os.MkdirAll(dir, data.Perm)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// Create the file.
		f, err := os.Create(fullPath)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// Get the transfer ID.
		id := uuid.New().String()

		// Create the transfer information.
		if data.TotalBytes == 0 {
			_ = f.Close()
		} else {
			transferMap[id] = &transferInformation{
				totalBytes:   data.TotalBytes,
				writer:       f,
			}
		}

		// Write the transfer ID.
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(id))
	})

	// Get a file/folder from the host.
	router.GET("/v1/GetHost", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// Get the query param for the relative path.
		relPath := r.URL.Query().Get("path")
		if relPath == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("no path given"))
			return
		}

		// Get the full path to the item.
		fullPath, err := filepath.Abs(relPath)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// Stat the file/folder if possible.
		s, err := os.Stat(fullPath)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if err == os.ErrNotExist {
				_, _ = w.Write([]byte("file or folder does not exist"))
			} else {
				_, _ = w.Write([]byte(err.Error()))
			}
			return
		}

		// Check if this is a folder.
		if s.IsDir() {
			type folderInfo struct {
				Perm os.FileMode
				Contents []string
			}
			c, err := ioutil.ReadDir(fullPath)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			a := make([]string, len(c))
			for i, v := range c {
				a[i] = v.Name()
			}
			buf := &bytes.Buffer{}
			info := &folderInfo{
				Perm:     s.Mode().Perm(),
				Contents: a,
			}
			encoder := gob.NewEncoder(buf)
			err = encoder.Encode(info)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			w.Header().Add("Is-Folder", "true")
			w.WriteHeader(http.StatusOK)
			_, _ = io.Copy(w, buf)
		} else {
			reader, err := os.Open(fullPath)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			w.Header().Set("Is-Folder", "false")
			w.Header().Set("Perm", strconv.FormatUint(uint64(s.Mode().Perm()), 10))
			w.WriteHeader(http.StatusOK)
			defer reader.Close()
			_, err = io.Copy(w, reader)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
		}
	})

	// Handle a transfer fragment.
	router.POST("/v1/HandleFragment", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// Get the transfer.
		TransferID := r.Header.Get("Transfer-ID")
		transfer, ok := transferMap[TransferID]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("transfer not active"))
			return
		}

		// Handle adding the bytes.
		length := r.ContentLength
		if length == -1 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("length unknown"))
			return
		}

		// Read the bytes.
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		r.Body.Close()

		// Check if length is greater than remaining.
		if uint(len(b)) > transfer.totalBytes-transfer.writtenBytes {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("length greater than total required"))
			return
		}

		// Add the length to the written bytes.
		transfer.writtenBytes += uint(length)

		// Write the bytes.
		_, err = transfer.writer.Write(b)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// Write 204.
		w.WriteHeader(http.StatusNoContent)
		if transfer.writtenBytes == transfer.totalBytes {
			_ = transfer.writer.Close()
			delete(transferMap, TransferID)
		}
	})

	// Create the server.
	s := http.Server{
		Handler: router,
	}
	return s.Serve(ln)
}
