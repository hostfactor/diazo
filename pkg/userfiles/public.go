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
	FetchLatestFile(name string, key Key) (*File, error)
	FetchLatestFileReader(name string, key Key) (*FileReader, error)
	FetchFileContent(name string, folder string, k Key) ([]byte, error)
	ListUserFolder(folder string, k Key) ([]*FileHandle, error)
	ListFolder(k Key) ([]*FileHandle, error)
	DeleteFolder(k Key) error
	DeleteFile(k Key, folder string, fileName string) error
	CreateFileWriter(filename string, folder string, key Key) io.WriteCloser
	SignedUrl(fileDesc FileDesc, httpMethod string, folder string, key Key) (string, error)
	FetchFileReader(key string) (*FileReader, error)
	CreateFileWriterRaw(p string) io.WriteCloser
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

type Key struct {
	InstanceId string
	UserId     string
	Title      string
}

type FileDesc struct {
	Name     string
	ByteSize uint64
	MIMEType string
}

func GenBaseKey(key Key) string {
	return path.Clean(path.Join(key.UserId, key.InstanceId, key.Title))
}

func GenFolderKey(name string, key Key) string {
	return path.Join(GenBaseKey(key), name)
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
