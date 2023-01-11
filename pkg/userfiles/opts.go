package userfiles

import (
	"io"
	"os"
)

type BlobCreator interface {
	CreateBlob(fp string) (io.WriteCloser, error)
}

var _ BlobCreator = &FileBlobCreator{}

type FileBlobCreator struct {
}

func (f *FileBlobCreator) CreateBlob(fp string) (io.WriteCloser, error) {
	return os.Create(fp)
}
