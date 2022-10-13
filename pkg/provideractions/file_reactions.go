package provideractions

import (
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/api/go/blueprint/reaction"
	"github.com/hostfactor/diazo/pkg/fileutils"
	"strings"
)

func NewFileReaction() FileReactionBuilder {
	return &fileReactionBuilder{}
}

type FileReactionBuilder interface {
	reactionBuilder
	When(builders ...FileReactionConditionBuilder) FileReactionBuilder
	Then(builders ...FileReactionActionBuilder) FileReactionBuilder
}

type FileReactionConditionBuilder interface {
	// Matches resolves to true if the file change matches the specified matcher.
	Matches(fm *filesystem.FileMatcher) FileReactionConditionBuilder

	// DoesntMatch resolves to true if the file change does not match the specified matcher.
	DoesntMatch(fm *filesystem.FileMatcher) FileReactionConditionBuilder

	// Op resolves to true if one of the specified blueprint.FileChangeOp match.
	Op(op ...reaction.FileReactionCondition_FileOp) FileReactionConditionBuilder

	// Directories adds a directory to be watched for all conditions.
	Directories(d ...string) FileReactionConditionBuilder

	Build() *reaction.FileReactionCondition
}

type FileReactionActionBuilder interface {
	Upload(from string, filename string, folder string) FileReactionActionBuilder

	// ZipFolder zips the absolute path f to a zip
	ZipFolder(to string, entries ...*actions.ZipFileEntry) FileReactionActionBuilder

	Build() *reaction.FileReactionAction
}

type fileReactionBuilder struct {
	Whens []FileReactionConditionBuilder
	Thens []FileReactionActionBuilder
}

func (f *fileReactionBuilder) Build() *reaction.Reaction {
	react := &reaction.Reaction{
		FileReaction: &reaction.FileReaction{
			When: make([]*reaction.FileReactionCondition, 0, len(f.Whens)),
			Then: make([]*reaction.FileReactionAction, 0, len(f.Thens)),
		},
	}
	for _, v := range f.Whens {
		react.FileReaction.When = append(react.FileReaction.When, v.Build())
	}
	for _, v := range f.Thens {
		react.FileReaction.Then = append(react.FileReaction.Then, v.Build())
	}

	return react
}

func (f *fileReactionBuilder) When(builders ...FileReactionConditionBuilder) FileReactionBuilder {
	f.Whens = builders
	return f
}

func (f *fileReactionBuilder) Then(builders ...FileReactionActionBuilder) FileReactionBuilder {
	f.Thens = builders
	return f
}

func NewFileReactionCondition() FileReactionConditionBuilder {
	return &fileReactionConditionBuilder{
		FileTriggerCondition: &reaction.FileReactionCondition{},
	}
}

type fileReactionConditionBuilder struct {
	FileTriggerCondition *reaction.FileReactionCondition
}

func (f *fileReactionConditionBuilder) Matches(fm *filesystem.FileMatcher) FileReactionConditionBuilder {
	f.FileTriggerCondition.Matches = fm
	return f
}

func (f *fileReactionConditionBuilder) DoesntMatch(fm *filesystem.FileMatcher) FileReactionConditionBuilder {
	f.FileTriggerCondition.DoesntMatch = fm
	return f
}

func (f *fileReactionConditionBuilder) Op(op ...reaction.FileReactionCondition_FileOp) FileReactionConditionBuilder {
	f.FileTriggerCondition.Op = op
	return f
}

func (f *fileReactionConditionBuilder) Directories(d ...string) FileReactionConditionBuilder {
	f.FileTriggerCondition.Directories = d
	return f
}

func (f *fileReactionConditionBuilder) Build() *reaction.FileReactionCondition {
	return f.FileTriggerCondition
}

func newFileReactionActionBuilder() FileReactionActionBuilder {
	return &fileReactionActionBuilder{Action: &reaction.FileReactionAction{}}
}

// ZipEntry adds a zip entry into the resulting zip file. f should be an absolute path to a file/folder on the system. pathPrefix is a relative path
// which the resulting file will be placed into in the zip file. If pathPrefix is not set, the file is placed into the root of the zip.
//
//    fi := "/my/file.txt"
//    pp := "path/prefix"
//    Zip(fi, pp) // adds entry in the zip folder with the path "path/prefix/file.txt"
func ZipEntry(f, pathPrefix string) *actions.ZipFileEntry {
	return &actions.ZipFileEntry{
		From:       f,
		PathPrefix: pathPrefix,
	}
}

func ZipFolder(to string, entries ...*actions.ZipFileEntry) FileReactionActionBuilder {
	return newFileReactionActionBuilder().ZipFolder(to, entries...)
}

func UploadSaveFile(from string, filename string) FileReactionActionBuilder {
	fn := fileutils.Filename(filename)
	if !strings.HasSuffix(fn, "_autosave") {
		filename = fn + "_autosave"
	} else {
		filename = fn
	}

	return newFileReactionActionBuilder().Upload(from, filename, "saves")
}

func NewFileReactionActionBuilder() FileReactionActionBuilder {
	return &fileReactionActionBuilder{
		Action: &reaction.FileReactionAction{},
	}
}

type fileReactionActionBuilder struct {
	Action *reaction.FileReactionAction
}

func (f *fileReactionActionBuilder) Upload(from, filename string, folder string) FileReactionActionBuilder {
	f.Action = &reaction.FileReactionAction{
		Upload: &actions.UploadFile{
			From: &actions.UploadFile_Source{Path: from},
			To: &filesystem.FileLocation{BucketFile: &filesystem.BucketFile{
				Name:   filename,
				Folder: folder,
			}},
		},
	}

	return f
}

func (f *fileReactionActionBuilder) ZipFolder(to string, entries ...*actions.ZipFileEntry) FileReactionActionBuilder {
	f.Action = &reaction.FileReactionAction{
		Zip: &actions.ZipFile{
			From: &actions.ZipFile_Source{
				Files: entries,
			},
			To: &actions.ZipFile_Destination{
				Path: to,
			},
		},
	}
	return f
}

func (f *fileReactionActionBuilder) Build() *reaction.FileReactionAction {
	return f.Action
}
