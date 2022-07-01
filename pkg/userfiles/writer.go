package userfiles

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func newWriter(addr string) *writer {
	r, w := io.Pipe()

	fs := &writer{
		Writer: w,
		Reader: r,
		Addr:   addr,
		done:   make(chan error, 1),
	}

	return fs
}

type writer struct {
	Writer *io.PipeWriter
	Reader *io.PipeReader
	Addr   string
	done   chan error
	opened bool
}

func (f *writer) Write(p []byte) (n int, err error) {
	if !f.opened {
		f.open()
	}
	return f.Writer.Write(p)
}

func (f *writer) open() {
	go func() {
		f.opened = true
		defer f.Reader.Close()
		resp, err := http.Post(f.Addr, "application/octet-stream", f.Reader)
		if err != nil {
			f.done <- err
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			defer resp.Body.Close()
			r := new(ErrorResponse)
			err := json.NewDecoder(resp.Body).Decode(r)
			if err != nil {
				f.done <- err
				return
			}

			f.done <- fmt.Errorf(r.Message)
		}
		f.done <- nil
	}()
}

func (f *writer) Close() error {
	if !f.opened {
		f.open()
	}
	err := f.Writer.Close()
	if err != nil {
		return err
	}

	err = <-f.done
	return err
}
