package fileactions

import (
	"archive/zip"
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/diazo/pkg/fileutils"
	"github.com/hostfactor/diazo/pkg/userfiles"
	"github.com/mholt/archiver/v3"
	"github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

var Default Client

type Client interface {
	Rename(r *actions.RenameFiles) error
	Unzip(file *actions.UnzipFile) error
	Extract(file *actions.ExtractFiles) error
	Download(instanceId, userId, title string, dl *actions.DownloadFile, opts DownloadOpts) error
	Upload(instanceId, userId, title string, u *actions.UploadFile, opts UploadOpts) error
	Zip(z *actions.ZipFile) error
}

type OnError func(err error)

type UploadOpts struct {
	OnUpload OnUploadFunc
	OnError  OnError
}

type DownloadOpts struct {
	OnDownload OnDownloadFunc
	OnError    OnError
}

type OnUploadFuncParams struct {
	BytesWritten int64
	Filename     string
	Key          userfiles.Key
	Folder       string
	Error        error
}

type OnUploadFunc func(params OnUploadFuncParams)

type OnDownloadFuncParams struct {
	// The absolute filepath to the file that was downloaded
	ToFilepath   string
	BytesWritten int64

	// The full key in the bucket
	Key string
}

type OnDownloadFunc func(params OnDownloadFuncParams)

type client struct {
	UserfilesClient userfiles.Client
}

func New(c userfiles.Client) Client {
	return &client{UserfilesClient: c}
}

func (i *client) Rename(r *actions.RenameFiles) error {
	return Rename(r)
}

func (i *client) Unzip(file *actions.UnzipFile) error {
	archiver.DefaultZip.MkdirAll = true
	archiver.DefaultZip.OverwriteExisting = true
	archiver.DefaultZip.ImplicitTopLevelFolder = false
	if err := archiver.DefaultZip.Unarchive(file.GetFrom(), file.GetTo()); err != nil {
		return err
	}

	_ = os.Remove(file.GetTo())

	return nil
}

func (i *client) Extract(file *actions.ExtractFiles) error {
	return extract(os.DirFS(file.GetFrom().GetDirectory()), file)
}

func (i *client) Download(instanceId, userId, title string, dl *actions.DownloadFile, opts DownloadOpts) error {
	storage := dl.GetStorage()
	if storage == nil {
		return nil
	}

	readers, err := MatchBucketFiles(i.UserfilesClient, userfiles.Key{
		InstanceId: instanceId,
		UserId:     userId,
		Title:      title,
	}, storage.GetFolder(), storage.GetMatches())
	if err != nil {
		if opts.OnError != nil {
			opts.OnError(err)
		}
		return err
	}

	defer func() {
		for _, v := range readers {
			_ = v.Reader.Close()
		}
	}()

	for _, v := range readers {
		df, err := userfiles.DownloadBucketFile(v, dl.GetTo())
		if err != nil {
			logrus.WithError(err).WithField("path", df.Filepath).WithField("key", v.Key).Error("Failed to write key to path")
			if opts.OnError != nil {
				opts.OnError(err)
			}
			return err
		}
		if opts.OnDownload != nil {
			opts.OnDownload(OnDownloadFuncParams{
				ToFilepath:   df.Filepath,
				BytesWritten: df.Size,
				Key:          v.Key,
			})
		}
	}
	return nil
}

func (i *client) Upload(instanceId, userId, title string, u *actions.UploadFile, opts UploadOpts) error {
	dir, fn := path.Split(u.GetPath())
	return i.upload(os.DirFS(dir), fn, userfiles.Key{
		InstanceId: instanceId,
		UserId:     userId,
		Title:      title,
	}, u, opts)
}

func (i *client) Zip(z *actions.ZipFile) error {
	return zipFs(os.DirFS(z.GetDirectory()), ".", z)
}

func Rename(r *actions.RenameFiles) error {
	return Default.Rename(r)
}

func rename(src fs.FS, destDir string, r *actions.RenameFiles) error {
	matches := GetFsMatches(src, r.GetFrom().GetMatches())
	for _, v := range matches {
		ext := filepath.Ext(v)
		d, _ := filepath.Split(v)
		to := filepath.Clean(filepath.Join(destDir, d, r.GetTo()+ext))
		if err := fileutils.Rename(src, v, to); err != nil {
			return err
		}
	}
	return nil
}

func Unzip(file *actions.UnzipFile) error {
	return Default.Unzip(file)
}

func Extract(file *actions.ExtractFiles) error {
	return Default.Extract(file)
}

func extract(fp fs.FS, file *actions.ExtractFiles) error {
	found, err := Find(fp, file.GetFrom().GetMatches())
	if err != nil {
		logrus.WithError(err).Error("Failed to find matching file when unpacking.")
		return err
	}
	if found == "" {
		return nil
	}

	logrus.WithField("found", found).Debug("Found path to unpack.")

	sub, err := fs.Sub(fp, filepath.Dir(found))
	if err != nil {
		return err
	}

	return fileutils.CopyDir(sub, file.GetTo())
}

func Find(directory fs.FS, matcher *filesystem.FileMatcher) (string, error) {
	found := ""
	err := fs.WalkDir(directory, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		logrus.WithField("path", path).Debug("Walking path.")
		if MatchPath(path, matcher) {
			found = path
			logrus.WithField("path", path).Debug("Found match.")
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil && !fileutils.IsWalkExitErr(err) {
		return found, err
	}

	return found, nil
}

func Download(instanceId, userId, title string, dl *actions.DownloadFile, opts DownloadOpts) error {
	return Default.Download(instanceId, userId, title, dl, opts)
}

func Upload(instanceId, userId, title string, u *actions.UploadFile, opts UploadOpts) error {
	return Default.Upload(instanceId, userId, title, u, opts)
}

func (i *client) upload(f fs.FS, fromPath string, key userfiles.Key, upload *actions.UploadFile, opts UploadOpts) error {
	fi, err := f.Open(fromPath)
	if err != nil {
		if opts.OnError != nil {
			opts.OnError(err)
		}
		return err
	}
	defer func(fi fs.File) {
		_ = fi.Close()
	}(fi)

	switch typ := upload.GetTo().Loc.(type) {
	case *filesystem.FileLocation_BucketFile:
		filename, ext := fileutils.SplitFile(typ.BucketFile.GetName())
		if ext == "" {
			ext = filepath.Ext(upload.GetPath())
		}
		fn := filename + ext

		w := i.UserfilesClient.CreateFileWriter(fn, typ.BucketFile.GetFolder(), key)

		written, err := io.Copy(w, fi)
		if err != nil {
			if opts.OnError != nil {
				opts.OnError(err)
			}
			return err
		}

		err = w.Close()
		if err != nil {
			if opts.OnError != nil {
				opts.OnError(err)
			}
			return err
		}

		if opts.OnUpload != nil {
			opts.OnUpload(OnUploadFuncParams{
				BytesWritten: written,
				Filename:     fn,
				Key:          key,
				Folder:       typ.BucketFile.GetFolder(),
			})
		}
	}

	return nil
}

func Zip(z *actions.ZipFile) error {
	return Default.Zip(z)
}

func zipFs(f fs.FS, dir string, z *actions.ZipFile) error {
	_ = os.MkdirAll(filepath.Dir(z.GetPath()), os.ModePerm)
	archive, err := os.Create(z.GetPath())
	if err != nil {
		return nil
	}
	defer func(archive *os.File) {
		_ = archive.Close()
	}(archive)

	writer := zip.NewWriter(archive)
	err = fs.WalkDir(f, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return err
		}

		w, err := writer.Create(path)
		if err != nil {
			return err
		}

		b, err := fs.ReadFile(f, path)
		if err != nil {
			return err
		}

		_, err = w.Write(b)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	return nil
}
