package trigger

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/diazo/pkg/fileactions"
	"github.com/hostfactor/diazo/pkg/fileutils"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"path/filepath"
	"regexp"
	"strings"
)

type ExecuteOpts struct {
	UploadOpts   fileactions.UploadOpts
	DownloadOpts fileactions.DownloadOpts
}

type ExecuteFileOpts struct {
	ExecuteOpts
	OnFileChange func(fn string)
}

func ExecuteFile(ctx context.Context, instanceId, userId, title string, ft *blueprint.FileTrigger, opts ExecuteFileOpts) error {
	logrus.WithField("data", ft.String()).Debug("Starting file triggers.")
	err := Watch(ctx, func(event fsnotify.Event) {
		if opts.OnFileChange != nil {
			opts.OnFileChange(event.Name)
		}
		for _, v := range ft.GetThen() {
			err := ExecuteFileTriggerAction(event.Name, instanceId, userId, title, v, opts.ExecuteOpts)
			if err != nil {
				logrus.WithError(err).WithField("file", event.Name).Error("Failed to execute action.")
			}
		}
	}, ft.GetWhen()...)
	if err != nil {
		return err
	}

	return nil
}

func ExecuteFileTriggerAction(fp, instanceId, userId, title string, action *blueprint.FileTriggerAction, opts ExecuteOpts) error {
	dir, filename := filepath.Split(fp)
	name, ext := fileutils.SplitFile(filename)
	templateData := &blueprint.FileTriggerTemplateData{
		Dir:      filepath.Clean(dir),
		Filename: filename,
		Ext:      strings.TrimPrefix(ext, "."),
		Abs:      fp,
		Name:     name,
	}
	logrus.WithField("data", templateData.String()).WithField("action", action.String()).Debug("Executing file trigger.")

	// Required so the template rendering doesn't update the original.
	action = proto.Clone(action).(*blueprint.FileTriggerAction)

	if v := action.GetRename(); v != nil {
		if t := v.GetTo(); t != "" {
			v.To = renderTemplateString(t, templateData)
		}

		if t := v.GetFrom(); t != nil {
			if t.Directory != "" {
				t.Directory = renderTemplateString(t.Directory, templateData)
			}
		}

		logrus.WithField("data", v.String()).Debug("Triggering rename.")
		return fileactions.Rename(v)
	} else if v := action.GetDownload(); v != nil {
		if t := v.GetTo(); t != "" {
			v.To = renderTemplateString(t, templateData)
		}

		logrus.WithField("data", v.String()).Debug("Triggering download.")
		return fileactions.Download(instanceId, userId, title, v, opts.DownloadOpts)
	} else if v := action.GetExtract(); v != nil {
		if t := v.GetTo(); t != "" {
			v.To = renderTemplateString(t, templateData)
		}

		if t := v.GetFrom(); t != nil {
			if t.Directory != "" {
				t.Directory = renderTemplateString(t.Directory, templateData)
			}
		}

		logrus.WithField("data", v.String()).Debug("Triggering extract.")
		return fileactions.Extract(v)
	} else if v := action.GetUnzip(); v != nil {
		if f := v.GetFrom(); f != "" {
			v.From = renderTemplateString(f, templateData)
		}

		if f := v.GetTo(); f != "" {
			v.To = renderTemplateString(f, templateData)
		}

		logrus.WithField("data", v.String()).Debug("Triggering unzip.")
		return fileactions.Unzip(v)
	} else if v := action.GetZip(); v != nil {
		if p := v.GetPath(); p != "" {
			v.To = &actions.ZipFile_Path{Path: renderTemplateString(p, templateData)}
		}
		if d := v.GetDirectory(); d != "" {
			v.From = &actions.ZipFile_Directory{Directory: renderTemplateString(d, templateData)}
		}

		logrus.WithField("data", v.String()).Debug("Triggering zip.")
		return fileactions.Zip(v)
	} else if v := action.GetUpload(); v != nil {
		if v.GetPath() != "" {
			v.From = &actions.UploadFile_Path{Path: renderTemplateString(v.GetPath(), templateData)}
		}

		switch typ := v.GetTo().Loc.(type) {
		case *filesystem.FileLocation_BucketFile:
			typ.BucketFile.Name = renderTemplateString(typ.BucketFile.Name, templateData)
		}

		logrus.WithField("data", v.String()).Debug("Triggering upload.")
		return fileactions.Upload(instanceId, userId, title, v, opts.UploadOpts)
	}

	return nil
}

type WatchFunc func(event fsnotify.Event)

func Watch(ctx context.Context, callback WatchFunc, conds ...*blueprint.FileTriggerCondition) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				logrus.WithField("fp", event.Name).WithField("op", event.Op).Trace("Detected file change.")
				if !ok {
					return
				}

				matches := false
				for _, v := range conds {
					matches = matches || FileTriggerCondition(event, v)
				}

				if matches {
					callback(event)
				}
			}
		}
	}()

	for _, v := range conds {
		for _, d := range v.GetDirectories() {
			logrus.WithField("directory", d).Debug("Watching directory for triggers.")
			err := watcher.Add(d)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ExecuteSetupAction(instanceId, userId, title string, act *blueprint.SetupAction, opts ExecuteOpts) error {
	if v := act.GetUnzip(); v != nil {
		return fileactions.Unzip(v)
	} else if v := act.GetRename(); v != nil {
		return fileactions.Rename(v)
	} else if v := act.GetExtract(); v != nil {
		return fileactions.Extract(v)
	} else if v := act.GetDownload(); v != nil {
		return fileactions.Download(instanceId, userId, title, v, opts.DownloadOpts)
	}
	return nil
}

// FileTriggerCondition checks if a condition matches an absolute filepath.
func FileTriggerCondition(ev fsnotify.Event, c *blueprint.FileTriggerCondition) (matched bool) {
	matched = true
	if m := c.GetMatches(); m != nil {
		matched = matched && fileactions.MatchPath(ev.Name, m)
	}

	if len(c.GetOp()) > 0 {
		matched = matched && matchesOp(ev.Op, c.GetOp()...)
	}

	if m := c.GetDoesntMatch(); m != nil {
		matched = matched && !fileactions.MatchPath(ev.Name, m)
	}

	return
}

var templateVarRegex = regexp.MustCompile("\\${(abs|dir|filename|name|ext)}")

func ContainsTemplateVariable(s string) bool {
	return templateVarRegex.MatchString(s)
}

func matchesOp(op fsnotify.Op, bp ...blueprint.FileChangeOp) bool {
	m := map[blueprint.FileChangeOp]bool{}
	for _, v := range bp {
		m[v] = true
	}
	return op&fsnotify.Write == fsnotify.Write && m[blueprint.FileChangeOp_FILE_CHANGE_OP_UPDATE] ||
		op&fsnotify.Create == fsnotify.Create && m[blueprint.FileChangeOp_FILE_CHANGE_OP_CREATE] ||
		op&fsnotify.Remove == fsnotify.Remove && m[blueprint.FileChangeOp_FILE_CHANGE_OP_DELETE]
}

func renderTemplateString(s string, data *blueprint.FileTriggerTemplateData) string {
	return strings.NewReplacer("${dir}", data.GetDir(), "${abs}", data.GetAbs(), "${filename}", data.GetFilename(), "${name}", data.GetName(), "${ext}", data.GetExt()).Replace(s)
}