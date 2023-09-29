package filesys

import (
	"io/fs"
	"path/filepath"
)

type File struct {
	fs.File

	// The path to the directory housing the File within the fs.FS.
	Dir string

	fileInfo fs.FileInfo
}

func (f *File) Stat() (fs.FileInfo, error) {
	if f.fileInfo != nil {
		return f.fileInfo, nil
	}
	var err error
	f.fileInfo, err = f.File.Stat()
	return f.fileInfo, err
}

func NewFile(dir string, f fs.File) fs.File {
	return &File{
		File:     f,
		Dir:      dir,
		fileInfo: nil,
	}
}

type DirEntry struct {
	fs.DirEntry

	// The path to the directory housing the DirEntry within the fs.FS.
	Dir string
}

func NewDirEntry(fp string, de fs.DirEntry) fs.DirEntry {
	return &DirEntry{
		DirEntry: de,
		Dir:      fp,
	}
}

// DirEntryDir returns the directory that housed this DirEntry in the fs.FS.
func DirEntryDir(de fs.DirEntry) string {
	v, ok := de.(*DirEntry)
	if ok {
		return v.Dir
	}
	return ""
}

// FileDir returns the directory that housed this File in the fs.FS.
func FileDir(f fs.File) string {
	v, ok := f.(*File)
	if ok {
		return v.Dir
	}
	return ""
}

// DirEntryAbs returns the full path needed to retrieve this DirEntry from the same fs.FS that it came from.
func DirEntryAbs(de fs.DirEntry) string {
	return filepath.Join(DirEntryDir(de), de.Name())
}

// FileAbs returns the full path needed to retrieve this file from the same fs.FS that it came from.
func FileAbs(f fs.File) string {
	st, err := f.Stat()
	if err != nil {
		return ""
	}
	return filepath.Join(FileDir(f), st.Name())
}
