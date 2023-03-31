package providerconfig

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/blueprint/steps"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/doccache"
	"github.com/hostfactor/diazo/pkg/ptr"
	jsoniter "github.com/json-iterator/go"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
	"io/fs"
	"path/filepath"
	"regexp"
)

var DefaultClient = NewClient()

var (
	DefaultProviderFilename = "provider.yaml"
	DefaultSettingsFilename = "settings.json"
)

type LoadedProviderConfig struct {
	Config      *providerconfig.ProviderConfig
	DocCache    doccache.DocCache
	Filename    string
	Settings    *gojsonschema.Schema
	RawSettings string

	// schemas by step IDs
	settingsSchema map[string]*CompiledStep
}

func (l *LoadedProviderConfig) Validate(val *blueprint.AppSettings) *blueprint.Validation {
	compVals := val.GetComponentValues()
	if compVals == nil {
		compVals = map[string]*blueprint.ValueSet{}
	}

	for k, v := range l.settingsSchema {
		valSet, ok := compVals[k]
		if !ok {
			valid := &blueprint.Validation{
				Problems: []*blueprint.ValidationProblem{
					{
						What:  fmt.Sprintf("Setting %s required.", k),
						Where: k,
					},
				},
			}

			return valid
		}

		for _, val := range valSet.GetValues() {
			err := v.Validate(val)
			if err != nil {
				valid := &blueprint.Validation{
					Problems: []*blueprint.ValidationProblem{
						{
							What:  err.Error(),
							Where: k,
						},
					},
				}

				return valid
			}
		}
	}

	return nil
}

type Validator[T any] interface {
	Validate(val T) error
}

type CompiledStep struct {
	Step       *steps.Step
	Components []*CompiledComponent
	Validation *CompiledStepValidation
}

func (c *CompiledStep) Validate(val *blueprint.Value) error {
	err := c.Validation.Validate(val)
	if err != nil {
		return err
	}

	for _, comp := range c.Components {
		err = comp.Validate(val)
		if err != nil {
			return err
		}
	}

	return err
}

type CompiledComponent struct {
	Component    *steps.Component
	JSONSchema   *gojsonschema.Schema
	VersionRegex *regexp.Regexp
}

func (c *CompiledComponent) Validate(val *blueprint.Value) error {
	if val.GetStringValue() == "" {
		return nil
	}

	if c.JSONSchema != nil {
		_, err := c.JSONSchema.Validate(gojsonschema.NewStringLoader(val.GetStringValue()))
		return err
	}

	return nil
}

type CompiledStepValidation struct {
	Regex   *regexp.Regexp
	Message string
}

func (c *CompiledStepValidation) Validate(val *blueprint.Value) error {
	if val.GetStringValue() != "" {
		if !c.Regex.MatchString(val.GetStringValue()) {
			return errors.New(c.Message)
		}
	}

	return nil
}

type Client interface {
	// Load loads a single LoadedProviderConfig from the specified directory.
	Load(f fs.FS, providerFilename string) (*LoadedProviderConfig, error)

	// LoadAll loads all provider configs within the directory. Assumes that every Provider directory is housed as a child
	// directory of the specified fs.FS.
	LoadAll(f fs.FS) ([]*LoadedProviderConfig, error)

	// LoadProviderFile loads the provider file from the fs.FS and the relative path of fp and unmarshalls it.
	LoadProviderFile(f fs.FS, fp string) (*providerconfig.ProviderConfig, error)
}

func NewClient() Client {
	return &client{}
}

func Load(f fs.FS) (*LoadedProviderConfig, error) {
	return DefaultClient.Load(f, DefaultProviderFilename)
}

type client struct {
}

