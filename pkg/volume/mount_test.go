package volume

import (
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/collection"
	"github.com/hostfactor/diazo/pkg/filesys"
	"github.com/hostfactor/diazo/pkg/ptr"
	"github.com/stretchr/testify/suite"
	"io/fs"
	"testing"
	"testing/fstest"
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
