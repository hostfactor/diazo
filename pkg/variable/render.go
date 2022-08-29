package variable

import (
	"github.com/flosch/pongo2/v6"
	"strings"
)

func RenderString(og string, store Store, entries ...*Entry) string {
	og = replaceVarsDeprecated(og, store, entries...)
	temp, err := pongo2.FromString(og)
	if err != nil {
		return og
	}

	out, err := temp.Execute(toPongoContext(store, entries...))
	if err != nil {
		return og
	}
	return out
}

func replaceVarsDeprecated(og string, store Store, entries ...*Entry) string {
	s := CombineEntries(store, entries...)
	return strings.NewReplacer(
		"${dir}", s.GetStringValue("dir"),
		"${abs}", s.GetStringValue("abs"),
		"${filename}", s.GetStringValue("filename"),
		"${name}", s.GetStringValue("name"),
		"${ext}", s.GetStringValue("ext"),
	).Replace(og)
}
