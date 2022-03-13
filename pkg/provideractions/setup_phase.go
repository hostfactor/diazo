package provideractions

import (
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/diazo/pkg/ptr"
)

type SetupPhaseBuilder interface {
	// DownloadBucketFile downloads the file e.g. save.zip to an absolute path directory on the server.
	DownloadBucketFile(filename, folder, to string) SetupPhaseBuilder

	// UnzipFile unzips the absolute path of a zip file to an absolute path to a directory.
	UnzipFile(from, to string) SetupPhaseBuilder

	// ExtractFiles moves all sibling files of a matched file to an absolute path of a directory.
	ExtractFiles(fromDirectory string, matches FileMatcher, to string) SetupPhaseBuilder

	Gid(i int) SetupPhaseBuilder

	Uid(i int) SetupPhaseBuilder

	Build() *blueprint.SetupPhase
}

func NewSetupPhaseBuilder() SetupPhaseBuilder {
	return &setupPhaseBuilder{
		SetupPhase: &blueprint.SetupPhase{},
	}
}

type setupPhaseBuilder struct {
	SetupPhase *blueprint.SetupPhase
}

func (s *setupPhaseBuilder) Gid(i int) SetupPhaseBuilder {
	s.SetupPhase.Gid = ptr.Int64(int64(i))
	return s
}

func (s *setupPhaseBuilder) Uid(i int) SetupPhaseBuilder {
	s.SetupPhase.Uid = ptr.Int64(int64(i))
	return s
}

func (s *setupPhaseBuilder) DownloadBucketFile(filename, folder, to string) SetupPhaseBuilder {
	s.SetupPhase.Actions = append(s.SetupPhase.Actions, &blueprint.SetupAction{
		File: &blueprint.SetupAction_Download{Download: &actions.DownloadFile{
			From: &actions.DownloadFile_Storage{Storage: &filesystem.BucketFileMatcher{
				Matches: &filesystem.FileMatcher{File: &filesystem.FileMatcher_Name{Name: filename}},
				Folder:  folder,
			}},
			To: to,
		}},
	})
	return s
}

func (s *setupPhaseBuilder) UnzipFile(from, to string) SetupPhaseBuilder {
	s.SetupPhase.Actions = append(s.SetupPhase.Actions, &blueprint.SetupAction{
		File: &blueprint.SetupAction_Unzip{Unzip: &actions.UnzipFile{
			From: from,
			To:   to,
		}},
	})

	return s
}

func (s *setupPhaseBuilder) ExtractFiles(fromDirectory string, matches FileMatcher, to string) SetupPhaseBuilder {
	s.SetupPhase.Actions = append(s.SetupPhase.Actions, &blueprint.SetupAction{
		File: &blueprint.SetupAction_Extract{
			Extract: &actions.ExtractFiles{
				From: &filesystem.DirectoryFileMatcher{
					Directory: fromDirectory,
					Matches:   matches.FileMatcher(),
				},
				To: to,
			},
		},
	})

	return s
}

func (s *setupPhaseBuilder) Build() *blueprint.SetupPhase {
	return s.SetupPhase
}
