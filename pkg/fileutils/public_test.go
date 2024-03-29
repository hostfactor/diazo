package fileutils

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

type PublicTestSuite struct {
	suite.Suite
}

func (p *PublicTestSuite) TestFsPathMapFS() {
	// -- Given
	//
	given := fstest.MapFS{}

	// -- When
	//
	actual := FsPath(given)

	// -- Then
	//
	p.Equal(".", actual)
}

func (p *PublicTestSuite) TestFsPathDirFS() {
	// -- Given
	//
	given := os.DirFS(os.TempDir())

	// -- When
	//
	actual := FsPath(given)

	// -- Then
	//
	p.Equal(filepath.Clean(os.TempDir()), actual)
}

func (p *PublicTestSuite) TestFsPathSubFS() {
	// -- Given
	//
	dir, name := filepath.Split(filepath.Clean(os.TempDir()))
	given, err := fs.Sub(os.DirFS(filepath.Dir(dir)), name)
	fmt.Println(err)

	// -- When
	//
	actual := FsPath(given)

	// -- Then
	//
	p.Equal(filepath.Clean(os.TempDir()), actual)
}

func (p *PublicTestSuite) TestFsPathSubMapFS() {
	// -- Given
	//
	given := fstest.MapFS{
		"the/path/file.txt": {},
	}

	sub, _ := fs.Sub(given, "the/path")

	// -- When
	//
	actual := FsPath(sub)

	// -- Then
	//
	p.Equal("the/path", actual)
}

func (p *PublicTestSuite) TestFilename() {
	type test struct {
		Given    string
		Expected string
	}

	tests := []test{
		{
			Given:    "I_FONE.wld.zip",
			Expected: "I_FONE",
		},
		{
			Given:    "I_FONE.zip",
			Expected: "I_FONE",
		},
		{},
	}

	for i, v := range tests {
		actual := Filename(v.Given)
		p.Equal(v.Expected, actual, "test %d. given %s", i, v.Given)
	}
}

func TestPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PublicTestSuite))
}
