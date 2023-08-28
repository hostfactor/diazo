package volume

import (
	"bytes"
	"github.com/bxcodec/faker/v3"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/mocks/userfilesmocks"
	"github.com/hostfactor/diazo/pkg/userfiles"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"
)

var _ suite.BeforeTest = &VolumeTestSuite{}
var _ suite.AfterTest = &VolumeTestSuite{}

type VolumeTestSuite struct {
	suite.Suite

	UserFilesClient *userfilesmocks.Client
}

func (v *VolumeTestSuite) AfterTest(_, _ string) {
	v.UserFilesClient.AssertExpectations(v.T())
}

func (v *VolumeTestSuite) BeforeTest(_, _ string) {
	v.UserFilesClient = new(userfilesmocks.Client)
}

func (v *VolumeTestSuite) TestMountFile() {
	// -- Given
	//
	d := filepath.Join(os.TempDir(), faker.Username())
	defer func() {
		_ = os.RemoveAll(d)
	}()
	bp := "derp/123"
	given := "_autosave.zip"
	mounter := NewMounter(v.UserFilesClient, bp)
	vol := &providerconfig.Volume{
		Name: "save",
		Mount: &providerconfig.VolumeMount{
			Path: d,
		},
	}
	key := path.Join(bp, "saves", given)
	fileContent := []byte("my file")

	v.UserFilesClient.EXPECT().FetchFileReader(key).
		Return(&userfiles.FileReader{Key: key, Reader: io.NopCloser(bytes.NewBuffer(fileContent))}, nil)

	// -- When
	//
	expectedLen, err := mounter.MountFile(vol, given)

	// -- Then
	//
	if v.NoError(err) {
		v.Equal(expectedLen, int64(len(fileContent)))
		content, _ := os.ReadFile(filepath.Join(d, given))
		v.Equal(fileContent, content)
	}
}

func TestVolumeTestSuite(t *testing.T) {
	suite.Run(t, new(VolumeTestSuite))
}
