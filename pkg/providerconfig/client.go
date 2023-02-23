package providerconfig

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/doccache"
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
)

type LoadedProviderConfig struct {
	Config *providerconfig.ProviderConfig
	// schemas by step IDs
	SettingsSchema map[string]*CompiledStep
	DocCache       doccache.DocCache
	Filename       string
}

type Validator interface {
	Validate(val *blueprint.Value) error
}

type CompiledStep struct {
	Step          *providerconfig.Step
	JSONSchema    *gojsonschema.Schema
	RawJSONSchema string
	Validation    *CompiledStepValidation
}

func (c *CompiledStep) Validate(val *blueprint.Value) error {
	err := c.Validation.Validate(val)
	if err != nil {
		return err
	}

	if val.GetStringValue() != "" {
		_, err = c.JSONSchema.Validate(gojsonschema.NewStringLoader(val.GetStringValue()))
	}

	return err
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
		SettingsSchema: map[string]*CompiledStep{},
	}

	out.Config, err = c.LoadProviderFile(f, providerFilename)
	if err != nil {
		return nil, err
	}

	out.DocCache, err = doccache.New(f, out.Config)
	if err != nil {
		return nil, err
	}

	for i := range out.Config.GetAppSettings().GetSteps() {
		step := out.Config.GetAppSettings().GetSteps()[i]
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

		if step.GetJsonSchema().GetPath() != "" {
			rawSettings, err := fs.ReadFile(f, step.GetJsonSchema().GetPath())
			if err != nil {
				return nil, fmt.Errorf("failed to load json schema for step \"%s\": %w", step.GetId(), err)
			}

			loader := gojsonschema.NewBytesLoader(rawSettings)
			compiled.JSONSchema, err = gojsonschema.NewSchema(loader)
			if err != nil {
				return nil, fmt.Errorf("invalid json schema for step \"%s\": %w", step.GetId(), err)
			}

			compiled.RawJSONSchema = string(rawSettings)
		}

		out.SettingsSchema[step.GetId()] = compiled
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
