package trigger

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/diazo/pkg/fileactions"
	"github.com/hostfactor/diazo/pkg/fileutils"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type ExecuteOpts struct {
	UploadOpts   fileactions.UploadOpts
	DownloadOpts fileactions.DownloadOpts
}

type ExecuteFileOpts struct {
	ExecuteOpts
	OnFileChange func(fn string)
}

// ExecuteFile executes the blueprint.FileTrigger using the root. The root is the base path of where to execute the action
// e.g. for download or upload.
func ExecuteFile(ctx context.Context, root string, ft *blueprint.FileTrigger, opts ExecuteFileOpts) error {
	logrus.WithField("data", ft.String()).Debug("Starting file triggers.")
	err := Watch(ctx, func(event fsnotify.Event) {
		if opts.OnFileChange != nil {
			opts.OnFileChange(event.Name)
		}
		for _, v := range ft.GetThen() {
			err := ExecuteFileTriggerAction(event.Name, root, v, opts.ExecuteOpts)
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

// ExecuteFileTriggerAction executes the blueprint.FileTriggerAction using the root. The root is the base path of where
// to execute the action e.g. for download or upload.
func ExecuteFileTriggerAction(fp, root string, action *blueprint.FileTriggerAction, opts ExecuteOpts) error {
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
		return fileactions.Download(root, v, opts.DownloadOpts)
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
		if p := v.GetTo().GetPath(); p != "" {
			v.To = &actions.ZipFile_Destination{Path: renderTemplateString(p, templateData)}
		}
		if d := v.GetFrom().GetDirectory(); d != "" {
			v.From = &actions.ZipFile_Source{Directory: renderTemplateString(d, templateData)}
		}

		for i, entry := range v.GetFrom().GetFiles() {
			if entry.From != "" {
				v.From.Files[i].From = renderTemplateString(entry.From, templateData)
			}
		}

		logrus.WithField("data", v.String()).Debug("Triggering zip.")
		return fileactions.Zip(v)
	} else if v := action.GetUpload(); v != nil {
		if v.GetFrom().GetPath() != "" {
			v.From = &actions.UploadFile_Source{Path: renderTemplateString(v.GetFrom().GetPath(), templateData)}
		}

		if to := v.GetTo().GetBucketFile(); to != nil {
			to.Name = renderTemplateString(to.Name, templateData)
		}

		logrus.WithField("data", v.String()).Debug("Triggering upload.")
		return fileactions.Upload(root, v, opts.UploadOpts)
	} else if v := action.GetMove(); v != nil {
		if v.GetFrom().GetDirectory() != "" {
			v.From.Directory = renderTemplateString(v.GetFrom().GetDirectory(), templateData)
		}

		if to := v.GetTo(); to != "" {
			v.To = renderTemplateString(v.To, templateData)
		}

		logrus.WithField("data", v.String()).Debug("Triggering move.")
		return fileactions.Move(v)
	}

	return nil
}

type WatchFunc func(event fsnotify.Event)

var DebounceInterval = 5 * time.Second

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
			case event, ok := <-Debounce(ctx, watcher.Events, DebounceInterval):
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
			_ = os.MkdirAll(d, os.ModePerm)
			logrus.WithField("directory", d).Debug("Watching directory for triggers.")
			err := watcher.Add(d)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ExecuteSetupAction(folder string, act *blueprint.SetupAction, opts ExecuteOpts) error {
	if v := act.GetUnzip(); v != nil {
		return fileactions.Unzip(v)
	} else if v := act.GetRename(); v != nil {
		return fileactions.Rename(v)
	} else if v := act.GetExtract(); v != nil {
		return fileactions.Extract(v)
	} else if v := act.GetDownload(); v != nil {
		return fileactions.Download(folder, v, opts.DownloadOpts)
	} else if v := act.GetMove(); v != nil {
		return fileactions.Move(v)
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

type debouncer struct {
	Ev        *fsnotify.Event
	Triggered time.Time
	Interval  time.Duration
	lock      sync.RWMutex
}

func Debounce(ctx context.Context, c chan fsnotify.Event, dur time.Duration) chan fsnotify.Event {
	deb := &debouncer{Interval: dur}
	return deb.Debounce(ctx, c)
}

func (d *debouncer) Debounce(ctx context.Context, c chan fsnotify.Event) chan fsnotify.Event {
	out := make(chan fsnotify.Event, 1)

	go func() {
		ticker := time.NewTicker(d.Interval / 5)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				d.lock.Lock()
				if d.Ev != nil && time.Now().Sub(d.Triggered) > d.Interval {
					out <- *d.Ev
					d.Ev = nil
				}
				d.lock.Unlock()
			case ev, ok := <-c:
				if !ok {
					return
				}

				d.lock.Lock()
				d.Triggered = time.Now()
				d.Ev = &ev
				d.lock.Unlock()
			}
		}
	}()

	return out
}
