package reaction

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/hostfactor/api/go/app"
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/api/go/blueprint/reaction"
	diazoactions "github.com/hostfactor/diazo/pkg/actions"
	fileactions2 "github.com/hostfactor/diazo/pkg/actions/fileactions"
	"github.com/hostfactor/diazo/pkg/fileutils"
	"github.com/hostfactor/diazo/pkg/variable"
	"github.com/nxadm/tail"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type ExecuteOpts struct {
	File ExecuteFileOpts
	Log  ExecuteLogOpts
}

type ExecuteFileOpts struct {
	OnFileChange func(fn string)
	UploadOpts   fileactions2.UploadOpts
	DownloadOpts fileactions2.DownloadOpts
}

type ExecuteLogOpts struct {
	Watcher WatchLogFunc

	// Called whenever the server is set to change its status. This func could be called with the same status multiple times.
	// It is up to the caller to track these status changes.
	OnStatusChange StatusChangeFunc
}

type StatusChangeFunc func(s actions.SetStatus_Status)

func ExecuteLog(ctx context.Context, logPath string, store variable.Store, appClient app.AppServiceClient, rx []*reaction.LogReaction, opts ExecuteLogOpts) error {
	type val struct {
		Regex *regexp.Regexp
		Rx    *reaction.LogReaction
	}
	matchers := make([]val, 0, len(rx))
	for i, v := range rx {
		when := v.GetWhen()
		if when == nil {
			continue
		}

		r := LogReactionConditionToRegex(when)
		if r == nil {
			continue
		}

		matchers = append(matchers, val{
			Regex: r,
			Rx:    rx[i],
		})
	}

	return WatchLog(ctx, logPath, func(line string, lineNum int) {
		if opts.Watcher != nil {
			opts.Watcher(line, lineNum)
		}

		for _, v := range matchers {
			m := v.Regex.FindStringSubmatch(line)
			if len(m) > 0 {
				if act := v.Rx.GetThen().GetSetVariable(); act != nil {
					if len(m) > 1 {
						m = m[1:]
					}
					err := diazoactions.SetVariable(appClient, store, act, variable.LogReactionTemplateDataEntries(&reaction.LogReactionTemplateData{
						Line:       line,
						Matches:    m,
						FirstMatch: m[0],
					})...)
					if err != nil {
						logrus.WithError(err).WithField("regex", v.Regex.String()).Error("Failed to set variable when matching regex.")
					}
				}

				if act := v.Rx.GetThen().GetSetStatus(); act != nil {
					if opts.OnStatusChange != nil {
						opts.OnStatusChange(act.GetStatus())
					}
				}
			}
		}
	})
}

