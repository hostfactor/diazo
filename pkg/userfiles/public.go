package userfiles

import (
	"github.com/hostfactor/diazo/pkg/fileutils"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Client interface {
	ListFolder(key string) ([]*FileHandle, error)
	DeleteFile(key string) error
	CreateFileWriter(key string) io.WriteCloser
	FetchFileReader(key string) (*FileReader, error)
}

type UrlSigner interface {
	SignedUrl(fileDesc FileDesc, httpMethod string, folder string) (string, error)
}

type Opts struct {
	BucketName          string
	ServiceAccountEmail string
	DisableUrlSigning   bool
}

type FileHandle struct {
	// The filename from the object key.
	Name string

	// The entirety of the object key.
	Key string

	// The time the object was created.
	Created time.Time

	// The name of the bucket that the object resides in.
	Bucket string

	ByteSize uint64

	MIME string
}

type File struct {
	FileHandle
	Content []byte
}

type FileReader struct {
	Key    string
	Bucket string
	Reader io.ReadCloser
}

type FileDesc struct {
	Name     string
	ByteSize uint64
	MIMEType string
}

type DownloadedFile struct {
	// The absolute path to the downloaded file
	Filepath string

	// The size of the file downloaded
	Size int64
}

// DownloadBucketFile downloads a FileReader to the absolute path of a file/folder. If toPath is a directory,
// the file is downloaded into it, if toPath is a file, the FileReader.Name is downloaded
// and renamed to that file. If the file already exists, it is overwritten. All subdirectories are created if they don't
// exist for toPath.
func DownloadBucketFile(reader *FileReader, toPath string) (DownloadedFile, error) {
	ext := filepath.Ext(toPath)
	fileDir := filepath.Dir(toPath)
	_, filename := path.Split(reader.Key)
	if ext == "" {
		fileDir = toPath
		toPath = filepath.Join(toPath, filename)
	}
	_ = os.MkdirAll(fileDir, os.ModePerm)

	df := DownloadedFile{
		Filepath: toPath,
	}

	var err error
	df.Size, err = fileutils.WriteFileFromReader(toPath, reader.Reader)
	if err != nil {
		return df, err
	}

	return df, nil
}
