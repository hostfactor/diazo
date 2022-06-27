package fileloc

import (
	"bytes"
	"github.com/bxcodec/faker/v3"
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/diazo/pkg/testutils"
	"github.com/hostfactor/diazo/pkg/userfiles"
	userfilesmock "github.com/hostfactor/diazo/pkg/userfiles/mocks"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

type PublicTestSuite struct {
	suite.Suite

	Client          *client
	UserfilesClient *userfilesmock.Client
}

func (p *PublicTestSuite) BeforeTest(_, _ string) {
	p.UserfilesClient = new(userfilesmock.Client)

	p.Client = &client{UserfilesClient: p.UserfilesClient}
}

func (p *PublicTestSuite) TestDownloadBucketFile() {
	// -- Given
	//
	tmp := filepath.Join(os.TempDir(), faker.Username())
	defer os.Remove(tmp)

	given := &filesystem.FileLocation{
		BucketFile: &filesystem.BucketFile{
			Name:   faker.Username(),
			Folder: path.Join(faker.Username(), faker.Username()),
		},
	}
	key := path.Join(given.GetBucketFile().GetFolder(), given.GetBucketFile().GetName())
	toPath := filepath.Join(tmp, "save.zip")
	content := &userfiles.FileReader{Reader: io.NopCloser(strings.NewReader("content"))}

	p.UserfilesClient.On("FetchFileReader", key).Return(content, nil)

	// -- When
	//
	w, err := p.Client.Download(given, toPath)

	// -- Then
	//
	if p.NoError(err) {
		p.True(w.Size > 0)
		actualContent, err := os.ReadFile(toPath)
		if p.NoError(err) {
			p.Equal("content", string(actualContent))
		}
		p.UserfilesClient.AssertExpectations(p.T())
	}
}

func (p *PublicTestSuite) TestDownloadBucketFileFolder() {
	// -- Given
	//
	tmp := filepath.Join(os.TempDir(), faker.Username())
	defer os.Remove(tmp)

	given := &filesystem.FileLocation{
		BucketFile: &filesystem.BucketFile{
			Name:   faker.Username() + ".txt",
			Folder: path.Join(faker.Username(), faker.Username()),
		},
	}

	key := path.Join(given.GetBucketFile().GetFolder(), given.GetBucketFile().GetName())
	content := &userfiles.FileReader{Reader: io.NopCloser(strings.NewReader("content")), Key: key}

	p.UserfilesClient.On("FetchFileReader", key).Return(content, nil)

	// -- When
	//
	w, err := p.Client.Download(given, tmp)

	// -- Then
	//
	if p.NoError(err) {
		p.True(w.Size > 0)
		actualContent, err := os.ReadFile(filepath.Join(tmp, given.GetBucketFile().GetName()))
		if p.NoError(err) {
			p.Equal("content", string(actualContent))
		}
		p.UserfilesClient.AssertExpectations(p.T())
	}
}

func (p *PublicTestSuite) TestUploadBucketFile() {
	// -- Given
	//
	givenPath := filepath.Join(faker.Username(), "save.zip")
	givenFs := fstest.MapFS{
		givenPath: {
			Data: []byte("derp"),
		},
	}

	given := &filesystem.BucketFile{
		Name:   faker.Username(),
		Folder: path.Join(faker.Username(), faker.Username()),
	}
	key := path.Join(given.Folder, given.Name)

	writer := &testutils.ByteBuffer{Buffer: bytes.Buffer{}}
	p.UserfilesClient.On("CreateFileWriter", key).Return(writer)

	// -- When
	//
	actual, err := p.Client.UploadBucketFile(givenFs, givenPath, given)

	// -- Then
	//
	if p.NoError(err) {
		p.Equal(int64(4), actual)
		p.Equal("derp", writer.String())
		p.UserfilesClient.AssertExpectations(p.T())
	}
}

func TestPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PublicTestSuite))
}
