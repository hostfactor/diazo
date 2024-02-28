package appcmd

import (
	"fmt"
	"github.com/hostfactor/api/go/blueprint/appcommand"
	"github.com/hostfactor/diazo/pkg/collection"
	"github.com/hostfactor/diazo/pkg/except"
	"slices"
	"strings"
)

func CompileExec(pl *appcommand.AppCommandPayload, cmds ...*appcommand.AppCommand) ([]byte, error) {
	cmd, err := FindCommand(pl.GetName(), cmds...)
	if err != nil {
		return nil, err
	}

	args := make([]string, 0, len(pl.GetArgs()))
	args = append(args, pl.GetName())

	return NewExec(pl.GetName(), pl.GetArgs(), cmd.GetSpec().GetOptions())
}

func NewExec(name string, args []*appcommand.AppCommandArg, opts []*appcommand.CommandOption) ([]byte, error) {
	output := make([]string, 0, len(args))
	output = append(output, name)
	activeArgs := collection.Index(opts, func(v *appcommand.CommandOption) string {
		return v.GetName()
	})
	for _, arg := range args {
		opt := activeArgs[arg.GetName()]
		if opt == nil {
			return nil, except.NewNotFound("%s is not a valid opt for %s command", arg.GetName(), name)
		}

		val, typ := GetVal(arg.GetValue())
		if opt.Type != typ {
			return nil, except.NewInvalid("expected %s for %s but got %s", strings.ToLower(opt.Type.String()), opt.GetName(), strings.ToLower(typ.String()))
		}
		if typ == appcommand.CommandOption_SUBCOMMAND {
			valArgs := val.([]*appcommand.AppCommandArg)
			cc, err := NewExec(arg.GetName(), valArgs, opt.GetSubcommand().GetOptions())
			if err != nil {
				return nil, err
			}
			output = append(output, string(cc))
		} else {
			output = append(output, fmt.Sprintf("%v", val))
		}
		delete(activeArgs, arg.GetName())
	}
	for _, v := range activeArgs {
		if v.GetRequired() {
			return nil, except.NewInvalid("arg %s is required", v.GetName())
		}
	}

	return []byte(strings.Join(output, " ")), nil
}

type Identifiable interface {
	GetName() string
	GetDescription() string
}

func FindCommand[T Identifiable](name string, cmds ...T) (out T, err error) {
	idx := slices.IndexFunc(cmds, func(c T) bool {
		return c.GetName() == name
	})
	if idx < 0 {
		err = except.NewNotFound("%s is not a valid command", name)
		return
	}

	return cmds[idx], nil
}

func GetVal(val *appcommand.CommandValue) (any, appcommand.CommandOption_Type) {
	if val == nil {
		return nil, appcommand.CommandOption_UNKNOWN
	}
	if val.BoolVal != nil {
		return val.GetBoolVal(), appcommand.CommandOption_BOOL
	} else if val.StrVal != nil {
		return val.GetStrVal(), appcommand.CommandOption_STRING
	} else if val.FloatVal != nil {
		return val.GetFloatVal(), appcommand.CommandOption_FLOAT
	} else if val.IntVal != nil {
		return val.GetIntVal(), appcommand.CommandOption_INT
	} else if val.ListVal != nil {
		return val.ListVal, appcommand.CommandOption_SUBCOMMAND
	}
	return nil, appcommand.CommandOption_UNKNOWN
}

func NewVal(a any) *appcommand.CommandValue {
	switch t := a.(type) {
	case string:
		return &appcommand.CommandValue{
			StrVal: &t,
		}
	case bool:
		return &appcommand.CommandValue{
			BoolVal: &t,
		}
	case float32:
		return &appcommand.CommandValue{
			FloatVal: &t,
		}
	case float64:
		v := float32(t)
		return &appcommand.CommandValue{
			FloatVal: &v,
		}
	case int:
		v := int32(t)
		return &appcommand.CommandValue{
			IntVal: &v,
		}
	case int32:
		return &appcommand.CommandValue{
			IntVal: &t,
		}
	case int16:
		v := int32(t)
		return &appcommand.CommandValue{
			IntVal: &v,
		}
	case int8:
		v := int32(t)
		return &appcommand.CommandValue{
			IntVal: &v,
		}
	case int64:
		v := int32(t)
		return &appcommand.CommandValue{
			IntVal: &v,
		}
	case uint:
		v := int32(t)
		return &appcommand.CommandValue{
			IntVal: &v,
		}
	case uint8:
		v := int32(t)
		return &appcommand.CommandValue{
			IntVal: &v,
		}
	case uint16:
		v := int32(t)
		return &appcommand.CommandValue{
			IntVal: &v,
		}
	case uint32:
		v := int32(t)
		return &appcommand.CommandValue{
			IntVal: &v,
		}
	case uint64:
		v := int32(t)
		return &appcommand.CommandValue{
			IntVal: &v,
		}
	case []*appcommand.AppCommandArg:
		return &appcommand.CommandValue{
			ListVal: t,
		}
	}
	return nil
}
