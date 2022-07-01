package userfiles

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func NewHttpClient(addr string) *HttpClient {
	return &HttpClient{
		Client: http.DefaultClient,
		Addr:   addr,
	}
}

type HttpClient struct {
	Client *http.Client
	Addr   string
}

type ListFolderResponse struct {
	Handlers []*FileHandle `json:"handlers"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func (h *HttpClient) ListFolder(key string) ([]*FileHandle, error) {
	resp, err := h.Client.Get(h.Addr + "/folder?fp=" + genFpQuery(key))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	out := new(ListFolderResponse)
	err = json.NewDecoder(resp.Body).Decode(out)
	if err != nil {
		return nil, err
	}

	return out.Handlers, nil
}

func (h *HttpClient) DeleteFile(key string) error {
	req, err := http.NewRequest(http.MethodDelete, h.Addr+"/file?fp="+genFpQuery(key), nil)
	if err != nil {
		return err
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (h *HttpClient) CreateFileWriter(key string) io.WriteCloser {
	return newWriter(h.Addr + "/file?fp=" + genFpQuery(key))
}

func (h *HttpClient) FetchFileReader(key string) (*FileReader, error) {
	resp, err := h.Client.Get(h.Addr + "/file?fp=" + genFpQuery(key))
	if err != nil {
		return nil, err
	}

	out := &FileReader{
		Key:    key,
		Reader: resp.Body,
	}

	return out, nil
}

func genFpQuery(key string) string {
	return url.QueryEscape(key)
}
