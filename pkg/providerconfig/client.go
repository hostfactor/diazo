package providerconfig

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/blueprint/steps"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/collection"
	"github.com/hostfactor/diazo/pkg/doccache"
	"github.com/hostfactor/diazo/pkg/fscache"
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
)

type LoadedProviderConfig struct {
	Config         *providerconfig.ProviderConfig
	DocCache       doccache.DocCache
	Filename       string
	Forms          []Form
	TagRegexFilter *regexp.Regexp
	// The directory that houses the provider manifest.
	Root fs.FS
}

func (l *LoadedProviderConfig) Validate(val *blueprint.BlueprintData) *blueprint.Validation {
	compVals := val.GetAppSettings().GetComponentValues()
	if compVals == nil {
		compVals = map[string]*blueprint.ValueSet{}
	}

	for _, form := range l.Forms {
		for k, v := range form.Steps {
			valSet, ok := compVals[k]
			if !ok {
				continue
			}

			err := v.Validate(valSet)
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

type Form struct {
	// Steps by ID
	Steps map[string]*CompiledStep

	raw *providerconfig.SettingsForm
}

type CompiledStep struct {
	step       *steps.Step
	components []*CompiledComponent
	validation *CompiledStepValidation
}

func (c *CompiledStep) Components() []*CompiledComponent {
	if c == nil {
		return nil
	}
	return c.components
}

func (c *CompiledStep) Step() *steps.Step {
	return &steps.Step{
		Id: c.step.Id,
		Components: collection.Map(c.components, func(f *CompiledComponent) *steps.Component {
			return f.Component()
		}),
		Title:   c.step.Title,
		Dynamic: c.step.Dynamic,
		Skip:    c.step.Skip,
	}
}

func (c *CompiledStep) Validate(val *blueprint.ValueSet) error {
	for _, comp := range c.components {
		err := comp.Validate(val.GetValues()[comp.Component().GetId()])
		if err != nil {
			return err
		}
	}

	return nil
}

type ComponentType int

const (
	ComponentTypeUnknown ComponentType = iota
	ComponentTypeJSONSchema
	ComponentTypeFileSelect
	ComponentTypeText
	ComponentTypeSelectButton
	ComponentTypeVersion
)

func CompileComponent(f fs.FS, comp *steps.Component) (*CompiledComponent, error) {
	co := &CompiledComponent{
		component: comp,
	}

	if comp.GetId() == "" {
		return nil, fmt.Errorf("id is required")
	}

	if comp.JsonSchema != nil {
		co.componentType = ComponentTypeJSONSchema

		if comp.JsonSchema.GetPath() != "" && comp.JsonSchema.GetSchema() == "" {
			rawSettings, err := fs.ReadFile(f, comp.GetJsonSchema().GetPath())
			if err != nil {
				return nil, fmt.Errorf("failed to load json schema: %w", err)
			}

			loader := gojsonschema.NewBytesLoader(rawSettings)
			co.jsonSchema, err = gojsonschema.NewSchema(loader)
			if err != nil {
				return nil, fmt.Errorf("invalid json schema: %w", err)
			}

			co.component.JsonSchema.Schema = ptr.String(string(rawSettings))
		}
	} else if comp.FileSelect != nil {
		co.componentType = ComponentTypeFileSelect
	} else if comp.Text != nil {
		co.componentType = ComponentTypeText
	} else if comp.Version != nil {
		co.componentType = ComponentTypeVersion
	} else if len(comp.GetSelectButton().GetOptions()) > 0 {
		co.componentType = ComponentTypeSelectButton
	} else {
		co.componentType = ComponentTypeUnknown
	}

	return co, nil
}

type CompiledComponent struct {
	component     *steps.Component
	jsonSchema    *gojsonschema.Schema
	versionRegex  *regexp.Regexp
	componentType ComponentType
}

func (c *CompiledComponent) DefaultValue() *blueprint.Value {
	switch c.componentType {
	case ComponentTypeSelectButton:
		opt := collection.Find(c.component.GetSelectButton().GetOptions(), func(t *steps.Button) bool {
			return t.GetDefault()
		})
		if opt != nil {
			return &blueprint.Value{StringValue: opt.Value}
		}
	}

	return nil
}

func (c *CompiledComponent) Type() ComponentType {
	return c.componentType
}

func (c *CompiledComponent) Component() *steps.Component {
	return c.component
}

func (c *CompiledComponent) Validate(val *blueprint.Value) error {
	if val.GetStringValue() == "" {
		return nil
	}

	if c.jsonSchema != nil {
		_, err := c.jsonSchema.Validate(gojsonschema.NewStringLoader(val.GetStringValue()))
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
	dir, _ := filepath.Split(providerFilename)
	root := f
	var err error
	if dir != "" {
		root, err = fs.Sub(f, dir)
		if err != nil {
			return nil, err
		}
	}
	root = fscache.New(root)

	raw, err := c.LoadProviderFile(root, providerFilename)
	if err != nil {
		return nil, err
	}

	comp, err := CompileProviderConfig(root, raw)
	if err != nil {
		return nil, err
	}

	comp.Root = root
	comp.Filename = providerFilename

	return comp, nil
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

func CompileProviderConfig(f fs.FS, conf *providerconfig.ProviderConfig) (*LoadedProviderConfig, error) {
	out := &LoadedProviderConfig{
		Config: conf,
	}
	var err error

	if len(out.Config.GetForms()) == 0 {
		out.Config.Forms = defaultForms(out.Config)
	}

	out.Forms, err = CompileForms(f, out.Config.GetForms()...)
	if err != nil {
		return nil, err
	}

	out.DocCache, err = doccache.New(f, out.Config)
	if err != nil {
		return nil, err
	}

	if conf.TagRegexFilter != nil {
		out.TagRegexFilter, err = regexp.Compile(conf.GetTagRegexFilter())
		if err != nil {
			return nil, fmt.Errorf("%s is not valid regex: %w", conf.GetTagRegexFilter(), err)
		}
	}

	return out, nil
}

func CompileStep(f fs.FS, step *steps.Step) (*CompiledStep, error) {
	compiled := &CompiledStep{
		step: step,
	}

	if step.GetId() == "" {
		return nil, fmt.Errorf("step ID is required")
	}

	if len(step.GetComponents()) == 0 {
		return nil, fmt.Errorf("step \"%s\" must have at least one component", step.GetId())
	}

	for _, comp := range step.GetComponents() {
		co, err := CompileComponent(f, comp)
		if err != nil {
			return nil, fmt.Errorf("problem loading step \"%s\": %w", step.GetId(), err)
		}

		compiled.components = append(compiled.components, co)
	}

	return compiled, nil
}

func CompileForm(f fs.FS, form *providerconfig.SettingsForm) (*Form, error) {
	frm := &Form{
		Steps: map[string]*CompiledStep{},
		raw:   form,
	}

	for i := range form.GetSteps() {
		step := form.Steps[i]
		compiled, err := CompileStep(f, step)
		if err != nil {
			return nil, err
		}

		frm.Steps[step.GetId()] = compiled
	}

	return frm, nil
}

func CompileForms(f fs.FS, forms ...*providerconfig.SettingsForm) ([]Form, error) {
	out := make([]Form, 0, len(forms))
	for _, form := range forms {
		frm, err := CompileForm(f, form)
		if err != nil {
			return nil, err
		}
		out = append(out, *frm)
	}
	return out, nil
}
