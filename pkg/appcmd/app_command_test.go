package appcmd

import (
	"github.com/hostfactor/api/go/blueprint/appcommand"
	"github.com/hostfactor/diazo/pkg/ptr"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AppCommandTestSuite struct {
	suite.Suite
}

func (a *AppCommandTestSuite) TestCompileCommand() {
	type test struct {
		Description string
		Given       *appcommand.AppCommandPayload
		Cmds        []*appcommand.AppCommand
		Expected    string
		ExpectedErr string
	}

	tests := []test{
		{
			Description: "no args",
			Given: &appcommand.AppCommandPayload{
				Name: "save",
			},
			Cmds: []*appcommand.AppCommand{
				{Name: "save"},
			},
			Expected: "save",
		},
		{
			Description: "all types of arg",
			Given: &appcommand.AppCommandPayload{
				Name: "save",
				Args: []*appcommand.AppCommandArg{
					{Name: "int", Value: NewVal(1)},
					{Name: "str", Value: NewVal("s")},
					{Name: "bool", Value: NewVal(true)},
					{Name: "float", Value: NewVal(1.1)},
				},
			},
			Cmds: []*appcommand.AppCommand{
				{Name: "save", Spec: &appcommand.AppCommandSpec{Options: []*appcommand.CommandOption{
					{Name: "int", Type: appcommand.CommandOption_INT},
					{Name: "str", Type: appcommand.CommandOption_STRING},
					{Name: "bool", Type: appcommand.CommandOption_BOOL},
					{Name: "float", Type: appcommand.CommandOption_FLOAT},
				}}},
			},
			Expected: "save 1 s true 1.1",
		},
		{
			Description: "invalid arg value",
			Given: &appcommand.AppCommandPayload{
				Name: "save",
				Args: []*appcommand.AppCommandArg{
					{Name: "arg1", Value: NewVal("1")},
				},
			},
			Cmds: []*appcommand.AppCommand{
				{Name: "save", Spec: &appcommand.AppCommandSpec{Options: []*appcommand.CommandOption{
					{Name: "arg1", Type: appcommand.CommandOption_INT},
				}}},
			},
			ExpectedErr: "expected int for arg1 but got string: invalid",
		},
		{
			Description: "missing required arg",
			Given: &appcommand.AppCommandPayload{
				Name: "save",
			},
			Cmds: []*appcommand.AppCommand{
				{Name: "save", Spec: &appcommand.AppCommandSpec{Options: []*appcommand.CommandOption{
					{Name: "arg1", Type: appcommand.CommandOption_INT, Required: ptr.Ptr(true)},
				}}},
			},
			ExpectedErr: "arg arg1 is required: invalid",
		},
		{
			Description: "missing optional arg",
			Given: &appcommand.AppCommandPayload{
				Name: "save",
			},
			Cmds: []*appcommand.AppCommand{
				{Name: "save", Spec: &appcommand.AppCommandSpec{Options: []*appcommand.CommandOption{
					{Name: "int", Type: appcommand.CommandOption_INT},
				}}},
			},
			Expected: "save",
		},
		{
			Description: "invalid command",
			Given: &appcommand.AppCommandPayload{
				Name: "save1",
			},
			Cmds: []*appcommand.AppCommand{
				{Name: "save"},
			},
			ExpectedErr: "save1 is not a valid command: not found",
		},
		{
			Description: "invalid arg",
			Given: &appcommand.AppCommandPayload{
				Name: "save",
				Args: []*appcommand.AppCommandArg{
					{Name: "arg2", Value: NewVal(1)},
				},
			},
			Cmds: []*appcommand.AppCommand{
				{Name: "save", Spec: &appcommand.AppCommandSpec{Options: []*appcommand.CommandOption{
					{Name: "arg1", Type: appcommand.CommandOption_INT},
				}}},
			},
			ExpectedErr: "arg2 is not a valid opt for save command: not found",
		},
		{
			Description: "subcommand",
			Given: &appcommand.AppCommandPayload{
				Name: "user",
				Args: []*appcommand.AppCommandArg{
					{
						Name: "add",
						Value: NewVal([]*appcommand.AppCommandArg{
							{
								Name:  "user name",
								Value: NewVal("hi"),
							},
						}),
					},
				},
			},
			Cmds: []*appcommand.AppCommand{
				{Name: "user", Spec: &appcommand.AppCommandSpec{Options: []*appcommand.CommandOption{
					{Name: "add", Type: appcommand.CommandOption_SUBCOMMAND, Spec: &appcommand.AppCommandSpec{
						Options: []*appcommand.CommandOption{
							{Name: "user name", Type: appcommand.CommandOption_STRING, Required: ptr.Ptr(true)},
						},
					}},
				}}},
			},
			Expected: "user add hi",
		},
		{
			Description: "missing subcommand",
			Given: &appcommand.AppCommandPayload{
				Name: "user",
				Args: []*appcommand.AppCommandArg{
					{
						Name: "ad",
						Value: NewVal([]*appcommand.AppCommandArg{
							{
								Name:  "user name",
								Value: NewVal("hi"),
							},
						}),
					},
				},
			},
			Cmds: []*appcommand.AppCommand{
				{Name: "user", Spec: &appcommand.AppCommandSpec{Options: []*appcommand.CommandOption{
					{Name: "add", Type: appcommand.CommandOption_SUBCOMMAND, Spec: &appcommand.AppCommandSpec{
						Options: []*appcommand.CommandOption{
							{Name: "user name", Type: appcommand.CommandOption_STRING, Required: ptr.Ptr(true)},
						},
					}},
				}}},
			},
			ExpectedErr: "ad is not a valid opt for user command: not found",
		},
	}

	for _, v := range tests {
		cmd, err := CompileExec(v.Given, v.Cmds...)
		if err != nil {
			a.EqualError(err, v.ExpectedErr, v.Description)
			continue
		}
		a.Equal(v.Expected, string(cmd), v.Description)
	}
}

func TestAppCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AppCommandTestSuite))
}
