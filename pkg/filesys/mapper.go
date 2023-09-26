package filesys

import (
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"io/fs"
	"path/filepath"
)

func FileInfoToFile(basePath string, info fs.FileInfo) *filesystem.File {
	return &filesystem.File{
		Path:    filepath.Join(basePath, info.Name()),
		Size:    info.Size(),
		Created: info.ModTime().UTC().Unix(),
		IsDir:   info.IsDir(),
	}
}

func DirEntryToFile(de fs.DirEntry) *filesystem.File {
	fi, err := de.Info()
	if err != nil {
		return nil
	}
	return FileInfoToFile(DirEntryDir(de), fi)
}

func FileToFile(f fs.File) *filesystem.File {
	fi, err := f.Stat()
	if err != nil {
		return nil
	}

	return FileInfoToFile(FileDir(f), fi)
}
