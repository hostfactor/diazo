package variable

import (
	"github.com/flosch/pongo2/v6"
	"strings"
)

func RenderString(og string, store Store, entries ...*Entry) string {
	og = replaceVarsDeprecated(og, store)
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

func replaceVarsDeprecated(og string, store Store) string {
	return strings.NewReplacer(
		"${dir}", store.GetString("dir"),
		"${abs}", store.GetString("abs"),
		"${filename}", store.GetString("filename"),
		"${name}", store.GetString("name"),
		"${ext}", store.GetString("ext"),
	).Replace(og)
}
