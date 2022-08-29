package reaction

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/hostfactor/api/go/blueprint"
	fileactions2 "github.com/hostfactor/diazo/pkg/actions/fileactions"
	"math"
	"sync"
	"time"
)

type ExecuteOpts struct {
	File ExecuteFileOpts
	Log  ExecuteLogOpts
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
