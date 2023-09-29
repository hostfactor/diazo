package volume

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/collection"
	"github.com/hostfactor/diazo/pkg/filesys"
	"github.com/hostfactor/diazo/pkg/ptr"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/testing/protocmp"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"
)

type MountTestSuite struct {
	suite.Suite
}

func (m *MountTestSuite) TestReadFile() {
	type test struct {
		Given       string
		Expected    string
		Access      []*filesystem.FileAccessPolicy
		ExpectedErr error
	}

	f := fstest.MapFS{
		"my/dir/hello.txt":     {Data: []byte("hello")},
		"hello.txt":            {Data: []byte("hello root")},
		"my/not/dir/hello.txt": {Data: []byte("hello nested")},
		"my/dir/again/bye.txt": {Data: []byte("byte")},
		"my/dir/cool.txt":      {Data: []byte("cool")},
		"my/dir/cool.zip":      {Data: []byte("cool zip")},
	}

	tests := []test{
		{
			Given:    "hello.txt",
			Expected: "hello root",
			Access: []*filesystem.FileAccessPolicy{
				{
					Perms: []filesystem.FileAccessPolicy_FilePerm{
						filesystem.FileAccessPolicy_read,
					},
					Matches: &filesystem.FileMatcher{Name: "hello.txt"},
				},
			},
		},
		{
			Given:    "my/dir/hello.txt",
			Expected: "hello",
			Access: []*filesystem.FileAccessPolicy{
				{
					Perms: []filesystem.FileAccessPolicy_FilePerm{
						filesystem.FileAccessPolicy_read,
					},
					Recursive: ptr.Ptr(true),
					Matches:   &filesystem.FileMatcher{Name: "hello.txt"},
				},
			},
		},
		{
			Given:       "hello.txt",
			ExpectedErr: fs.ErrNotExist,
		},
	}

	for i, v := range tests {
		mnt := NewMount(&providerconfig.VolumeMount{
			Access: v.Access,
		}, f)

		actual, err := fs.ReadFile(mnt, v.Given)
		if v.ExpectedErr != nil {
			m.EqualError(v.ExpectedErr, err.Error(), "test %d. Given: %s", i, v.Given)
			continue
		}
		if m.NoError(err, "test %d. Given: %s", i, v.Given) {
			m.Equal(v.Expected, string(actual), "test %d. Given: %s", i, v.Given)
		}
	}
}

func (m *MountTestSuite) TestOpenFile() {
	type test struct {
		Given           string
		ExpectedContent string
		Access          []*filesystem.FileAccessPolicy
		ExpectedErr     error
		ExpectedFile    *filesystem.File
	}

	modTime := time.Now()
	f := fstest.MapFS{
		"my/dir/hello.txt":     {Data: []byte("hello"), ModTime: modTime},
		"hello.txt":            {Data: []byte("hello root"), ModTime: modTime},
		"my/not/dir/hello.txt": {Data: []byte("hello nested"), ModTime: modTime},
		"my/dir/again/bye.txt": {Data: []byte("byte"), ModTime: modTime},
		"my/dir/cool.txt":      {Data: []byte("cool"), ModTime: modTime},
		"my/dir/cool.zip":      {Data: []byte("cool zip"), ModTime: modTime},
	}

	tests := []test{
		{
			Given:           "hello.txt",
			ExpectedContent: "hello root",
			Access: []*filesystem.FileAccessPolicy{
				{
					Perms: []filesystem.FileAccessPolicy_FilePerm{
						filesystem.FileAccessPolicy_read,
					},
					Matches: &filesystem.FileMatcher{Name: "hello.txt"},
				},
			},
			ExpectedFile: &filesystem.File{
				Path:    "hello.txt",
				Size:    10,
				Created: modTime.UTC().Unix(),
			},
		},
		{
			Given:           "my/dir/hello.txt",
			ExpectedContent: "hello",
			Access: []*filesystem.FileAccessPolicy{
				{
					Perms: []filesystem.FileAccessPolicy_FilePerm{
						filesystem.FileAccessPolicy_read,
					},
					Recursive: ptr.Ptr(true),
					Matches:   &filesystem.FileMatcher{Name: "hello.txt"},
				},
			},
			ExpectedFile: &filesystem.File{
				Path:    "my/dir/hello.txt",
				Size:    5,
				Created: modTime.UTC().Unix(),
			},
		},
		{
			Given:       "hello.txt",
			ExpectedErr: fs.ErrNotExist,
		},
	}

	for i, v := range tests {
		mnt := NewMount(&providerconfig.VolumeMount{
			Access: v.Access,
		}, f)

		f, err := mnt.Open(v.Given)
		if v.ExpectedErr != nil {
			m.EqualError(v.ExpectedErr, err.Error(), "test %d. Given: %s", i, v.Given)
			continue
		}

		buf := bytes.NewBuffer([]byte{})
		_, _ = io.Copy(buf, f)
		if m.NoError(err, "test %d. Given: %s", i, v.Given) {
			m.Equal(v.ExpectedContent, string(buf.Bytes()), "test %d. Given: %s", i, v.Given)
		}

		m.Empty(cmp.Diff(v.ExpectedFile, filesys.FileToFile(f), protocmp.Transform()), "test %d. Given: %s", i, v.Given)
	}
}

