package userfiles

import (
	"bytes"
	"github.com/bxcodec/faker/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/testing/protocmp"
	"io"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"
)

type HttpClientTestSuite struct {
	suite.Suite
}

func (h *HttpClientTestSuite) TestWriteFile() {
	// -- Given
	//
	f := filepath.Join(os.TempDir(), faker.Username())
	_ = os.MkdirAll(f, os.ModePerm)
	defer os.RemoveAll(f)
	s := httptest.NewServer(CreateFileWriterHandler(f))

	given := NewHttpClient(s.URL)

	// -- When
	//
	w := given.CreateFileWriter("key.txt")
	content := []byte("hi")
	_, err := io.Copy(w, bytes.NewBuffer(content))
	if !h.NoError(err) {
		h.FailNow(err.Error())
	}
	err = w.Close()

	// -- Then
	//
	if h.NoError(err) {
		actual, err := os.ReadFile(filepath.Join(f, "key.txt"))
		if h.NoError(err) {
			h.Equal(content, actual)
		}
	}
}

func (h *HttpClientTestSuite) TestFetchFile() {
	// -- Given
	//
	content := []byte("hi")
	key := path.Join("user", "inst", "title", "key.txt")
	f := fstest.MapFS{
		key: {
			Data: content,
		},
	}
	s := httptest.NewServer(FetchFileHandler(f))
	given := NewHttpClient(s.URL)

	// -- When
	//
	reader, err := given.FetchFileReader(key)

	// -- Then
	//
	if h.NoError(err) {
		buf := bytes.NewBuffer([]byte{})
		_, _ = io.Copy(buf, reader.Reader)
		h.Equal(key, reader.Key)
		h.Equal("", reader.Bucket)
		h.Equal(content, buf.Bytes())
	}
}

func (h *HttpClientTestSuite) TestListFiles() {
	// -- Given
	//
	now := time.Now()

	tFs := fstest.MapFS{
		"user/inst/title/saves/key1.txt": {
			ModTime: now.Add(1 * time.Second),
			Data:    []byte(`key1`),
		},
		"user/inst/title/saves/key2.txt": {
			ModTime: now.Add(2 * time.Second),
			Data:    []byte(`key1`),
		},
		"user/inst/title/saves/something/key1.zip": {
			ModTime: now.Add(3 * time.Second),
			Data:    []byte(`key1`),
		},
		"user/inst/title/mods/key1.txt": {
			ModTime: now.Add(1 * time.Second),
			Data:    []byte(`key1`),
		},
	}

	s := httptest.NewServer(ListFolderHandler(tFs, "."))
	given := NewHttpClient(s.URL)
	expected := []*FileHandle{
		{
			Key:      "user/inst/title/saves/key1.txt",
			Name:     "key1.txt",
			Created:  now.Add(1 * time.Second),
			MIME:     "text/plain; charset=utf-8",
			ByteSize: 4,
		},
		{
			Key:      "user/inst/title/saves/key2.txt",
			Name:     "key2.txt",
			Created:  now.Add(2 * time.Second),
			MIME:     "text/plain; charset=utf-8",
			ByteSize: 4,
		},
		{
			Key:      "user/inst/title/saves/something/key1.zip",
			Name:     "key1.zip",
			Created:  now.Add(3 * time.Second),
			MIME:     "application/zip",
			ByteSize: 4,
		},
	}

	// -- When
	//
	actual, err := given.ListFolder("user/inst/title/saves")

	// -- Then
	//
	if h.NoError(err) {
		h.Empty(cmp.Diff(expected, actual, protocmp.Transform()))
	}
}

func (h *HttpClientTestSuite) TestDeleteFile() {
	// -- Given
	//
	f := filepath.Join(os.TempDir(), faker.Username())
	_ = os.MkdirAll(f, os.ModePerm)
	defer os.RemoveAll(f)
	keyPath := filepath.Join(f, "key.txt")
	_ = os.WriteFile(keyPath, []byte{}, os.ModePerm)

	s := httptest.NewServer(DeleteFileHandler(f))
	given := NewHttpClient(s.URL)

	// -- When
	//
	err := given.DeleteFile("key.txt")

	// -- Then
	//
	if h.NoError(err) {
		_, err := os.Stat(keyPath)
		h.False(err == nil)
	}
}

func TestHttpTestSuite(t *testing.T) {
	suite.Run(t, new(HttpClientTestSuite))
}
