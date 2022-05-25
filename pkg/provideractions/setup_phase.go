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
	ExtractFiles(fromDirectory string, matches *filesystem.FileMatcher, to string) SetupPhaseBuilder

	// Rename renames all matched file(s) in-place.
	//
	// srcDir is the absolute path to a directory.
	//
	// matches will source all files within srcDir to be renamed.
	//
	// dstFilename is the name of the file that all matches are renamed to. This should be the filename minus the
	// extension e.g. to rename a matching "save.zip" to "autosave.zip" this field would be "autosave"
	Rename(srcDir string, matches *filesystem.FileMatcher, dstFilename string) SetupPhaseBuilder

	// MoveFile moves a singular file. If the matcher matches multiple files, the last will be selected. dst must be absolute paths to files.
	MoveFile(fromDirectory string, matches *filesystem.FileMatcher, dst string) SetupPhaseBuilder

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

func (s *setupPhaseBuilder) MoveFile(fromDirectory string, matches *filesystem.FileMatcher, dst string) SetupPhaseBuilder {
	s.SetupPhase.Actions = append(s.SetupPhase.Actions, &blueprint.SetupAction{
		Move: &actions.MoveFile{
			From: &filesystem.DirectoryFileMatcher{
				Directory: fromDirectory,
				Matches:   matches,
			},
			To: dst,
		},
	})

	return s
}

func (s *setupPhaseBuilder) Rename(srcDir string, matches *filesystem.FileMatcher, dstFilename string) SetupPhaseBuilder {
	s.SetupPhase.Actions = append(s.SetupPhase.Actions, &blueprint.SetupAction{
		Rename: &actions.RenameFiles{
			From: &filesystem.DirectoryFileMatcher{
				Directory: srcDir,
				Matches:   matches,
			},
			To: dstFilename,
		},
	})

	return s
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
		Download: &actions.DownloadFile{
			Source: &actions.DownloadFile_Source{Storage: &filesystem.BucketFileMatcher{
				Matches: &filesystem.FileMatcher{Name: filename},
				Folder:  folder,
			}},
			To: to,
		},
	})
	return s
}

func (s *setupPhaseBuilder) UnzipFile(from, to string) SetupPhaseBuilder {
	s.SetupPhase.Actions = append(s.SetupPhase.Actions, &blueprint.SetupAction{
		Unzip: &actions.UnzipFile{
			From: from,
			To:   to,
		},
	})

	return s
}

func (s *setupPhaseBuilder) ExtractFiles(fromDirectory string, matches *filesystem.FileMatcher, to string) SetupPhaseBuilder {
	s.SetupPhase.Actions = append(s.SetupPhase.Actions, &blueprint.SetupAction{
		Extract: &actions.ExtractFiles{
			From: &filesystem.DirectoryFileMatcher{
				Directory: fromDirectory,
				Matches:   matches,
			},
			To: to,
		},
	})

	return s
}

func (s *setupPhaseBuilder) Build() *blueprint.SetupPhase {
	return s.SetupPhase
}
