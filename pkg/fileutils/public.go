package fileutils

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing/fstest"
)

var ErrWalkExit = fmt.Errorf("exit walk")

// SplitFile takes a file f and splits it into the name and extensions e.g. save.zip returns (save, .zip).
func SplitFile(f string) (filename, ext string) {
	if f == "" {
		return
	}
	ext = filepath.Ext(f)
	return f[0 : len(f)-len(ext)], ext
}

// MoveFile moves the file from src to dst while retaining permissions. src should be the relative path of f.
// dst should be an absolute path.
func MoveFile(f fs.FS, src, dst string) error {
	in, err := f.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %s", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to open dest file: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}

	err = out.Sync()
	if err != nil {
		return fmt.Errorf("sync error: %s", err)
	}

	si, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat error: %s", err)
	}

	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return fmt.Errorf("chmod error: %s", err)
	}

	err = os.Remove(src)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}

	return nil
}

func WriteFileFromReader(name string, reader io.Reader) (int64, error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return 0, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	return io.Copy(f, reader)
}

func FsPath(f fs.FS) string {
	if f == nil {
		return ""
	}
	return extractFsPaths(reflect.ValueOf(f))
}

func FindFilenames(f fs.FS) ([]string, error) {
	files := make([]string, 0)
	err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		files = append(files, path)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

type tempFile struct {
	Info fs.DirEntry
	Path string
}

func IsWalkExitErr(err error) bool {
	if err == nil {
		return false
	}
	return err == ErrWalkExit
}

func CopyDir(from fs.FS, to string) error {
	renameFiles := map[string]tempFile{}
	err := fs.WalkDir(from, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fullPath := filepath.Join(to, path)
		renameFiles[fullPath] = tempFile{Info: d, Path: path}
		return nil
	})
	if err != nil && !IsWalkExitErr(err) {
		return err
	}

	for k, v := range renameFiles {
		b, err := fs.ReadFile(from, v.Path)
		if err != nil {
			return err
		}
		_ = os.MkdirAll(filepath.Dir(k), os.ModePerm)
		if err := os.WriteFile(k, b, os.ModePerm); err != nil {
			return err
		}
	}

	return err
}

func Rename(src fs.FS, from, to string) error {
	f, err := fs.ReadFile(src, from)
	if err != nil {
		return err
	}

	_ = os.MkdirAll(filepath.Dir(to), os.ModePerm)

	err = os.WriteFile(to, f, os.ModePerm)
	if err != nil {
		return err
	}

	return Remove(src, from)
}

func Remove(src fs.FS, rel string) error {
	switch t := src.(type) {
	case fstest.MapFS:
		delete(t, rel)
	default:
		return os.Remove(filepath.Join(FsPath(src), rel))
	}
	return nil
}

func InspectZipFile(fp string) (*zip.Reader, error) {
	d, f := filepath.Split(fp)
	return InspectZipFileFs(os.DirFS(d), f)
}

func InspectZipFileFs(f fs.FS, fp string) (*zip.Reader, error) {
	fi, err := f.Open(fp)
	if err != nil {
		return nil, err
	}
	defer func(fi fs.File) {
		_ = fi.Close()
	}(fi)

	stat, err := fi.Stat()
	if err != nil {
		return nil, err
	}

	readerAt, ok := fi.(io.ReaderAt)
	if !ok {
		return &zip.Reader{}, nil
	}

	read, err := zip.NewReader(readerAt, stat.Size())
	if err != nil {
		return nil, err
	}

	return read, nil
}

func extractFsPaths(v reflect.Value) string {
	sl := make([]string, 0)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface || v.Kind() == reflect.Struct {
		for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			dirField := v.FieldByName("dir")
			if !dirField.IsZero() {
				sl = append(sl, dirField.String())
			}
			v = v.FieldByName("fsys")
		} else {
			break
		}
	}

	if v.Kind() == reflect.String {
		sl = append(sl, v.String())
	}

	builder := strings.Builder{}
	for j := len(sl) - 1; j >= 0; j-- {
		builder.WriteString(sl[j])
		builder.WriteRune(filepath.Separator)
	}
	return filepath.Clean(builder.String())
}
