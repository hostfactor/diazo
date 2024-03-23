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

var IsTest = false

// SplitFile takes a file f and splits it into the name and extensions e.g. save.zip returns (save, .zip).
func SplitFile(f string) (filename, ext string) {
	if f == "" {
		return
	}
	ext = filepath.Ext(f)
	return f[0 : len(f)-len(ext)], ext
}

func Filename(f string) string {
	if f == "" {
		return ""
	}

	ss := strings.Split(f, ".")
	if len(ss) > 0 {
		return ss[0]
	}
	return ""
}

// MoveFile moves the file from src to dst while retaining permissions. src should be the relative path of f.
// dst should be an absolute path.
func MoveFile(f fs.FS, src, dst string) error {
	in, err := f.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %s", err)
	}
	defer in.Close()

	_ = os.MkdirAll(filepath.Dir(dst), os.ModePerm)

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

	si, err := fs.Stat(f, src)
	if err != nil {
		return fmt.Errorf("stat error: %s", err)
	}

	mode := si.Mode()
	if mode == 0 {
		mode = os.ModePerm
	}
	err = os.Chmod(dst, mode)
	if err != nil {
		return fmt.Errorf("chmod error: %s", err)
	}

	err = Remove(f, src)
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

	_, fps := extractFsPaths(reflect.ValueOf(f))

	return filepath.Clean(strings.Join(fps, string(filepath.Separator)))
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
	if IsTest {
		return nil
	}
	switch t := src.(type) {
	case fstest.MapFS:
		delete(t, rel)
	default:
		return os.Remove(filepath.Join(FsPath(src), rel))
	}
	return nil
}

func PersistMapFS(targetDir string, f fstest.MapFS) error {
	for k, v := range f {
		dir, name := filepath.Split(k)
		if dir != "" {
			dir = filepath.Clean(filepath.Join(targetDir, dir))
		} else {
			dir = targetDir
		}
		_ = os.MkdirAll(dir, os.ModePerm)

		if v.Mode == 0 {
			v.Mode = os.ModePerm
		}
		err := os.WriteFile(filepath.Join(dir, name), v.Data, v.Mode)
		if err != nil {
			return err
		}
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

func ChownR(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chown(name, uid, gid)
		}
		return err
	})
}

func extractFsPaths(v reflect.Value) (fs.FS, []string) {
	sl := make([]string, 0)
	var f fs.FS
	if v.CanInterface() {
		i := v.Interface()
		if i != nil {
			f, _ = v.Interface().(fs.FS)
		}
	}

	if f != nil {
		v, ok := f.(fstest.MapFS)
		if ok {
			return v, sl
		}
	}

	for !v.IsZero() && (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		dirField := v.FieldByName("dir")
		if !dirField.IsValid() {
			return f, sl
		}
		if !dirField.IsZero() {
			sl = append(sl, dirField.String())
		}
		fp, s := extractFsPaths(v.FieldByName("fsys"))
		if fp != nil {
			f = fp
		}
		sl = append(sl, s...)
	case reflect.String:
		sl = append(sl, v.String())
	}

	reverseSlice(sl)

	return f, sl
}

func reverseSlice(input []string) {
	for i, j := 0, len(input)-1; i < j; i, j = i+1, j-1 {
		input[i], input[j] = input[j], input[i]
	}
}
