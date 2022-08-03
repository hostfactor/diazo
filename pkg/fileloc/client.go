package fileloc

import (
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/diazo/pkg/userfiles"
	"io"
	"io/fs"
	"path"
)

type Client interface {
	// UploadBucketFile uploads a file with the relative path of the fs.FS to the filesystem.BucketFile. The folder of the
	// filesystem.BucketFile must be the full GCS folder path and the file must be the filename with the ext e.g. "save.zip"
	UploadBucketFile(f fs.FS, fromPath string, b *filesystem.BucketFile) (int64, error)

	// Download downloads a filesystem.FileLocation to the absolute path of a file/folder. The folder of the
	// filesystem.FileLocation must be the full GCS folder path and the file must be the filename with the ext e.g. "save.zip".
	// If the toPath is a directory, the file is downloaded into it, if toPath is a file, the filesystem.FileLocation is downloaded
	// and renamed to that file. If the file already exists, it is overwritten. All subdirectories are created if they don't
	// exist for toPath.
	Download(b *filesystem.FileLocation, toPath string) (userfiles.DownloadedFile, error)
}

// New creates a new Client. Ultimately, it acts as a wrapper to convert filesystem definitions to userfiles.Client calls.
// The basePath is a prefix path used for all keys with the userfiles.Client.
func New(c userfiles.Client, basePath string) Client {
	return &client{
		UserfilesClient: c,
		BasePath:        basePath,
	}
}

type client struct {
	UserfilesClient userfiles.Client

	// This path is prefixed to all keys when using the UserfilesClient.
	BasePath string
}

func (c *client) Download(b *filesystem.FileLocation, toPath string) (userfiles.DownloadedFile, error) {
	if source := b.GetBucketFile(); source != nil {
		reader, err := c.UserfilesClient.FetchFileReader(path.Join(c.BasePath, source.GetFolder(), source.GetName()))
		if err != nil {
			return userfiles.DownloadedFile{}, err
		}

		defer func() {
			_ = reader.Reader.Close()
		}()

		return userfiles.DownloadBucketFile(reader, toPath)
	}

	return userfiles.DownloadedFile{}, nil
}

func (c *client) UploadBucketFile(f fs.FS, fromPath string, b *filesystem.BucketFile) (int64, error) {
	fi, err := f.Open(fromPath)
	if err != nil {
		return 0, err
	}
	defer func(fi fs.File) {
		_ = fi.Close()
	}(fi)

	w := c.UserfilesClient.CreateFileWriter(path.Join(c.BasePath, b.Folder, b.Name))

	written, err := io.Copy(w, fi)
	if err != nil {
		return 0, err
	}

	err = w.Close()
	if err != nil {
		return 0, err
	}

	return written, nil
}
