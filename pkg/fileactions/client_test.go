package fileactions

import (
	"bytes"
	"errors"
	"github.com/bxcodec/faker/v3"
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/diazo/pkg/fileutils"
	"github.com/hostfactor/diazo/pkg/testutils"
	"github.com/hostfactor/diazo/pkg/userfiles"
	"github.com/hostfactor/diazo/pkg/userfiles/mocks"
	"github.com/stretchr/testify/suite"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

type ClientTestSuite struct {
	suite.Suite

	UserfilesClient *mocks.Client
	Svc             *client
}

func (p *ClientTestSuite) BeforeTest(_, _ string) {
	p.UserfilesClient = new(mocks.Client)
	p.Svc = &client{UserfilesClient: p.UserfilesClient}
	fileutils.IsTest = true
	Default = p.Svc
}

func (p *ClientTestSuite) TestExtractValheim() {
	// -- Given
	//
	to := filepath.Join(os.TempDir(), faker.Username())
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(to)

	given := &actions.ExtractFiles{
		From: &filesystem.DirectoryFileMatcher{
			Matches: &filesystem.FileMatcher{
				Regex: ".+\\.fwl",
			},
		},
		To: to,
	}

	givenFs := fstest.MapFS{
		"valheim_save/nested/world.db":      {Data: []byte("")},
		"valheim_save/nested/world.db.old":  {Data: []byte("")},
		"valheim_save/nested/world.fwl":     {Data: []byte("")},
		"valheim_save/nested/world.fwl.old": {Data: []byte("")},
		"valheim_save/nested/deeper/a.txt":  {Data: []byte("")},
		"valheim_save/blank.txt":            {Data: []byte("")},
	}

	// -- When
	//
	err := extract(givenFs, given)

	// -- Then
	//
	if p.NoError(err) {
		p.EqualFiles([]string{"world.db", "world.db.old", "world.fwl", "world.fwl.old", "deeper/a.txt"}, os.DirFS(to))
	}
}

func (p *ClientTestSuite) TestRename() {
	// -- Given
	//
	to := filepath.Join(os.TempDir(), faker.Username())
	_ = os.MkdirAll(to, os.ModePerm)
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(to)
	givenFs := fstest.MapFS{
		"match.txt":           {Data: []byte("")},
		"match.zip":           {Data: []byte("")},
		"not_match.txt":       {Data: []byte("")},
		"match/not_match.txt": {Data: []byte("")},
		"match/match.jpg":     {Data: []byte("")},
	}
	given := &actions.RenameFiles{
		From: &filesystem.DirectoryFileMatcher{
			Matches: &filesystem.FileMatcher{
				Glob: &filesystem.GlobMatcher{Value: []string{"match.*", "match/match.*"}},
			},
		},
		To: "derp",
	}

	// -- When
	//
	err := rename(givenFs, to, given)

	// -- Then
	//
	if p.NoError(err) {
		p.EqualFiles([]string{
			"derp.txt",
			"derp.zip",
			"match/derp.jpg",
		}, os.DirFS(to))
	}
}

func (p *ClientTestSuite) TestUnzip() {
	// -- Given
	//
	givenPath := filepath.Join(filepath.Dir(testutils.GetCurrentFile()), "testdata", "test_unzip.zip")
	unzipTo := filepath.Join(os.TempDir(), faker.Username(), "test")
	unpackTo := filepath.Join(os.TempDir(), faker.Username(), "unzip")
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(unpackTo)
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(unzipTo)
	given := &actions.UnzipFile{
		From: givenPath,
		To:   unzipTo,
	}

	// -- When
	//
	err := Unzip(given)

	// -- Then
	//
	if p.NoError(err) {
		p.EqualFiles([]string{"test_unzip/nested/a.txt", "test_unzip/world.db", "test_unzip/world.db.old", "test_unzip/world.fwl", "test_unzip/world.fwl.old"}, os.DirFS(unzipTo))
	}
}

func (p *ClientTestSuite) TestMatchFs() {
	// -- Given
	//
	type test struct {
		Fs       fs.FS
		Given    *filesystem.FileMatcher
		Expected []string
	}

	tests := []test{
		{
			Fs:       fstest.MapFS{},
			Given:    &filesystem.FileMatcher{Glob: &filesystem.GlobMatcher{Value: []string{"match.jpg"}}},
			Expected: []string{},
		},
		{
			Fs: fstest.MapFS{
				"match.jpg":  {Data: []byte("")},
				"match1.txt": {Data: []byte("")},
			},
			Given:    &filesystem.FileMatcher{Glob: &filesystem.GlobMatcher{Value: []string{"match.*"}}},
			Expected: []string{"match.jpg"},
		},
		{
			Fs: fstest.MapFS{
				"nested/match.jpg": {Data: []byte("")},
				"match.jpg":        {Data: []byte("")},
				"match1.txt":       {Data: []byte("")},
			},
			Given:    &filesystem.FileMatcher{Name: "match.jpg"},
			Expected: []string{"match.jpg"},
		},
		{
			Fs: fstest.MapFS{
				"nested/match.jpg": {Data: []byte("")},
				"match.jpg":        {Data: []byte("")},
				"match1.txt":       {Data: []byte("")},
			},
			Given:    &filesystem.FileMatcher{Regex: ".+\\.jpg"},
			Expected: []string{"match.jpg"},
		},
	}

	// -- When
	//
	for i, v := range tests {
		actual := GetFsMatches(v.Fs, v.Given)
		p.ElementsMatch(v.Expected, actual, "test %d", i)
	}
}

func (p *ClientTestSuite) TestMatchPath() {
	// -- Given
	//
	type test struct {
		Name     string
		Matcher  *filesystem.FileMatcher
		Expected bool
	}

	tests := []test{
		{
			Name:     "a.txt",
			Matcher:  &filesystem.FileMatcher{Glob: &filesystem.GlobMatcher{Value: []string{"b.txt", "a.*"}}},
			Expected: true,
		},
		{
			Name:    "b.txt",
			Matcher: &filesystem.FileMatcher{Glob: &filesystem.GlobMatcher{Value: []string{"a.*"}}},
		},
		{
			Name:     "a.txt",
			Matcher:  &filesystem.FileMatcher{Regex: "a.+"},
			Expected: true,
		},
		{
			Name:    "b.txt",
			Matcher: &filesystem.FileMatcher{Regex: "a.+"},
		},
		{
			Name:     "a.txt",
			Matcher:  &filesystem.FileMatcher{Name: "a.txt"},
			Expected: true,
		},
		{
			Name:    "b.txt",
			Matcher: &filesystem.FileMatcher{},
		},
	}

	// -- When
	//
	for i, v := range tests {
		actual := MatchPath(v.Name, v.Matcher)
		p.Equal(v.Expected, actual, "test %d", i)
	}
}

func (p *ClientTestSuite) TestZip() {
	// -- Given
	//
	f := fstest.MapFS{
		"file/a.txt": {Data: []byte("")},
		"file/b.txt": {Data: []byte("")},
		"opt/c.txt":  {Data: []byte("")},
	}

	dest := filepath.Join(os.TempDir(), faker.Username(), "test.zip")
	given := &actions.ZipFile{
		From: &actions.ZipFile_Source{Directory: "."},
		To:   &actions.ZipFile_Destination{Path: dest},
	}

	// -- When
	//
	err := zipFs(f, "file", given)
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(dest)

	// -- Then
	//
	if p.NoError(err) {
		r, err := fileutils.InspectZipFile(dest)
		if p.NoError(err) {
			filenames := make([]string, len(r.File))
			for i, v := range r.File {
				filenames[i] = v.Name
			}
			p.ElementsMatch([]string{"file/a.txt", "file/b.txt"}, filenames)
		}
	}
}

func (p *ClientTestSuite) TestMatchDirectoryFile() {
	// -- Given
	//
	type test struct {
		F        fs.FS
		Given    *filesystem.DirectoryFileMatcher
		Expected bool
	}

	matcher := &filesystem.FileMatcher{
		Regex: ".+",
	}

	cwd := filepath.Dir(testutils.GetCurrentFile())
	tests := []test{
		{
			F: fstest.MapFS{
				"a.txt": {},
			},
			Given: &filesystem.DirectoryFileMatcher{
				Directory: ".",
				Matches:   matcher,
			},
			Expected: true,
		},
		{
			F: os.DirFS(cwd),
			Given: &filesystem.DirectoryFileMatcher{
				Directory: filepath.Join(cwd, ".."),
				Matches:   matcher,
			},
		},
	}

	// -- When
	//
	for i, v := range tests {
		actual := MatchDirectoryFile(v.F, v.Given)
		p.Equal(v.Expected, actual, "test %d", i)
	}
}

func (p *ClientTestSuite) TestDownload() {
	// -- Given
	//
	instanceId := faker.Username()
	userId := faker.Username()
	title := faker.Username()

	type test struct {
		Match         *filesystem.FileMatcher
		To            func(d string) string
		Before        func(d string)
		After         func(d string)
		ExpectedError error
	}

	key := userfiles.Key{
		InstanceId: instanceId,
		UserId:     userId,
		Title:      title,
	}
	folderKey := userfiles.GenFolderKey("saves", key)
	tests := []test{
		{
			Match: &filesystem.FileMatcher{
				Name: "save.zip",
			},
			Before: func(d string) {
				p.UserfilesClient.On("FetchFileReader", path.Join(folderKey, "save.zip")).Return(&userfiles.FileReader{
					Key:    path.Join(folderKey, "save.zip"),
					Reader: io.NopCloser(strings.NewReader("save")),
				}, nil)
			},
			After: func(d string) {
				de, err := os.ReadDir(d)
				if p.NoError(err) {
					if p.Len(de, 1) {
						b, err := os.ReadFile(filepath.Join(d, "save.zip"))
						if p.NoError(err) {
							p.Equal("save", string(b))
						}
					}
				}
			},
		},
		{
			Match: &filesystem.FileMatcher{
				Name: "save.zip",
			},
			To: func(d string) string {
				return filepath.Join(d, "save.txt")
			},
			Before: func(d string) {
				p.UserfilesClient.On("FetchFileReader", path.Join(folderKey, "save.zip")).Return(&userfiles.FileReader{
					Key:    path.Join(folderKey, "save.zip"),
					Reader: io.NopCloser(strings.NewReader("save")),
				}, nil)
			},
			After: func(d string) {
				de, err := os.ReadDir(d)
				if p.NoError(err) {
					if p.Len(de, 1) {
						b, err := os.ReadFile(filepath.Join(d, "save.txt"))
						if p.NoError(err) {
							p.Equal("save", string(b))
						}
					}
				}
			},
		},
		{
			Match: &filesystem.FileMatcher{
				Name: "save.zip",
			},
			Before: func(d string) {
				p.UserfilesClient.On("FetchFileReader", path.Join(folderKey, "save.zip")).Return(nil, errors.New("error"))
			},
			ExpectedError: errors.New("error"),
			After: func(d string) {
				de, err := os.ReadDir(d)
				if p.NoError(err) {
					p.Len(de, 0)
				}
			},
		},
		{
			Match: &filesystem.FileMatcher{
				Regex: ".+\\.zip",
			},
			Before: func(d string) {
				handles := []*userfiles.FileHandle{
					{Key: path.Join(folderKey, "derp.zip"), Name: "derp.zip"},
					{Key: path.Join(folderKey, "asd.zip"), Name: "asd.zip"},
					{Key: path.Join(folderKey, "asd.jpg"), Name: "asd.jpg"},
				}
				p.UserfilesClient.On("ListUserFolder", "saves", key).Return(handles, nil)
				for _, v := range handles {
					p.UserfilesClient.On("FetchFileReader", v.Key).Return(&userfiles.FileReader{
						Key:    v.Key,
						Reader: io.NopCloser(strings.NewReader(v.Key)),
					}, nil)
				}
			},
			After: func(d string) {
				de, err := os.ReadDir(d)
				if p.NoError(err) {
					p.Len(de, 2)
					b, err := os.ReadFile(filepath.Join(d, "derp.zip"))
					if p.NoError(err) {
						p.Equal(path.Join(folderKey, "derp.zip"), string(b))
					}
					b, err = os.ReadFile(filepath.Join(d, "asd.zip"))
					if p.NoError(err) {
						p.Equal(path.Join(folderKey, "asd.zip"), string(b))
					}
				}
			},
		},
		{
			Match: &filesystem.FileMatcher{
				Regex: ".+\\.zip",
			},
			Before: func(d string) {
				p.UserfilesClient.On("ListUserFolder", "saves", key).Return(make([]*userfiles.FileHandle, 0), nil)
			},
			After: func(d string) {
				de, err := os.ReadDir(d)
				if p.NoError(err) {
					p.Len(de, 0)
				}
			},
		},
		{
			Match: &filesystem.FileMatcher{
				Glob: &filesystem.GlobMatcher{Value: []string{"*.zip", "*.jpg"}},
			},
			Before: func(d string) {
				handles := []*userfiles.FileHandle{
					{Key: path.Join(folderKey, "derp.zip"), Name: "derp.zip"},
					{Key: path.Join(folderKey, "asd.zip"), Name: "asd.zip"},
					{Key: path.Join(folderKey, "asd.jpg"), Name: "asd.jpg"},
					{Key: path.Join(folderKey, "asd.png"), Name: "asd.png"},
				}
				p.UserfilesClient.On("ListUserFolder", "saves", key).Return(handles, nil)
				for _, v := range handles {
					p.UserfilesClient.On("FetchFileReader", v.Key).Return(&userfiles.FileReader{
						Key:    v.Key,
						Reader: io.NopCloser(strings.NewReader(v.Key)),
					}, nil)
				}
			},
			After: func(d string) {
				de, err := os.ReadDir(d)
				if p.NoError(err) {
					p.Len(de, 3)
					b, err := os.ReadFile(filepath.Join(d, "derp.zip"))
					if p.NoError(err) {
						p.Equal(path.Join(folderKey, "derp.zip"), string(b))
					}
					b, err = os.ReadFile(filepath.Join(d, "asd.zip"))
					if p.NoError(err) {
						p.Equal(path.Join(folderKey, "asd.zip"), string(b))
					}
					b, err = os.ReadFile(filepath.Join(d, "asd.jpg"))
					if p.NoError(err) {
						p.Equal(path.Join(folderKey, "asd.jpg"), string(b))
					}
				}
			},
		},
		{
			Match: &filesystem.FileMatcher{
				Glob: &filesystem.GlobMatcher{Value: []string{"*.zip", "*.jpg"}},
			},
			Before: func(d string) {
				p.UserfilesClient.On("ListUserFolder", "saves", key).Return(make([]*userfiles.FileHandle, 0), nil)
			},
			After: func(d string) {
				de, err := os.ReadDir(d)
				if p.NoError(err) {
					p.Len(de, 0)
				}
			},
		},
	}

	// -- When
	//
	for i, v := range tests {
		dir := filepath.Join(os.TempDir(), faker.Username())
		given := &actions.DownloadFile{
			Source: &actions.DownloadFile_Source{
				Storage: &filesystem.BucketFileMatcher{
					Matches: v.Match,
					Folder:  "saves",
				},
			},
			To: dir,
		}

		_ = os.MkdirAll(given.To, os.ModePerm)
		defer func() {
			_ = os.RemoveAll(given.To)
		}()

		if v.To != nil {
			given.To = v.To(given.To)
		}

		v.Before(dir)
		err := p.Svc.Download(instanceId, userId, title, given, DownloadOpts{})
		v.After(dir)
		p.Equal(v.ExpectedError, err, "test %d", i)
		p.UserfilesClient.AssertExpectations(p.T())
		p.UserfilesClient = new(mocks.Client)
		p.Svc.UserfilesClient = p.UserfilesClient
	}
}

func (p *ClientTestSuite) TestUpload() {
	// -- Given
	//
	type test struct {
		Before        func(closer *testutils.ByteBuffer)
		After         func(closer *testutils.ByteBuffer)
		ExpectedError error
		Fs            fs.FS
		Given         *actions.UploadFile
	}

	key := userfiles.Key{
		InstanceId: faker.Username(),
		UserId:     faker.Username(),
		Title:      faker.Username(),
	}
	tests := []test{
		{
			Fs: fstest.MapFS{
				"opt/a.txt": {Data: []byte("text")},
			},
			Given: &actions.UploadFile{
				From: &actions.UploadFile_Source{Path: "opt/a.txt"},
				To: &filesystem.FileLocation{BucketFile: &filesystem.BucketFile{
					Name:   "save.txt",
					Folder: "saves",
				}},
			},
			Before: func(b *testutils.ByteBuffer) {
				p.UserfilesClient.On("CreateFileWriter", "save.txt", "saves", key).Return(b, nil)
			},
			After: func(b *testutils.ByteBuffer) {
				p.Equal(b.String(), "text")
			},
		},
		{
			Fs: fstest.MapFS{
				"opt/a.txt": {Data: []byte("text")},
			},
			Given: &actions.UploadFile{
				From: &actions.UploadFile_Source{Path: "opt/a.txt"},
				To: &filesystem.FileLocation{BucketFile: &filesystem.BucketFile{
					Name:   "save",
					Folder: "saves",
				}},
			},
			Before: func(b *testutils.ByteBuffer) {
				p.UserfilesClient.On("CreateFileWriter", "save.txt", "saves", key).Return(b)
			},
			After: func(b *testutils.ByteBuffer) {
				p.Equal(b.String(), "text")
			},
		},
	}

	// -- When
	//
	for i, v := range tests {
		b := &testutils.ByteBuffer{
			Buffer: bytes.Buffer{},
		}
		v.Before(b)
		err := p.Svc.upload(v.Fs, v.Given.GetFrom().GetPath(), key, v.Given, UploadOpts{})
		v.After(b)
		p.Equal(err, v.ExpectedError, "test %d", i)
		p.UserfilesClient.AssertExpectations(p.T())
		p.UserfilesClient = new(mocks.Client)
		p.Svc.UserfilesClient = p.UserfilesClient
	}
}

func (p *ClientTestSuite) TestMove() {
	// -- Given
	//
	to := filepath.Join(os.TempDir(), faker.Username())
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(to)

	givenFs := fstest.MapFS{
		"my/dir/file.txt": {
			Data: []byte(`file`),
		},
	}
	given := &actions.MoveFile{
		From: &filesystem.DirectoryFileMatcher{
			Directory: "/my/dir",
			Matches: &filesystem.FileMatcher{
				Name: "file.txt",
			},
		},
		To: filepath.Join(to, "file.txt"),
	}

	// -- When
	//
	err := move(givenFs, given)

	// -- Then
	//
	if p.NoError(err) {
		b, err := os.ReadFile(given.GetTo())
		if p.NoError(err) {
			p.Equal("file", string(b))
		}
	}
}

func (p *ClientTestSuite) EqualFiles(expectedNames []string, f fs.FS) bool {
	files, err := fileutils.FindFilenames(f)
	if p.NoError(err) {
		return p.ElementsMatch(expectedNames, files)
	}
	return false
}

func TestPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
