package actions

import (
	"context"
	"github.com/hostfactor/api/go/app"
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/diazo/pkg/variable"
)

func SetVariable(client app.AppServiceClient, store variable.Store, act *actions.SetVariable, entries ...*variable.Entry) error {
	name := variable.RenderString(act.GetName(), store, entries...)
	value := variable.RenderString(act.GetValue(), store, entries...)

	if act.GetSave() {
		_, err := client.SetVariable(context.Background(), &app.SetVariable_Request{
			Name:        name,
			Value:       value,
			DisplayName: variable.RenderString(act.GetDisplayName(), store, entries...),
		})
		if err != nil {
			return err
		}
	}

	store.AddEntries(variable.NewEntry(name, value))
	return nil
}
