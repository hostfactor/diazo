package variable

import (
	"fmt"
	"github.com/flosch/pongo2/v6"
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/blueprint/reaction"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
	"sync"
)

func NewStore(vars ...*blueprint.Variable) Store {
	out := &store{}
	out.AddVariable(vars...)

	return out
}

func CombineEntries(s Store, e ...*Entry) EntryStore {
	m := map[string]*Entry{}
	for i, v := range e {
		m[v.Key] = e[i]
	}

	return &combinedStore{
		Store:    s,
		EntryMap: m,
	}
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

func VariableEntries(vars ...*blueprint.Variable) []*Entry {
	e := make([]*Entry, 0, len(vars))
	for _, v := range vars {
		e = append(e, &Entry{
			Key:         v.GetName(),
			Val:         v.GetValue(),
			DisplayName: v.GetDisplayName(),
		})
	}
	return e
}

type EntryStore interface {
	StringValueGetter
}

type StringValueGetter interface {
	GetStringValue(key string) string
}

type Store interface {
	fmt.Stringer
	StringValueGetter

	Get(key string) *blueprint.Variable
	RemoveVariable(vars ...*blueprint.Variable)
	AddFileTemplateData(d ...*reaction.FileReactionTemplateData)
	AddLogTemplateData(d ...*reaction.LogReactionTemplateData)
	AddVariable(vars ...*blueprint.Variable)
	AddEntries(entries ...*Entry)
	Range(f func(key string, value *blueprint.Variable) bool)
	Len() int
}

type Entry struct {
	Key         string
	Val         interface{}
	DisplayName string
}

func NewEntry(key string, val interface{}) *Entry {
	return &Entry{
		Key: key,
		Val: val,
	}
}

type storeValue struct {
	Variable *blueprint.Variable
	RawValue interface{}
}

type combinedStore struct {
	Store
	EntryMap map[string]*Entry
}

func (c *combinedStore) GetStringValue(key string) string {
	out := c.Store.GetStringValue(key)
	if out != "" {
		return out
	}

	o := c.EntryMap[key]
	if o != nil {
		s, _ := o.Val.(string)
		return s
	}

	return ""
}

type store struct {
	Map sync.Map
}

func (s *store) Get(key string) *blueprint.Variable {
	out, ok := s.Map.Load(key)
	if !ok {
		return nil
	}
	return out.(*storeValue).Variable
}

func (s *store) Len() int {
	l := 0
	s.Range(func(_ string, _ *blueprint.Variable) bool {
		l++
		return true
	})
	return l
}

func (s *store) AddEntries(entries ...*Entry) {
	for _, v := range entries {
		out := &storeValue{
			Variable: &blueprint.Variable{
				Name:        v.Key,
				Value:       fmt.Sprintf("%v", v.Val),
				DisplayName: v.DisplayName,
			},
			RawValue: v.Val,
		}

		if out.Variable.DisplayName == "" {
			out.Variable.DisplayName = cases.Title(language.English, cases.Compact).String(v.Key)
		}

		s.Map.Store(v.Key, out)
	}
}

func (s *store) Range(f func(key string, value *blueprint.Variable) bool) {
	s.Map.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(*storeValue).Variable)
	})
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

func (s *store) GetStringValue(key string) string {
	out, _ := s.Map.Load(key)
	if out == nil {
		return ""
	}

	return out.(*storeValue).Variable.GetValue()
}

func (s *store) AddLogTemplateData(d ...*reaction.LogReactionTemplateData) {
	for _, v := range d {
		s.AddEntries(LogReactionTemplateDataEntries(v)...)
	}
}

func (s *store) AddVariable(vars ...*blueprint.Variable) {
	s.AddEntries(VariableEntries(vars...)...)
}

func (s *store) RemoveVariable(vars ...*blueprint.Variable) {
	for _, v := range vars {
		s.Map.Delete(v.GetName())
	}
}

func (s *store) rawRange(f func(key string, val *storeValue) bool) {
	s.Map.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(*storeValue))
	})
}

func toPongoContext(s Store, entries ...*Entry) pongo2.Context {
	ctx := pongo2.Context{}
	s.(*store).rawRange(func(k string, v *storeValue) bool {
		ctx[k] = v.RawValue
		return true
	})

	for _, v := range entries {
		ctx[v.Key] = v.Val
	}

	return ctx
}
