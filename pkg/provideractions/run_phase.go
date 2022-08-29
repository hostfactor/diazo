package provideractions

import (
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/blueprint/reaction"
	"github.com/hostfactor/diazo/pkg/ptr"
)

type RunPhaseBuilder interface {
	FileReaction(builder FileReactionBuilder) RunPhaseBuilder
	LogReaction(builder LogReactionBuilder) RunPhaseBuilder

	Gid(i int) RunPhaseBuilder

	Uid(i int) RunPhaseBuilder

	Build() *blueprint.RunPhase
}

type reactionBuilder interface {
	Build() *reaction.Reaction
}

func NewRunPhaseBuilder() RunPhaseBuilder {
	return &runPhaseBuilder{RunPhase: &blueprint.RunPhase{}}
}

type runPhaseBuilder struct {
	RunPhase         *blueprint.RunPhase
	ReactionBuilders []reactionBuilder
}

func (r *runPhaseBuilder) FileReaction(builder FileReactionBuilder) RunPhaseBuilder {
	r.ReactionBuilders = append(r.ReactionBuilders, builder)
	return r
}

func (r *runPhaseBuilder) LogReaction(builder LogReactionBuilder) RunPhaseBuilder {
	r.ReactionBuilders = append(r.ReactionBuilders, builder)
	return r
}

func (r *runPhaseBuilder) Gid(i int) RunPhaseBuilder {
	r.RunPhase.Gid = ptr.Int64(int64(i))
	return r
}

func (r *runPhaseBuilder) Uid(i int) RunPhaseBuilder {
	r.RunPhase.Uid = ptr.Int64(int64(i))
	return r
}

func (r *runPhaseBuilder) Build() *blueprint.RunPhase {
	reactions := make([]*reaction.Reaction, 0, len(r.ReactionBuilders))
	for _, v := range r.ReactionBuilders {
		reactions = append(reactions, v.Build())
	}

	r.ReactionBuilders = nil
	r.RunPhase.Reactions = reactions
	return r.RunPhase
}
