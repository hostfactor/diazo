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
	FS      fs.FS

	OnErrorFunc OnErrorFunc
}

type OnErrorFunc func(err error, req *http.Request)

func (h *HttpServer) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func (h *HttpServer) ListenAndServe() error {
	return h.server.ListenAndServe()
}

func NewServer(addr, baseDir string, f fs.FS) *HttpServer {
	out := &HttpServer{
		server:      &http.Server{Addr: addr},
		BaseDir:     baseDir,
		FS:          f,
		OnErrorFunc: func(_ error, _ *http.Request) {},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/folder", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			out.ListFolderHandler(w, req)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/file", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodDelete:
			out.DeleteFileHandler(w, req)
		case http.MethodGet:
			out.FetchFileHandler(w, req)
		case http.MethodPost:
			out.CreateFileWriterHandler(w, req)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	out.server.Handler = mux

	return out
}

func (h *HttpServer) CreateFileWriterHandler(w http.ResponseWriter, req *http.Request) {
	fp := req.URL.Query().Get("fp")
	if fp == "" {
		w.WriteHeader(http.StatusNotFound)
		h.OnErrorFunc(errors.New("query param fp required"), req)
		return
	}
	_ = os.MkdirAll(filepath.Join(h.BaseDir, filepath.Dir(fp)), os.ModePerm)
	keyPath := filepath.Join(h.BaseDir, fp)
	f, err := os.Create(keyPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		resp := &ErrorResponse{Message: fmt.Sprintf("Failed to create file %s: %s", keyPath, err.Error())}
		_ = json.NewEncoder(w).Encode(resp)
		h.OnErrorFunc(err, req)
		return
	}
	_, err = io.Copy(f, req.Body)
	if err != nil {
		resp := &ErrorResponse{Message: fmt.Sprintf("Failed to upload file %s: %s", keyPath, err.Error())}
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(resp)
		h.OnErrorFunc(err, req)
		return
	}
}

func (h *HttpServer) FetchFileHandler(w http.ResponseWriter, req *http.Request) {
	fp := req.URL.Query().Get("fp")
	if fp == "" {
		w.WriteHeader(http.StatusNotFound)
		h.OnErrorFunc(errors.New("query param fp required"), req)
		return
	}

	f, err := h.FS.Open(fp)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		resp := &ErrorResponse{Message: fmt.Sprintf("Failed to fetch file %s: %s", fp, err.Error())}
		_ = json.NewEncoder(w).Encode(resp)
		h.OnErrorFunc(err, req)
		return
	}

	_, err = io.Copy(w, f)
	if err != nil {
		resp := &ErrorResponse{Message: fmt.Sprintf("Failed to download file %s: %s", fp, err.Error())}
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(resp)
		h.OnErrorFunc(err, req)
		return
	}
}

func (h *HttpServer) DeleteFileHandler(w http.ResponseWriter, req *http.Request) {
	fp := req.URL.Query().Get("fp")
	if fp == "" {
		w.WriteHeader(http.StatusNotFound)
		h.OnErrorFunc(errors.New("query param fp required"), req)
		return
	}

	keyPath := filepath.Join(h.BaseDir, fp)
	err := os.Remove(keyPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		resp := &ErrorResponse{Message: fmt.Sprintf("Failed to remove file %s: %s", keyPath, err.Error())}
		_ = json.NewEncoder(w).Encode(resp)
		h.OnErrorFunc(err, req)
		return
	}
}

func (h *HttpServer) ListFolderHandler(w http.ResponseWriter, req *http.Request) {
	out := &ListFolderResponse{Handlers: make([]*FileHandle, 0)}
	fp := req.URL.Query().Get("fp")

	keyPath := filepath.Join(h.BaseDir, fp)
	err := fs.WalkDir(h.FS, fp, func(path string, d fs.DirEntry, err error) error {
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
		h.OnErrorFunc(err, req)
		return
	}

	_ = json.NewEncoder(w).Encode(out)
}
