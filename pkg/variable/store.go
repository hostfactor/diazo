package variable

import (
	"fmt"
	"github.com/flosch/pongo2/v6"
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/blueprint/reaction"
	"strings"
	"sync"
)

func NewStore(vars ...*blueprint.Variable) Store {
	out := &store{}
	out.AddVariable(vars...)

	return out
}

func LogReactionTemplateDataEntries(l *reaction.LogReactionTemplateData) []*Entry {
	entries := make([]*Entry, 0, 3)
	entries = append(entries, &Entry{
		Key: "line",
		Val: l.GetLine(),
	})
	entries = append(entries, &Entry{
		Key: "first_match",
		Val: l.GetFirstMatch(),
	})
	entries = append(entries, &Entry{
		Key: "matches",
		Val: l.GetMatches(),
	})
	return entries
}

func FileReactionTemplateDataEntries(v *reaction.FileReactionTemplateData) []*Entry {
	entries := make([]*Entry, 0, 5)

	entries = append(entries, &Entry{
		Key: "dir",
		Val: v.GetDir(),
	})
	entries = append(entries, &Entry{
		Key: "name",
		Val: v.GetName(),
	})
	entries = append(entries, &Entry{
		Key: "abs",
		Val: v.GetAbs(),
	})
	entries = append(entries, &Entry{
		Key: "ext",
		Val: v.GetExt(),
	})
	entries = append(entries, &Entry{
		Key: "filename",
		Val: v.GetFilename(),
	})

	return entries
}

type Store interface {
	fmt.Stringer
	GetString(key string) string
	RemoveVariable(vars ...*blueprint.Variable)
	AddFileTemplateData(d ...*reaction.FileReactionTemplateData)
	AddLogTemplateData(d ...*reaction.LogReactionTemplateData)
	AddVariable(vars ...*blueprint.Variable)
	AddEntries(entries ...*Entry)
	Range(f func(key interface{}, value interface{}) bool)
	Len() int
}

type Entry struct {
	Key string
	Val interface{}
}

func NewEntry(key string, val interface{}) *Entry {
	return &Entry{
		Key: key,
		Val: val,
	}
}

type store struct {
	Map sync.Map
}

func (s *store) Len() int {
	l := 0
	s.Range(func(_ interface{}, _ interface{}) bool {
		l++
		return true
	})
	return l
}

func (s *store) AddEntries(entries ...*Entry) {
	for _, v := range entries {
		s.Map.Store(v.Key, v.Val)
	}
}

func (s *store) Range(f func(key interface{}, value interface{}) bool) {
	s.Map.Range(f)
}

func (s *store) String() string {
	out := make([]string, 0, 10)
	s.Map.Range(func(k, v interface{}) bool {
		out = append(out, fmt.Sprintf("%s=%v", k, v))
		return true
	})
	return strings.Join(out, ", ")
}

func (s *store) AddFileTemplateData(d ...*reaction.FileReactionTemplateData) {
	for _, v := range d {
		s.AddEntries(FileReactionTemplateDataEntries(v)...)
	}
}

func (s *store) GetString(key string) string {
	out, _ := s.Map.Load(key)
	if out == nil {
		return ""
	}

	val, _ := out.(string)
	return val
}

func (s *store) AddLogTemplateData(d ...*reaction.LogReactionTemplateData) {
	for _, v := range d {
		s.AddEntries(LogReactionTemplateDataEntries(v)...)
	}
}

func (s *store) AddVariable(vars ...*blueprint.Variable) {
	for _, v := range vars {
		s.Map.Store(v.GetName(), v.GetValue())
	}
}

func (s *store) RemoveVariable(vars ...*blueprint.Variable) {
	for _, v := range vars {
		s.Map.Delete(v.GetName())
	}
}

func toPongoContext(s Store, entries ...*Entry) pongo2.Context {
	ctx := pongo2.Context{}
	s.Range(func(k, v interface{}) bool {
		ctx[k.(string)] = v
		return true
	})

	for _, v := range entries {
		ctx[v.Key] = v.Val
	}

	return ctx
}
