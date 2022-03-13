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
	Matches(fm FileMatcher) FileTriggerConditionBuilder

	// DoesntMatch resolves to true if the file change does not match the specified matcher.
	DoesntMatch(fm FileMatcher) FileTriggerConditionBuilder

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

func (f *fileTriggerConditionBuilder) Matches(fm FileMatcher) FileTriggerConditionBuilder {
	f.FileTriggerCondition.Matches = fm.FileMatcher()
	return f
}

func (f *fileTriggerConditionBuilder) DoesntMatch(fm FileMatcher) FileTriggerConditionBuilder {
	f.FileTriggerCondition.DoesntMatch = fm.FileMatcher()
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

	// ZipFolder zips the absolute path to the dir and moves the zip file to the absolute path.
	ZipFolder(dir, to string) FileTriggerActionBuilder

	Build() *blueprint.FileTriggerAction
}

func newFileTriggerActionBuilder() FileTriggerActionBuilder {
	return &fileTriggerActionBuilder{Action: &blueprint.FileTriggerAction{}}
}

func ZipFolder(dir, to string) FileTriggerActionBuilder {
	return newFileTriggerActionBuilder().ZipFolder(dir, to)
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
		Action: &blueprint.FileTriggerAction_Upload{
			Upload: &actions.UploadFile{
				From: &actions.UploadFile_Path{Path: from},
				To: &filesystem.FileLocation{Loc: &filesystem.FileLocation_BucketFile{BucketFile: &filesystem.BucketFile{
					Name:   filename,
					Folder: folder,
				}}},
			},
		},
	}

	return f
}

func (f *fileTriggerActionBuilder) ZipFolder(dir, to string) FileTriggerActionBuilder {
	f.Action = &blueprint.FileTriggerAction{
		Action: &blueprint.FileTriggerAction_Zip{
			Zip: &actions.ZipFile{
				From: &actions.ZipFile_Directory{Directory: dir},
				To:   &actions.ZipFile_Path{Path: to},
			},
		},
	}
	return f
}

func (f *fileTriggerActionBuilder) Build() *blueprint.FileTriggerAction {
	return f.Action
}
