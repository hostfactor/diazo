package provideractions

import (
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/api/go/blueprint/reaction"
)

func NewLogReaction() LogReactionBuilder {
	return &logReactionBuilder{}
}

type LogReactionBuilder interface {
	reactionBuilder
	When(b ...LogReactionConditionBuilder) LogReactionBuilder
	Then(b ...LogReactionActionBuilder) LogReactionBuilder
}

func NewLogReactionCondition() LogReactionConditionBuilder {
	return &logReactionConditionBuilder{
		Cond: &reaction.LogReactionCondition{},
	}
}

func NewLogReactionAction() LogReactionActionBuilder {
	return &logReactionActionBuilder{
		Act: &reaction.LogReactionAction{},
	}
}

type LogReactionConditionBuilder interface {
	MatchesRegex(regex string) LogReactionConditionBuilder
	Build() *reaction.LogReactionCondition
}

type LogReactionActionBuilder interface {
	SetStatus(status actions.SetStatus_Status) LogReactionActionBuilder
	SetVariable(v *actions.SetVariable) LogReactionActionBuilder
	Build() *reaction.LogReactionAction
}

type logReactionBuilder struct {
	Whens []LogReactionConditionBuilder
	Thens []LogReactionActionBuilder
}

func (l *logReactionBuilder) Build() *reaction.Reaction {
	react := &reaction.Reaction{
		LogReaction: &reaction.LogReaction{
			When: make([]*reaction.LogReactionCondition, 0, len(l.Whens)),
			Then: make([]*reaction.LogReactionAction, 0, len(l.Thens)),
		},
	}
	for _, v := range l.Whens {
		react.LogReaction.When = append(react.LogReaction.When, v.Build())
	}
	for _, v := range l.Thens {
		react.LogReaction.Then = append(react.LogReaction.Then, v.Build())
	}

	return react
}

func (l *logReactionBuilder) When(b ...LogReactionConditionBuilder) LogReactionBuilder {
	l.Whens = append(l.Whens, b...)
	return l
}

func (l *logReactionBuilder) Then(b ...LogReactionActionBuilder) LogReactionBuilder {
	l.Thens = append(l.Thens, b...)
	return l
}

type logReactionConditionBuilder struct {
	Cond *reaction.LogReactionCondition
}

func (l *logReactionConditionBuilder) MatchesRegex(regex string) LogReactionConditionBuilder {
	l.Cond.Matches = &reaction.LogMatcher{
		Regex: regex,
	}
	return l
}

func (l *logReactionConditionBuilder) Build() *reaction.LogReactionCondition {
	return l.Cond
}

type logReactionActionBuilder struct {
	Act *reaction.LogReactionAction
}

func (l *logReactionActionBuilder) SetStatus(status actions.SetStatus_Status) LogReactionActionBuilder {
	l.Act.SetStatus = &actions.SetStatus{
		Status: status,
	}
	return l
}

func (l *logReactionActionBuilder) SetVariable(v *actions.SetVariable) LogReactionActionBuilder {
	l.Act.SetVariable = v
	return l
}

func (l *logReactionActionBuilder) Build() *reaction.LogReactionAction {
	return l.Act
}
