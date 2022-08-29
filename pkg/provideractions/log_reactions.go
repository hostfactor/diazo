package provideractions

import (
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

type LogReactionConditionBuilder interface {
	Build() *reaction.LogReactionCondition
}

type LogReactionActionBuilder interface {
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
	//TODO implement me
	panic("implement me")
}

func (l *logReactionBuilder) Then(b ...LogReactionActionBuilder) LogReactionBuilder {
	//TODO implement me
	panic("implement me")
}
