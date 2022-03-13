package trigger

import (
	"fmt"
	"github.com/bxcodec/faker/v3"
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/diazo/pkg/fileactions"
	fileactionsmocks "github.com/hostfactor/diazo/pkg/fileactions/mocks"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PublicTestSuite struct {
	suite.Suite

	FileActions *fileactionsmocks.Client
}

func (p *PublicTestSuite) BeforeTest(_, _ string) {
	p.FileActions = new(fileactionsmocks.Client)
	fileactions.Default = p.FileActions
}

func (p *PublicTestSuite) TestExecuteFileTriggerAction() {
	// -- Given
	//
	instanceId := faker.Username()
	userId := faker.Username()
	title := faker.Username()

	type test struct {
		GivenFp       string
		Given         *blueprint.FileTriggerAction
		ExpectedError error
		Before        func(fp string)
	}

	tests := []test{
		{
			GivenFp: "/opt/file/save.zip",
			Given: &blueprint.FileTriggerAction{
				Action: &blueprint.FileTriggerAction_Rename{Rename: &actions.RenameFiles{To: "${dir}/${filename}", From: &filesystem.DirectoryFileMatcher{Directory: "${abs}"}}},
			},
			Before: func(fp string) {
				p.FileActions.On("Rename", &actions.RenameFiles{
					From: &filesystem.DirectoryFileMatcher{
						Directory: fp,
					},
					To: fp,
				}).Return(nil)
			},
		},
		{
			GivenFp: "/opt/file/save.zip",
			Given: &blueprint.FileTriggerAction{
				Action: &blueprint.FileTriggerAction_Extract{Extract: &actions.ExtractFiles{To: "${dir}/${filename}", From: &filesystem.DirectoryFileMatcher{Directory: "${abs}"}}},
			},
			Before: func(fp string) {
				p.FileActions.On("Extract", &actions.ExtractFiles{To: fp, From: &filesystem.DirectoryFileMatcher{Directory: fp}}).Return(nil)
			},
		},
		{
			GivenFp: "/opt/file/save.zip",
			Given: &blueprint.FileTriggerAction{
				Action: &blueprint.FileTriggerAction_Upload{
					Upload: &actions.UploadFile{
						From: &actions.UploadFile_Path{Path: "${dir}/${filename}"},
						To:   &filesystem.FileLocation{Loc: &filesystem.FileLocation_BucketFile{BucketFile: &filesystem.BucketFile{Name: "${name}1.${ext}"}}},
					},
				},
			},
			Before: func(fp string) {
				p.FileActions.On("Upload", instanceId, userId, title, &actions.UploadFile{
					From: &actions.UploadFile_Path{Path: fp},
					To:   &filesystem.FileLocation{Loc: &filesystem.FileLocation_BucketFile{BucketFile: &filesystem.BucketFile{Name: "save1.zip"}}},
				}, fileactions.UploadOpts{}).Return(nil)
			},
		},
		{
			GivenFp: "/opt/file/save.zip",
			Given: &blueprint.FileTriggerAction{
				Action: &blueprint.FileTriggerAction_Download{Download: &actions.DownloadFile{To: "${ext} ${name}"}},
			},
			Before: func(fp string) {
				p.FileActions.On("Download", instanceId, userId, title, &actions.DownloadFile{To: "zip save"}, fileactions.DownloadOpts{}).Return(nil)
			},
		},
		{
			GivenFp: "/opt/file/save.zip",
			Given: &blueprint.FileTriggerAction{
				Action: &blueprint.FileTriggerAction_Zip{Zip: &actions.ZipFile{From: &actions.ZipFile_Directory{Directory: "${dir}"}}},
			},
			Before: func(fp string) {
				p.FileActions.On("Zip", &actions.ZipFile{From: &actions.ZipFile_Directory{Directory: "/opt/file"}}).Return(nil)
			},
		},
		{
			GivenFp: "/opt/file/save.zip",
			Given: &blueprint.FileTriggerAction{
				Action: &blueprint.FileTriggerAction_Unzip{Unzip: &actions.UnzipFile{From: "${ext} ${dir}", To: "/my/file/${filename}"}},
			},
			Before: func(fp string) {
				p.FileActions.On("Unzip", &actions.UnzipFile{From: "zip /opt/file", To: "/my/file/save.zip"}).Return(nil)
			},
		},
	}

	// -- When
	//
	for i, v := range tests {
		fmt.Println("test ", i)
		v.Before(v.GivenFp)
		err := ExecuteFileTriggerAction(v.GivenFp, instanceId, userId, title, v.Given, ExecuteOpts{})
		p.Equal(v.ExpectedError, err)
		p.FileActions.AssertExpectations(p.T())
		p.FileActions = new(fileactionsmocks.Client)
		fileactions.Default = p.FileActions
	}
}

func TestPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PublicTestSuite))
}
