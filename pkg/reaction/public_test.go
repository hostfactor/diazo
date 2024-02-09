package reaction

import (
	"context"
	"fmt"
	"github.com/bxcodec/faker/v3"
	"github.com/fsnotify/fsnotify"
	"github.com/hostfactor/api/go/app"
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/api/go/blueprint/reaction"
	"github.com/hostfactor/api/go/mocks"
	actions2 "github.com/hostfactor/diazo/pkg/actions"
	"github.com/hostfactor/diazo/pkg/mocks/actionsmocks"
	"github.com/hostfactor/diazo/pkg/variable"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

type PublicTestSuite struct {
	suite.Suite

	FileActions *actionsmocks.Client
}

func (p *PublicTestSuite) BeforeTest(_, _ string) {
	p.FileActions = new(actionsmocks.Client)
	actions2.Default = p.FileActions
}

func (p *PublicTestSuite) TestWatch() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", "echo 123 && sleep 1 && echo 456")
	stdout, _ := cmd.StdoutPipe()
	str := stdout
	actual := make([]LogLine, 0)
	go func() {
		c := Watch(ctx, str, func(ll LogLine) {
			actual = append(actual, ll)
		})
		<-c.Done()
	}()

	err := cmd.Run()
	if err != nil {
		p.NoError(err)
		return
	}

	cancel()
	expected := []LogLine{
		{Text: "123\n", Num: 1},
		{Text: "456\n", Num: 2},
	}
	p.Equal(expected, actual)
}

