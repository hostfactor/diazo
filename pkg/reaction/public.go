package reaction

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/diazo/pkg/actions"
	"github.com/hostfactor/diazo/pkg/fileutils"
	"github.com/hostfactor/diazo/pkg/ptr"
	"github.com/sirupsen/logrus"
	"math"
	"sync"
	"time"
)

type ExecuteOpts struct {
	File ExecuteFileOpts
	Log  ExecuteLogOpts
	Uid  *int
	Gid  *int
}

func ExecuteSetupAction(ctx context.Context, folder string, act *blueprint.SetupAction, opts ExecuteOpts) (err error) {
	var createdDir string
	if v := act.GetUnzip(); v != nil {
		createdDir = v.To
		err = actions.Unzip(v)
	} else if v := act.GetRename(); v != nil {
		createdDir = v.To
		err = actions.Rename(v)
	} else if v := act.GetExtract(); v != nil {
		createdDir = v.To
		err = actions.Extract(v)
	} else if v := act.GetDownload(); v != nil {
		createdDir = v.To
		err = actions.Download(folder, v, opts.File.DownloadOpts)
	} else if v := act.GetMove(); v != nil {
		createdDir = v.To
		err = actions.Move(v)
	} else if v := act.GetShell(); v != nil {
		_, err = actions.Shell(ctx, v)
	}
	if err != nil {
		return
	}

	if createdDir != "" && (opts.Gid != nil || opts.Uid != nil) {
		er := fileutils.ChownR(createdDir, ptr.Deref(opts.Uid), ptr.Deref(opts.Gid))
		if err != nil {
			logrus.WithError(err).Error("Failed to chown dir.")
			return er
		}
	}

	return
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
