package reaction

import (
	"context"
	"github.com/hostfactor/api/go/app"
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/api/go/blueprint/reaction"
	diazoactions "github.com/hostfactor/diazo/pkg/actions"
	"github.com/hostfactor/diazo/pkg/variable"
	"github.com/nxadm/tail"
	"github.com/sirupsen/logrus"
	"regexp"
)

type ExecuteLogOpts struct {
	Watcher WatchLogFunc

	// Called whenever the server is set to change its status. This func could be called with the same status multiple times.
	// It is up to the caller to track these status changes.
	OnStatusChange StatusChangeFunc
}

type WatchLogFunc func(ll LogLine)

type StatusChangeFunc func(s actions.SetStatus_Status)

type LogLine struct {
	Text string
	Num  int
}

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
				callback(LogLine{
					Text: line.Text,
					Num:  line.Num,
				})
			}
		}
	}()

	return nil
}

func ExecuteLog(ctx context.Context, logPath string, store variable.Store, appClient app.AppServiceClient, rx []*reaction.LogReaction, opts ExecuteLogOpts) error {
	matchers, err := CompileLogReactions(rx...)
	if err != nil {
		return err
	}

	return WatchLog(ctx, logPath, func(ll LogLine) {
		err := ReactToLog(ll, store, appClient, matchers, opts)
		if err != nil {
			logrus.WithError(err).Error("Failed to execute log action.")
		}
	})
}

func ReactToLog(l LogLine, store variable.Store, appClient app.AppServiceClient, rx []*CompiledLogReaction, opts ExecuteLogOpts) error {
	if opts.Watcher != nil {
		opts.Watcher(l)
	}

	matches := AllMatches(l.Text, rx...)
	if len(matches) == 0 {
		return nil
	}

	for _, match := range matches {
		err := ExecuteLogActions(l.Text, appClient, store, match.RegexMatches, opts, match.Reaction.Then...)
		if err != nil {
			return err
		}
	}
	return nil
}

type CompiledLogReaction struct {
	When []*CompiledLogCondition
	Then []*reaction.LogReactionAction
}

type CompiledLogCondition struct {
	Regex *regexp.Regexp
	Cond  *reaction.LogReactionCondition
}

func (c *CompiledLogCondition) Matches(line string) []string {
	if c.Regex != nil {
		m := c.Regex.FindStringSubmatch(line)
		return m
	}
	return nil
}

type LogMatch struct {
	Reaction     *CompiledLogReaction
	RegexMatches []string
}

func AllMatches(line string, reactions ...*CompiledLogReaction) []*LogMatch {
	matches := make([]*LogMatch, 0, len(reactions))
	for i, r := range reactions {
		for _, v := range r.When {
			out := v.Matches(line)
			if len(out) > 0 {
				matches = append(matches, &LogMatch{
					Reaction:     reactions[i],
					RegexMatches: out,
				})
			}
		}
	}
	return matches
}

func ExecuteLogActions(line string, appClient app.AppServiceClient, store variable.Store, m []string, opts ExecuteLogOpts, a ...*reaction.LogReactionAction) error {
	if len(m) == 0 {
		return nil
	}

	for _, act := range a {
		logrus.
			WithField("variable_store", store.String()).
			WithField("action", act.String()).
			WithField("line", line).
			WithField("matches", m).
			Debug("Executing log action.")
		if v := act.GetSetVariable(); v != nil {
			if len(m) > 1 {
				m = m[1:]
			}
			err := diazoactions.SetVariable(appClient, store, v, variable.LogReactionTemplateDataEntries(&reaction.LogReactionTemplateData{
				Line:       line,
				Matches:    m,
				FirstMatch: m[0],
			})...)
			if err != nil {
				return err
			}
		}

		if v := act.GetSetStatus(); v != nil {
			if opts.OnStatusChange != nil {
				opts.OnStatusChange(v.GetStatus())
			}
		}
	}
	return nil
}

func CompileLogReactions(reactions ...*reaction.LogReaction) ([]*CompiledLogReaction, error) {
	out := make([]*CompiledLogReaction, 0, len(reactions))
	for _, v := range reactions {
		conds, err := CompileLogConditions(v.GetWhen()...)
		if err != nil {
			return nil, err
		}
		out = append(out, &CompiledLogReaction{
			When: conds,
			Then: v.GetThen(),
		})
	}
	return out, nil
}

func CompileLogConditions(conds ...*reaction.LogReactionCondition) ([]*CompiledLogCondition, error) {
	out := make([]*CompiledLogCondition, 0, len(conds))
	for i, v := range conds {
		if r := v.GetMatches().GetRegex(); r != "" {
			re, err := regexp.Compile(v.GetMatches().GetRegex())
			if err != nil {
				return nil, err
			}
			out = append(out, &CompiledLogCondition{
				Cond:  conds[i],
				Regex: re,
			})
		}
	}
	return out, nil
}

// ReadyCheckToLogReaction maintains backwards compatibility with the old ready check system and the new one.
func ReadyCheckToLogReaction(rc *blueprint.ReadyCheck) *reaction.LogReaction {
	return &reaction.LogReaction{
		When: []*reaction.LogReactionCondition{
			{
				Matches: &reaction.LogMatcher{
					Regex: rc.GetRegex(),
				},
			},
		},
		Then: []*reaction.LogReactionAction{
			{
				SetStatus: &actions.SetStatus{
					Status: actions.SetStatus_ready,
				},
			},
		},
	}
}