func (m *MountTestSuite) TestReadDir() {
	type test struct {
		Given         string
		ExpectedPaths []string
		Access        []*filesystem.FileAccessPolicy
		ExpectedErr   error
	}

	f := fstest.MapFS{
		"my/dir/hello.txt":     {},
		"hello.txt":            {},
		"my/not/dir/hello.txt": {},
		"my/dir/again/bye.txt": {},
		"my/dir/cool.txt":      {},
		"my/dir/cool.zip":      {},
	}

	tests := []test{
		{
			Given: ".",
			ExpectedPaths: []string{
				"hello.txt",
			},
			Access: []*filesystem.FileAccessPolicy{
				{
					Perms: []filesystem.FileAccessPolicy_FilePerm{
						filesystem.FileAccessPolicy_read,
					},
					Matches: &filesystem.FileMatcher{Name: "hello.txt"},
				},
			},
		},
		{
			Given: "my/dir",
			Access: []*filesystem.FileAccessPolicy{
				{
					Perms: []filesystem.FileAccessPolicy_FilePerm{
						filesystem.FileAccessPolicy_read,
					},
					Matches: &filesystem.FileMatcher{Name: "hello.txt"},
				},
			},
		},
		{
			Given: "my/dir",
			ExpectedPaths: []string{
				"my/dir/hello.txt",
			},
			Access: []*filesystem.FileAccessPolicy{
				{
					Perms: []filesystem.FileAccessPolicy_FilePerm{
						filesystem.FileAccessPolicy_read,
					},
					Recursive: ptr.Ptr(true),
					Matches:   &filesystem.FileMatcher{Name: "hello.txt"},
				},
			},
		},
		{
			Given: ".",
			Access: []*filesystem.FileAccessPolicy{
				{
					Matches: &filesystem.FileMatcher{Name: "hello.zip"},
				},
			},
		},
		{
			Given: ".",
			Access: []*filesystem.FileAccessPolicy{
				{
					Perms: []filesystem.FileAccessPolicy_FilePerm{
						filesystem.FileAccessPolicy_read,
					},
					Matches: &filesystem.FileMatcher{Name: "hello.txt"},
				},
				{
					Perms: []filesystem.FileAccessPolicy_FilePerm{
						filesystem.FileAccessPolicy_none,
					},
					Matches: &filesystem.FileMatcher{Name: "hello.txt"},
				},
			},
		},
		{
			Given: ".",
			ExpectedPaths: []string{
				"my",
				"hello.txt",
			},
			Access: []*filesystem.FileAccessPolicy{
				{
					Perms: []filesystem.FileAccessPolicy_FilePerm{
						filesystem.FileAccessPolicy_read,
					},
					Matches: &filesystem.FileMatcher{Glob: &filesystem.GlobMatcher{Value: []string{"**"}}},
				},
			},
		},
		{
			Given: "my/dir",
			ExpectedPaths: []string{
				"my/dir/cool.txt",
				"my/dir/cool.zip",
				"my/dir/again",
				"my/dir/hello.txt",
			},
			Access: []*filesystem.FileAccessPolicy{
				{
					Recursive: ptr.Ptr(true),
					Perms: []filesystem.FileAccessPolicy_FilePerm{
						filesystem.FileAccessPolicy_read,
					},
				},
			},
		},
		{
			Given: ".",
			ExpectedPaths: []string{
				"hello.txt",
				"my",
			},
			Access: []*filesystem.FileAccessPolicy{
				{
					Recursive: ptr.Ptr(true),
					Perms: []filesystem.FileAccessPolicy_FilePerm{
						filesystem.FileAccessPolicy_read,
					},
				},
			},
		},
	}

	for i, v := range tests {
		mnt := NewMount(&providerconfig.VolumeMount{
			Access: v.Access,
		}, f)

		actual, err := fs.ReadDir(mnt, v.Given)
		if v.ExpectedErr != nil {
			m.EqualError(v.ExpectedErr, err.Error(), "test %d. Given: %s", i, v.Given)
			continue
		}
		if m.NoError(err, "test %d. Given: %s", i, v.Given) {
			m.ElementsMatch(v.ExpectedPaths, collection.Map(actual, filesys.DirEntryAbs), "test %d. Given: %s", i, v.Given)
		}
	}
}

func TestMountTestSuite(t *testing.T) {
	suite.Run(t, new(MountTestSuite))
}