// ExecuteFile executes the blueprint.FileTrigger using the root. The root is the base path of where to execute the action
// e.g. for download or upload.
func ExecuteFile(ctx context.Context, store variable.Store, root string, ft *reaction.FileReaction, opts ExecuteFileOpts) error {
	logrus.WithField("data", ft.String()).Debug("Starting file triggers.")
	err := WatchFile(ctx, func(event fsnotify.Event) {
		if opts.OnFileChange != nil {
			opts.OnFileChange(event.Name)
		}
		for _, v := range ft.GetThen() {
			err := ExecuteFileReactionAction(event.Name, root, store, v, opts)
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

// ExecuteFileReactionAction executes the reaction.FileReactionAction using the root. The root is the base path of where
// to execute the action e.g. for download or upload.
func ExecuteFileReactionAction(fp, root string, s variable.Store, action *reaction.FileReactionAction, opts ExecuteFileOpts) error {
	dir, filename := filepath.Split(fp)
	name, ext := fileutils.SplitFile(filename)
	s.AddFileTemplateData(&reaction.FileReactionTemplateData{
		Dir:      filepath.Clean(dir),
		Filename: filename,
		Ext:      strings.TrimPrefix(ext, "."),
		Abs:      fp,
		Name:     name,
	})
	logrus.WithField("data", s.String()).WithField("action", action.String()).Debug("Executing file trigger.")

	// Required so the template rendering doesn't update the original.
	action = proto.Clone(action).(*reaction.FileReactionAction)

	if v := action.GetRename(); v != nil {
		if t := v.GetTo(); t != "" {
			v.To = variable.RenderString(t, s)
		}

		if t := v.GetFrom(); t != nil {
			if t.Directory != "" {
				t.Directory = variable.RenderString(t.Directory, s)
			}
		}

		logrus.WithField("data", v.String()).Debug("Triggering rename.")
		return fileactions2.Rename(v)
	} else if v := action.GetDownload(); v != nil {
		if t := v.GetTo(); t != "" {
			v.To = variable.RenderString(t, s)
		}

		logrus.WithField("data", v.String()).Debug("Triggering download.")
		return fileactions2.Download(root, v, opts.DownloadOpts)
	} else if v := action.GetExtract(); v != nil {
		if t := v.GetTo(); t != "" {
			v.To = variable.RenderString(t, s)
		}

		if t := v.GetFrom(); t != nil {
			if t.Directory != "" {
				t.Directory = variable.RenderString(t.Directory, s)
			}
		}

		logrus.WithField("data", v.String()).Debug("Triggering extract.")
		return fileactions2.Extract(v)
	} else if v := action.GetUnzip(); v != nil {
		if f := v.GetFrom(); f != "" {
			v.From = variable.RenderString(f, s)
		}

		if f := v.GetTo(); f != "" {
			v.To = variable.RenderString(f, s)
		}

		logrus.WithField("data", v.String()).Debug("Triggering unzip.")
		return fileactions2.Unzip(v)
	} else if v := action.GetZip(); v != nil {
		if p := v.GetTo().GetPath(); p != "" {
			v.To = &actions.ZipFile_Destination{Path: variable.RenderString(p, s)}
		}
		if d := v.GetFrom().GetDirectory(); d != "" {
			v.From = &actions.ZipFile_Source{Directory: variable.RenderString(d, s)}
		}

		for i, entry := range v.GetFrom().GetFiles() {
			if entry.From != "" {
				v.From.Files[i].From = variable.RenderString(entry.From, s)
			}
		}

		logrus.WithField("data", v.String()).Debug("Triggering zip.")
		return fileactions2.Zip(v)
	} else if v := action.GetUpload(); v != nil {
		if v.GetFrom().GetPath() != "" {
			v.From = &actions.UploadFile_Source{Path: variable.RenderString(v.GetFrom().GetPath(), s)}
		}

		if to := v.GetTo().GetBucketFile(); to != nil {
			to.Name = variable.RenderString(to.Name, s)
		}

		logrus.WithField("data", v.String()).Debug("Triggering upload.")
		return fileactions2.Upload(root, v, opts.UploadOpts)
	} else if v := action.GetMove(); v != nil {
		if v.GetFrom().GetDirectory() != "" {
			v.From.Directory = variable.RenderString(v.GetFrom().GetDirectory(), s)
		}

		if to := v.GetTo(); to != "" {
			v.To = variable.RenderString(v.To, s)
		}

		logrus.WithField("data", v.String()).Debug("Triggering move.")
		return fileactions2.Move(v)
	}

	return nil
}

type WatchLogFunc func(line string, lineNum int)

func WatchLog(ctx context.Context, fp string, callback WatchLogFunc) error {
	tailer, err := tail.TailFile(fp, tail.Config{
		Follow: true,
		Logger: logrus.StandardLogger(),
	})
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				err := ctx.Err()
				logrus.WithError(err).WithField("log", fp).Info("Stopping log watcher.")
				return
			case line := <-tailer.Lines:
				callback(line.Text, line.Num)
			}
		}
	}()

	return nil
}

type WatchFileFunc func(event fsnotify.Event)

func WatchFile(ctx context.Context, callback WatchFileFunc, conds ...*reaction.FileReactionCondition) error {
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

				if event.Op&fsnotify.Chmod == fsnotify.Chmod || event.Op&fsnotify.Rename == fsnotify.Rename {
					continue
				}

				matches := false
				for _, v := range conds {
					matches = matches || FileReactionCondition(event, v)
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
		return fileactions2.Unzip(v)
	} else if v := act.GetRename(); v != nil {
		return fileactions2.Rename(v)
	} else if v := act.GetExtract(); v != nil {
		return fileactions2.Extract(v)
	} else if v := act.GetDownload(); v != nil {
		return fileactions2.Download(folder, v, opts.File.DownloadOpts)
	} else if v := act.GetMove(); v != nil {
		return fileactions2.Move(v)
	}
	return nil
}

func LogReactionConditionToRegex(c *reaction.LogReactionCondition) *regexp.Regexp {
	if m := c.GetMatches(); m != nil {
		return logLineMatches(m)
	}

	return nil
}

func logLineMatches(m *reaction.LogMatcher) *regexp.Regexp {
	if rawRegex := m.GetRegex(); rawRegex != "" {
		r, err := regexp.Compile(m.GetRegex())
		if err != nil {
			logrus.WithError(err).WithField("regex", rawRegex).Error("Failed to compile regex.")
			return nil
		}
		return r
	}
	return nil
}

// FileReactionCondition checks if a condition matches an absolute filepath.
func FileReactionCondition(ev fsnotify.Event, c *reaction.FileReactionCondition) (matched bool) {
	matched = true
	if m := c.GetMatches(); m != nil {
		matched = matched && fileactions2.MatchPath(ev.Name, m)
	}

	if len(c.GetOp()) > 0 {
		matched = matched && matchesOp(ev.Op, c.GetOp()...)
	}

	if m := c.GetDoesntMatch(); m != nil {
		matched = matched && !fileactions2.MatchPath(ev.Name, m)
	}

	return
}

var templateVarRegex = regexp.MustCompile("\\${(abs|dir|filename|name|ext)}")

func ContainsTemplateVariable(s string) bool {
	return templateVarRegex.MatchString(s)
}

func matchesOp(op fsnotify.Op, bp ...reaction.FileReactionCondition_FileOp) bool {
	m := map[reaction.FileReactionCondition_FileOp]bool{}
	for _, v := range bp {
		m[v] = true
	}
	return op&fsnotify.Write == fsnotify.Write && m[reaction.FileReactionCondition_update] ||
		op&fsnotify.Create == fsnotify.Create && m[reaction.FileReactionCondition_create] ||
		op&fsnotify.Remove == fsnotify.Remove && m[reaction.FileReactionCondition_delete]
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
		timer := time.NewTimer(math.MaxInt64)
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				if d.Ev != nil {
					out <- *d.Ev
					d.Ev = nil
				}
			case ev, ok := <-c:
				if !ok {
					return
				}

				d.Ev = &ev
				timer.Reset(d.Interval)
			}
		}
	}()

	return out
}