func (p *PublicTestSuite) TestExecuteFileTriggerAction() {
	// -- Given
	//
	root := faker.Username()

	type test struct {
		GivenFp       string
		Given         *reaction.FileReactionAction
		ExpectedError error
		Before        func(fp string)
	}

	tests := []test{
		{
			GivenFp: "/opt/file/save.zip",
			Given: &reaction.FileReactionAction{
				Rename: &actions.RenameFiles{To: "${dir}/${filename}", From: &filesystem.DirectoryFileMatcher{Directory: "${abs}"}},
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
			Given: &reaction.FileReactionAction{
				Extract: &actions.ExtractFiles{To: "${dir}/${filename}", From: &filesystem.DirectoryFileMatcher{Directory: "${abs}"}},
			},
			Before: func(fp string) {
				p.FileActions.On("Extract", &actions.ExtractFiles{To: fp, From: &filesystem.DirectoryFileMatcher{Directory: fp}}).Return(nil)
			},
		},
		{
			GivenFp: "/opt/file/save.zip",
			Given: &reaction.FileReactionAction{
				Upload: &actions.UploadFile{
					From: &actions.UploadFile_Source{Path: "${dir}/${filename}"},
					To:   &filesystem.FileLocation{BucketFile: &filesystem.BucketFile{Name: "${name}1.${ext}"}},
				},
			},
			Before: func(fp string) {
				p.FileActions.On("Upload", root, &actions.UploadFile{
					From: &actions.UploadFile_Source{Path: fp},
					To:   &filesystem.FileLocation{BucketFile: &filesystem.BucketFile{Name: "save1.zip"}},
				}, actions2.UploadOpts{}).Return(nil)
			},
		},
		{
			GivenFp: "/opt/file/save.zip",
			Given: &reaction.FileReactionAction{
				Download: &actions.DownloadFile{To: "${ext} ${name}"},
			},
			Before: func(fp string) {
				p.FileActions.On("Download", root, &actions.DownloadFile{To: "zip save"}, actions2.DownloadOpts{}).Return(nil)
			},
		},
		{
			GivenFp: "/opt/file/save.zip",
			Given: &reaction.FileReactionAction{
				Zip: &actions.ZipFile{From: &actions.ZipFile_Source{Directory: "${dir}"}},
			},
			Before: func(fp string) {
				p.FileActions.On("Zip", &actions.ZipFile{From: &actions.ZipFile_Source{Directory: "/opt/file"}}).Return(nil)
			},
		},
		{
			GivenFp: "/opt/file/save.zip",
			Given: &reaction.FileReactionAction{
				Unzip: &actions.UnzipFile{From: "${ext} ${dir}", To: "/my/file/${filename}"},
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
		err := ExecuteFileReactionAction(v.GivenFp, root, variable.NewStore(), v.Given, ExecuteFileOpts{})
		p.Equal(v.ExpectedError, err)
		p.FileActions.AssertExpectations(p.T())
		p.FileActions = new(actionsmocks.Client)
		actions2.Default = p.FileActions
	}
}

func (p *PublicTestSuite) TestExecuteLog() {
	// -- Given
	//
	appService := new(mocks.AppServiceClient)
	type test struct {
		ExpectedErr error
		Store       func() variable.Store
		WriteLines  func(f *os.File)
		Reactions   []*reaction.LogReaction
		Opts        ExecuteLogOpts
		Before      func()
		After       func(store variable.Store)
	}
	log := filepath.Join(os.TempDir(), faker.Username(), "log.txt")
	_ = os.MkdirAll(filepath.Dir(log), os.ModePerm)
	f, _ := os.Create(log)
	defer os.RemoveAll(filepath.Dir(log))

	tests := []test{
		{
			Before: func() {
				appService.On("SetVariable", mock.Anything, &app.SetVariable_Request{
					Name:        "code - 12345",
					Value:       "12345 - ",
					DisplayName: "Code hi",
				}).Return(nil, nil)
			},
			After: func(store variable.Store) {
				p.Equal(2, store.Len())
				p.Equal("hi", store.GetStringValue("value"))
				p.Equal("12345 - ", store.GetStringValue("code - 12345"))
			},
			Store: func() variable.Store {
				store := variable.NewStore()
				store.AddEntries(&variable.Entry{
					Key: "value",
					Val: "hi",
				})
				return store
			},
			WriteLines: func(f *os.File) {
				_, _ = f.WriteString("derp is ready: 12345")
			},
			Opts: ExecuteLogOpts{
				OnStatusChange: func(s actions.SetStatus_Status) {
					p.Equal(s, actions.SetStatus_ready)
				},
			},
			Reactions: []*reaction.LogReaction{
				{
					When: []*reaction.LogReactionCondition{
						{
							Matches: &reaction.LogMatcher{
								Regex: "is ready: (\\d+)",
							},
						},
					},
					Then: []*reaction.LogReactionAction{
						{
							SetVariable: &actions.SetVariable{
								Name:        "code - {{matches.0}}",
								Value:       "{{first_match}} - {{matches.1}}",
								Save:        true,
								DisplayName: "Code {{value}}",
							},
							SetStatus: &actions.SetStatus{
								Status: actions.SetStatus_ready,
							},
						},
					},
				},
			},
		},
	}

	for i, v := range tests {
		fmt.Println("test ", i)
		var store variable.Store
		if v.Store != nil {
			store = v.Store()
		}

		if store == nil {
			store = variable.NewStore()
		}

		if v.Before != nil {
			v.Before()
		}
		ctx, cancel := context.WithCancel(context.Background())
		err := ExecuteLog(ctx, log, store, appService, v.Reactions, v.Opts)
		v.WriteLines(f)

		time.Sleep(100 * time.Millisecond)
		cancel()

		p.Equal(v.ExpectedErr, err)
		if v.After != nil {
			v.After(store)
		}
		appService.AssertExpectations(p.T())
		_ = f.Close()
		f, _ = os.Create(log)
		appService = new(mocks.AppServiceClient)
	}
}

func (p *PublicTestSuite) TestDebounce() {
	// -- Given
	//
	ch := make(chan fsnotify.Event, 1)

	// -- When
	//
	out := Debounce(context.Background(), ch, 5*time.Millisecond)
	timer := time.NewTimer(1 * time.Millisecond)
	count := 0

	// -- Then
	//
	go func() {
		for {
			select {
			case <-timer.C:
				ch <- fsnotify.Event{Name: "derp"}
				return
			default:
			}
			count++
			ch <- fsnotify.Event{Name: "dorp"}
		}
	}()

	evs := make([]fsnotify.Event, 0)
	func() {
		timeout := time.NewTimer(7 * time.Millisecond)
		for {
			select {
			case <-timeout.C:
				return
			case e := <-out:
				evs = append(evs, e)
			}
		}
	}()
	p.Equal(fsnotify.Event{Name: "derp"}, evs[0])
	p.True(count > 1)
}

func TestPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PublicTestSuite))
}
