package provideractions

import (
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/diazo/pkg/fileutils"
	"github.com/hostfactor/diazo/pkg/ptr"
	"strings"
)

type RunPhaseBuilder interface {
	When(when ...FileTriggerConditionBuilder) FileTriggerBuilder

	Gid(i int) RunPhaseBuilder

	Uid(i int) RunPhaseBuilder

	Build() *blueprint.RunPhase
}

type FileTriggerBuilder interface {
	Then(then ...FileTriggerActionBuilder) RunPhaseBuilder

	build() *blueprint.FileTrigger
}

func NewRunPhaseBuilder() RunPhaseBuilder {
	return &runPhaseBuilder{RunPhase: &blueprint.RunPhase{}}
}

type runPhaseBuilder struct {
	RunPhase        *blueprint.RunPhase
	TriggerBuilders []FileTriggerBuilder
}

type fileTriggerBuilder struct {
	Whens           []FileTriggerConditionBuilder
	Thens           []FileTriggerActionBuilder
	RunPhaseBuilder *runPhaseBuilder
}

func (f *fileTriggerBuilder) build() *blueprint.FileTrigger {
	whens := make([]*blueprint.FileTriggerCondition, 0, len(f.Whens))
	for _, v := range f.Whens {
		whens = append(whens, v.Build())
	}

	thens := make([]*blueprint.FileTriggerAction, 0, len(f.Thens))
	for _, v := range f.Thens {
		thens = append(thens, v.Build())
	}

	return &blueprint.FileTrigger{
		When: whens,
		Then: thens,
	}
}

func (f *fileTriggerBuilder) Then(then ...FileTriggerActionBuilder) RunPhaseBuilder {
	f.Thens = then
	return f.RunPhaseBuilder
}

func (r *runPhaseBuilder) When(when ...FileTriggerConditionBuilder) FileTriggerBuilder {
	trigger := &fileTriggerBuilder{
		Whens:           when,
		RunPhaseBuilder: r,
	}
	r.TriggerBuilders = append(r.TriggerBuilders, trigger)

	return trigger
}

func (r *runPhaseBuilder) Gid(i int) RunPhaseBuilder {
	r.RunPhase.Gid = ptr.Int64(int64(i))
	return r
}

func (r *runPhaseBuilder) Uid(i int) RunPhaseBuilder {
	r.RunPhase.Uid = ptr.Int64(int64(i))
	return r
}

func (r *runPhaseBuilder) Build() *blueprint.RunPhase {
	triggers := make([]*blueprint.FileTrigger, 0, len(r.TriggerBuilders))
	for _, v := range r.TriggerBuilders {
		triggers = append(triggers, v.build())
	}

	r.TriggerBuilders = nil
	r.RunPhase.Triggers = triggers
	return r.RunPhase
}

type FileTriggerConditionBuilder interface {
	// Matches resolves to true if the file change matches the specified matcher.
	Matches(fm *filesystem.FileMatcher) FileTriggerConditionBuilder

	// DoesntMatch resolves to true if the file change does not match the specified matcher.
	DoesntMatch(fm *filesystem.FileMatcher) FileTriggerConditionBuilder

	// Op resolves to true if one of the specified blueprint.FileChangeOp match.
	Op(op ...blueprint.FileChangeOp) FileTriggerConditionBuilder

	// Directories adds a directory to be watched for all conditions.
	Directories(d ...string) FileTriggerConditionBuilder
	Build() *blueprint.FileTriggerCondition
}

func NewFileTriggerCondition() FileTriggerConditionBuilder {
	return &fileTriggerConditionBuilder{
		FileTriggerCondition: &blueprint.FileTriggerCondition{},
	}
}

type fileTriggerConditionBuilder struct {
	FileTriggerCondition *blueprint.FileTriggerCondition
}

func (f *fileTriggerConditionBuilder) Matches(fm *filesystem.FileMatcher) FileTriggerConditionBuilder {
	f.FileTriggerCondition.Matches = fm
	return f
}

func (f *fileTriggerConditionBuilder) DoesntMatch(fm *filesystem.FileMatcher) FileTriggerConditionBuilder {
	f.FileTriggerCondition.DoesntMatch = fm
	return f
}

func (f *fileTriggerConditionBuilder) Op(op ...blueprint.FileChangeOp) FileTriggerConditionBuilder {
	f.FileTriggerCondition.Op = op
	return f
}

func (f *fileTriggerConditionBuilder) Directories(d ...string) FileTriggerConditionBuilder {
	f.FileTriggerCondition.Directories = d
	return f
}

func (f *fileTriggerConditionBuilder) Build() *blueprint.FileTriggerCondition {
	return f.FileTriggerCondition
}

type FileTriggerActionBuilder interface {
	Upload(from string, filename string, folder string) FileTriggerActionBuilder

	// ZipFolder zips the absolute path f to a zip
	ZipFolder(to string, entries ...*actions.ZipFileEntry) FileTriggerActionBuilder

	Build() *blueprint.FileTriggerAction
}

func newFileTriggerActionBuilder() FileTriggerActionBuilder {
	return &fileTriggerActionBuilder{Action: &blueprint.FileTriggerAction{}}
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

func ZipFolder(to string, entries ...*actions.ZipFileEntry) FileTriggerActionBuilder {
	return newFileTriggerActionBuilder().ZipFolder(to, entries...)
}

func UploadSaveFile(from string, filename string) FileTriggerActionBuilder {
	fn, _ := fileutils.SplitFile(filename)
	if !strings.HasSuffix(fn, "_autosave") {
		filename = fn + "_autosave"
	} else {
		filename = fn
	}

	return newFileTriggerActionBuilder().Upload(from, filename, "saves")
}

type fileTriggerActionBuilder struct {
	Action *blueprint.FileTriggerAction
}

func (f *fileTriggerActionBuilder) Upload(from, filename string, folder string) FileTriggerActionBuilder {
	f.Action = &blueprint.FileTriggerAction{
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

func (f *fileTriggerActionBuilder) ZipFolder(to string, entries ...*actions.ZipFileEntry) FileTriggerActionBuilder {
	f.Action = &blueprint.FileTriggerAction{
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

func (f *fileTriggerActionBuilder) Build() *blueprint.FileTriggerAction {
	return f.Action
}
