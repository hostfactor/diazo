package reaction

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/diazo/pkg/actions"
	"math"
	"sync"
	"time"
)

type ExecuteOpts struct {
	File ExecuteFileOpts
	Log  ExecuteLogOpts
}

func ExecuteSetupAction(ctx context.Context, folder string, act *blueprint.SetupAction, opts ExecuteOpts) error {
	if v := act.GetUnzip(); v != nil {
		return actions.Unzip(v)
	} else if v := act.GetRename(); v != nil {
		return actions.Rename(v)
	} else if v := act.GetExtract(); v != nil {
		return actions.Extract(v)
	} else if v := act.GetDownload(); v != nil {
		return actions.Download(folder, v, opts.File.DownloadOpts)
	} else if v := act.GetMove(); v != nil {
		return actions.Move(v)
	} else if v := act.GetShell(); v != nil {
		_, err := actions.Shell(ctx, v)
		return err
	}
	return nil
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
