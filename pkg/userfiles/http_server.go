package userfiles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

type Server interface {
	Shutdown(ctx context.Context) error
	ListenAndServe() error
}

type HttpServer struct {
	server  *http.Server
	BaseDir string

	CreateFileWriter http.HandlerFunc
	ListFiles        http.HandlerFunc
	DeleteFile       http.HandlerFunc
	FetchFile        http.HandlerFunc
}

func (h *HttpServer) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func (h *HttpServer) ListenAndServe() error {
	return h.server.ListenAndServe()
}

func NewServer(addr, baseDir string) *HttpServer {
	d := os.DirFS(baseDir)
	out := &HttpServer{
		server:           &http.Server{Addr: addr},
		BaseDir:          baseDir,
		CreateFileWriter: CreateFileWriterHandler(baseDir),
		ListFiles:        ListFolderHandler(d, baseDir),
		DeleteFile:       DeleteFileHandler(baseDir),
		FetchFile:        FetchFileHandler(d),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/folder", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			out.ListFiles(w, req)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/file", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodDelete:
			out.DeleteFile(w, req)
		case http.MethodGet:
			out.FetchFile(w, req)
		case http.MethodPost:
			out.CreateFileWriter(w, req)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	out.server.Handler = mux

	return out
}

func CreateFileWriterHandler(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fp := req.URL.Query().Get("fp")
		if fp == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = os.MkdirAll(filepath.Join(baseDir, filepath.Dir(fp)), os.ModePerm)
		keyPath := filepath.Join(baseDir, fp)
		f, err := os.Create(keyPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			resp := &ErrorResponse{Message: fmt.Sprintf("Failed to create file %s: %s", keyPath, err.Error())}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		_, err = io.Copy(f, req.Body)
		if err != nil {
			resp := &ErrorResponse{Message: fmt.Sprintf("Failed to upload file %s: %s", keyPath, err.Error())}
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	}
}

func FetchFileHandler(f fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fp := req.URL.Query().Get("fp")
		if fp == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		f, err := f.Open(fp)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			resp := &ErrorResponse{Message: fmt.Sprintf("Failed to fetch file %s: %s", fp, err.Error())}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		_, err = io.Copy(w, f)
		if err != nil {
			resp := &ErrorResponse{Message: fmt.Sprintf("Failed to download file %s: %s", fp, err.Error())}
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	}
}

func DeleteFileHandler(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fp := req.URL.Query().Get("fp")
		if fp == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		keyPath := filepath.Join(baseDir, fp)
		err := os.Remove(keyPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			resp := &ErrorResponse{Message: fmt.Sprintf("Failed to remove file %s: %s", keyPath, err.Error())}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	}
}

func ListFolderHandler(f fs.FS, baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		out := &ListFolderResponse{Handlers: make([]*FileHandle, 0)}
		fp := req.URL.Query().Get("fp")

		keyPath := filepath.Join(baseDir, fp)
		err := fs.WalkDir(f, fp, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			i, err := d.Info()
			if err != nil {
				return err
			}

			out.Handlers = append(out.Handlers, &FileHandle{
				Name:     filepath.Base(path),
				Key:      path,
				Created:  i.ModTime(),
				ByteSize: uint64(i.Size()),
				MIME:     mime.TypeByExtension(filepath.Ext(path)),
			})

			return nil
		})
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			resp := &ErrorResponse{Message: fmt.Sprintf("Failed to list files %s: %s", keyPath, err.Error())}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		_ = json.NewEncoder(w).Encode(out)
	}
}