func (c *client) LoadAll(f fs.FS) ([]*LoadedProviderConfig, error) {
	entries, err := fs.ReadDir(f, ".")
	if err != nil {
		return nil, err
	}

	confs := make([]*LoadedProviderConfig, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sub, err := fs.Sub(f, entry.Name())
		if err != nil {
			return nil, err
		}

		conf, err := c.Load(sub, DefaultProviderFilename)
		if err != nil {
			return nil, err
		}
		conf.Filename = filepath.Join(entry.Name(), conf.Filename)

		confs = append(confs, conf)
	}

	return confs, nil
}

func (c *client) Load(f fs.FS, providerFilename string) (*LoadedProviderConfig, error) {
	var err error
	out := &LoadedProviderConfig{
		Filename:       providerFilename,
		settingsSchema: map[string]*CompiledStep{},
	}

	out.Config, err = c.LoadProviderFile(f, providerFilename)
	if err != nil {
		return nil, err
	}

	// TODO: remove block when fully deprecated default settings.json
	if len(out.Config.GetAppSettings().GetSteps()) == 0 {
		rawSettings, err := fs.ReadFile(f, DefaultSettingsFilename)
		if err != nil {
			return nil, fmt.Errorf("failed to load json schema at %s: %w", DefaultSettingsFilename, err)
		}

		out.RawSettings = string(rawSettings)

		loader := gojsonschema.NewBytesLoader(rawSettings)
		out.Settings, err = gojsonschema.NewSchema(loader)
		if err != nil {
			return nil, fmt.Errorf("invalid json schema %s: %w", DefaultSettingsFilename, err)
		}
	} else {
		for i := range out.Config.GetAppSettings().GetSteps() {
			step := out.Config.AppSettings.Steps[i]
			compiled := &CompiledStep{
				Step: step,
			}

			if step.GetValidation().GetRegex() != "" {
				reg, err := regexp.Compile(step.GetValidation().GetRegex())
				if err != nil {
					return nil, fmt.Errorf("invalid regex validation for step \"%s\": %w", step.GetId(), err)
				}

				compiled.Validation = &CompiledStepValidation{
					Regex:   reg,
					Message: step.GetValidation().GetMessage(),
				}
			}

			for _, comp := range step.GetComponents() {
				co := &CompiledComponent{
					Component: comp,
				}

				if comp.GetJsonSchema().GetPath() != "" {
					rawSettings, err := fs.ReadFile(f, comp.GetJsonSchema().GetPath())
					if err != nil {
						return nil, fmt.Errorf("failed to load json schema for step \"%s\": %w", step.GetId(), err)
					}

					loader := gojsonschema.NewBytesLoader(rawSettings)
					co.JSONSchema, err = gojsonschema.NewSchema(loader)
					if err != nil {
						return nil, fmt.Errorf("invalid json schema for step \"%s\": %w", step.GetId(), err)
					}

					co.Component.JsonSchema.Schema = ptr.String(string(rawSettings))
				}

				compiled.Components = append(compiled.Components, co)
			}

			out.settingsSchema[step.GetId()] = compiled
		}
	}

	out.DocCache, err = doccache.New(f, out.Config)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *client) LoadProviderFile(f fs.FS, fp string) (*providerconfig.ProviderConfig, error) {
	_, filename := filepath.Split(fp)
	ext := filepath.Ext(filename)
	content, err := fs.ReadFile(f, fp)
	if err != nil {
		return nil, err
	}

	if ext == ".yaml" || ext == ".yml" {
		m := map[string]interface{}{}
		err := yaml.Unmarshal(content, &m)
		if err != nil {
			return nil, err
		}

		content, err = jsoniter.Marshal(m)
		if err != nil {
			return nil, err
		}
	} else if ext != ".json" {
		return nil, fmt.Errorf("%s is not a supported extension for provider configs", ext)

	}

	conf := new(providerconfig.ProviderConfig)

	err = (&jsonpb.Unmarshaler{AllowUnknownFields: true}).Unmarshal(bytes.NewReader(content), conf)
	if err != nil {
		return nil, err
	}

	return conf, err
}
