package reaction

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/api/go/blueprint/reaction"
	actions2 "github.com/hostfactor/diazo/pkg/actions"
	"github.com/hostfactor/diazo/pkg/fileutils"
	"github.com/hostfactor/diazo/pkg/variable"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"strings"
)

type ExecuteFileOpts struct {
	OnFileChange func(fn string)
	UploadOpts   actions2.UploadOpts
	DownloadOpts actions2.DownloadOpts
}

// ExecuteFile executes the blueprint.FileTrigger using the root. The root is the base path of where to execute the action
// e.g. for download or upload.
func ExecuteFile(ctx context.Context, store variable.Store, root string, ft *reaction.FileReaction, opts ExecuteFileOpts) (context.Context, error) {
	logrus.WithField("data", ft.String()).Debug("Starting file triggers.")
	c, err := WatchFile(ctx, func(event fsnotify.Event) {
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
		return nil, err
	}

	return c, nil
}

// ExecuteFileReactionAction executes the reaction.FileReactionAction using the root. The root is the base path of where
// to execute the action e.g. for download or upload.
func ExecuteFileReactionAction(fp, root string, s variable.Store, action *reaction.FileReactionAction, opts ExecuteFileOpts) error {
	dir, filename := filepath.Split(fp)
	name, ext := fileutils.SplitFile(filename)
	templateEntries := variable.FileReactionTemplateDataEntries(&reaction.FileReactionTemplateData{
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
			v.To = variable.RenderString(t, s, templateEntries...)
		}

		if t := v.GetFrom(); t != nil {
			if t.Directory != "" {
				t.Directory = variable.RenderString(t.Directory, s, templateEntries...)
			}
		}

		logrus.WithField("data", v.String()).Debug("Triggering rename.")
		return actions2.Rename(v)
	} else if v := action.GetDownload(); v != nil {
		if t := v.GetTo(); t != "" {
			v.To = variable.RenderString(t, s, templateEntries...)
		}

		logrus.WithField("data", v.String()).Debug("Triggering download.")
		return actions2.Download(root, v, opts.DownloadOpts)
	} else if v := action.GetExtract(); v != nil {
		if t := v.GetTo(); t != "" {
			v.To = variable.RenderString(t, s, templateEntries...)
		}

		if t := v.GetFrom(); t != nil {
			if t.Directory != "" {
				t.Directory = variable.RenderString(t.Directory, s, templateEntries...)
			}
		}

		logrus.WithField("data", v.String()).Debug("Triggering extract.")
		return actions2.Extract(v)
	} else if v := action.GetUnzip(); v != nil {
		if f := v.GetFrom(); f != "" {
			v.From = variable.RenderString(f, s, templateEntries...)
		}

		if f := v.GetTo(); f != "" {
			v.To = variable.RenderString(f, s, templateEntries...)
		}

		logrus.WithField("data", v.String()).Debug("Triggering unzip.")
		return actions2.Unzip(v)
	} else if v := action.GetZip(); v != nil {
		if p := v.GetTo().GetPath(); p != "" {
			v.To = &actions.ZipFile_Destination{Path: variable.RenderString(p, s, templateEntries...)}
		}
		if d := v.GetFrom().GetDirectory(); d != "" {
			v.From = &actions.ZipFile_Source{Directory: variable.RenderString(d, s, templateEntries...)}
		}

		for i, entry := range v.GetFrom().GetFiles() {
			if entry.From != "" {
				v.From.Files[i].From = variable.RenderString(entry.From, s, templateEntries...)
			}
		}

		logrus.WithField("data", v.String()).Debug("Triggering zip.")
		return actions2.Zip(v)
	} else if v := action.GetUpload(); v != nil {
		if v.GetFrom().GetPath() != "" {
			v.From = &actions.UploadFile_Source{Path: variable.RenderString(v.GetFrom().GetPath(), s, templateEntries...)}
		}

		if to := v.GetTo().GetBucketFile(); to != nil {
			to.Name = variable.RenderString(to.Name, s, templateEntries...)
		}

		logrus.WithField("data", v.String()).Debug("Triggering upload.")
		return actions2.Upload(root, v, opts.UploadOpts)
	} else if v := action.GetMove(); v != nil {
		if v.GetFrom().GetDirectory() != "" {
			v.From.Directory = variable.RenderString(v.GetFrom().GetDirectory(), s, templateEntries...)
		}

		if to := v.GetTo(); to != "" {
			v.To = variable.RenderString(v.To, s, templateEntries...)
		}

		logrus.WithField("data", v.String()).Debug("Triggering move.")
		return actions2.Move(v)
	}

	return nil
}

type WatchFileFunc func(event fsnotify.Event)

func WatchFile(ctx context.Context, callback WatchFileFunc, conds ...*reaction.FileReactionCondition) (context.Context, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancelCause(ctx)

	go func() {
		defer func() {
			cancel(err)
		}()
		for {
			select {
			case <-ctx.Done():
				err = context.Cause(ctx)
				return
			case event, ok := <-watcher.Events:
				logrus.WithField("fp", event.Name).WithField("op", event.Op).Trace("Detected file change.")
				if !ok {
					err = fmt.Errorf("watch closed")
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
			err = os.MkdirAll(d, os.ModePerm)
			if err != nil {
				logrus.WithError(err).WithField("directory", d).Warn("Failed to create dir. Waiting 10 seconds for the dir to appear.")
			}
			logrus.WithField("directory", d).Debug("Watching directory for triggers.")

			err = watcher.Add(d)
			if err != nil {
				logrus.WithError(err).WithField("directory", d).Error("Failed to watch directory.")
				return nil, err
			}
		}
	}

	return ctx, nil
}

// FileReactionCondition checks if a condition matches an absolute filepath.
func FileReactionCondition(ev fsnotify.Event, c *reaction.FileReactionCondition) (matched bool) {
	matched = true
	if m := c.GetMatches(); m != nil {
		matched = matched && actions2.MatchPath(ev.Name, m)
	}

	if len(c.GetOp()) > 0 {
		matched = matched && matchesOp(ev.Op, c.GetOp()...)
	}

	if m := c.GetDoesntMatch(); m != nil {
		matched = matched && !actions2.MatchPath(ev.Name, m)
	}

	return
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
