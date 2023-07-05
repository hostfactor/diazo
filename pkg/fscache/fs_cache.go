package fscache

import (
	"io/fs"
)

type FSCache interface {
	fs.FS
	fs.ReadFileFS
	fs.ReadDirFS
	fs.GlobFS
	fs.StatFS
	fs.SubFS
}

func New(f fs.FS) FSCache {
	return &fsCache{
		FS:    f,
		cache: map[string][]byte{},
	}
}

type fsCache struct {
	fs.FS
	cache map[string][]byte
}

func (f *fsCache) Glob(pattern string) ([]string, error) {
	return fs.Glob(f.FS, pattern)
}

func (f *fsCache) Stat(name string) (fs.FileInfo, error) {
	return fs.Stat(f.FS, name)
}

func (f *fsCache) Sub(dir string) (fs.FS, error) {
	o, err := fs.Sub(f.FS, dir)
	if err != nil {
		return nil, err
	}

	return &fsCache{
		FS:    o,
		cache: f.cache,
	}, nil
}

func (f *fsCache) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(f.FS, name)
}

func (f *fsCache) ReadFile(name string) ([]byte, error) {
	v, ok := f.cache[name]
	if ok {
		return v, nil
	}

	b, err := fs.ReadFile(f.FS, name)
	if err != nil {
		return nil, err
	}

	f.cache[name] = b
	return b, nil
}
