package providerconfig

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/hostfactor/api/go/blueprint/steps"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/collection"
	"github.com/hostfactor/diazo/pkg/ptr"
	"github.com/stretchr/testify/suite"
	"github.com/xeipuuv/gojsonschema"
	"google.golang.org/protobuf/testing/protocmp"
	"testing"
	"testing/fstest"
)

type ClientTestSuite struct {
	suite.Suite

	Target *client
}

func (c *ClientTestSuite) TestLoad() {
	// -- Given
	//
	expectedSettings := []byte(`{}`)
	testFs := fstest.MapFS{
		"provider.yaml": {
			Data: []byte(`
title: minecraft
image: itzg/minecraft-server
docs:
  entries:
    - title: Getting started
      path: ./docs/getting_started.md
forms:
  - steps:
    - id: settings
      components:
        - json_schema:
            path: settings.json
    screens:
      - start
  - steps:
    - id: select
      components:
        - file_select:
            accept_props:
              - .jar
            help_text: 'A *.zip file of your minecraft world data.'
            description: 'The save file that will be used by your server. We will automatically backup any new saves to [your save]_autosave.'
            destination:
              bucket_folder: saves
    screens:
      - start
volumes:
  - name: derp
    mount:
      path: /path/to/file
`),
		},
		"settings.json": {
			Data: expectedSettings,
		},
		"docs/getting_started.md": {
			Data: []byte(`
# Getting started			
			`),
		},
	}

	expectedSettingsSchema, _ := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(expectedSettings))

	expected := &LoadedProviderConfig{
		Config: &providerconfig.ProviderConfig{
			Title: "minecraft",
			Image: "itzg/minecraft-server",
			Docs: &providerconfig.Docs{
				Entries: []*providerconfig.Docs_Entry{
					{
						Title: "Getting started",
						Path:  "./docs/getting_started.md",
					},
				},
			},
			Volumes: []*providerconfig.Volume{
				{
					Name:  "derp",
					Mount: &providerconfig.VolumeMount{Path: "/path/to/file"},
				},
			},
			Forms: []*providerconfig.SettingsForm{
				{
					Screens: []providerconfig.Screen_Enum{
						providerconfig.Screen_start,
					},
					Steps: []*steps.Step{
						{
							Id: "settings",
							Components: []*steps.Component{
								{
									JsonSchema: &steps.JSONSchemaComponent{
										Path:   ptr.String("settings.json"),
										Schema: ptr.String(string(expectedSettings)),
									},
								},
							},
						},
					},
				},
				{
					Screens: []providerconfig.Screen_Enum{
						providerconfig.Screen_start,
					},
					Steps: []*steps.Step{
						{
							Id: "select",
							Components: []*steps.Component{
								{
									FileSelect: &steps.FileSelectComponent{
										AcceptProps: []string{".jar"},
										HelpText:    ptr.Ptr("A *.zip file of your minecraft world data."),
										Description: ptr.Ptr("The save file that will be used by your server. We will automatically backup any new saves to [your save]_autosave."),
									},
								},
							},
						},
					},
				},
			},
		},
		Filename: "provider.yaml",
	}

	expected.Forms = []Form{
		{
			raw: expected.Config.Forms[0],
			Steps: map[string]*CompiledStep{
				"settings": {
					step: expected.Config.Forms[0].Steps[0],
					components: []*CompiledComponent{
						{
							componentType: ComponentTypeJSONSchema,
							component:     expected.Config.Forms[0].Steps[0].Components[0],
							jsonSchema:    expectedSettingsSchema,
						},
					},
				},
			},
		},
		{
			raw: expected.Config.Forms[1],
			Steps: map[string]*CompiledStep{
				"select": {
					step: expected.Config.Forms[1].Steps[0],
					components: []*CompiledComponent{
						{
							componentType: ComponentTypeFileSelect,
							component:     expected.Config.Forms[1].Steps[0].Components[0],
						},
					},
				},
			},
		},
	}

	// -- When
	//
	actual, err := DefaultClient.Load(testFs, "provider.yaml")

	// -- Then
	//
	if c.NoError(err) {
		c.EqualProviderConfig(expected, actual)
	}
}

func (c *ClientTestSuite) TestLoadNoAppSettings() {
	// -- Given
	//
	expectedSettings := []byte(`{}`)
	testFs := fstest.MapFS{
		"provider.yaml": {
			Data: []byte(`
title: minecraft
image: itzg/minecraft-server
docs:
  entries:
    - title: Getting started
      path: ./docs/getting_started.md
volumes:
  - name: derp
    mount:
      path: /path/to/file
    source:
      file_input:
        accept_props:
          - .jar
        help_text: 'A *.zip file of your minecraft world data.'
        description: 'The save file that will be used by your server. We will automatically backup any new saves to [your save]_autosave.'
        destination:
          bucket_folder: saves
`),
		},
		"settings.json": {
			Data: expectedSettings,
		},
		"docs/getting_started.md": {
			Data: []byte(`
# Getting started			
			`),
		},
	}

	expectedSettingsSchema, _ := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(expectedSettings))

	expected := &LoadedProviderConfig{
		Config: &providerconfig.ProviderConfig{
			Title: "minecraft",
			Image: "itzg/minecraft-server",
			Docs: &providerconfig.Docs{
				Entries: []*providerconfig.Docs_Entry{
					{
						Title: "Getting started",
						Path:  "./docs/getting_started.md",
					},
				},
			},
			Volumes: []*providerconfig.Volume{
				{
					Name:  "derp",
					Mount: &providerconfig.VolumeMount{Path: "/path/to/file"},
					Source: &providerconfig.VolumeSource{
						FileInput: &providerconfig.FileInputSpec{
							AcceptProps: []string{".jar"},
							HelpText:    "A *.zip file of your minecraft world data.",
							Destination: &providerconfig.FileInputDestination{BucketFolder: "saves"},
							Description: "The save file that will be used by your server. We will automatically backup any new saves to [your save]_autosave.",
						},
					},
				},
			},
		},
		Filename: "provider.yaml",
		Forms:    []Form{},
	}

	expected.Settings = expectedSettingsSchema
	expected.RawSettings = string(expectedSettings)

	// -- When
	//
	actual, err := DefaultClient.Load(testFs, "provider.yaml")

	// -- Then
	//
	if c.NoError(err) {
		c.EqualProviderConfig(expected, actual)
	}
}

func (c *ClientTestSuite) TestLoadJson() {
	// -- Given
	//
	expectedSettings := []byte(`{}`)
	testFs := fstest.MapFS{
		"provider.json": {
			Data: []byte(`
{
  "title": "minecraft",
  "image": "itzg/minecraft-server",
  "volumes": [
    {
      "name": "derp",
      "mount": {
        "path": "/path/to/file"
      },
      "source": {
        "file_input": {
          "accept_props": [
            ".jar"
          ],
          "help_text": "A *.zip file of your minecraft world data.",
          "description": "The save file that will be used by your server. We will automatically backup any new saves to [your save]_autosave.",
          "destination": {
            "bucket_folder": "saves"
          }
        }
      }
    }
  ],
  "forms": [
			{
    "steps": [
      {
        "id": "settings",
        "components": [
          {
            "json_schema": {
              "path": "settings.json"
            }
          }
        ]
      }
    ]
			}
  ]
}
`),
		},
		"settings.json": {
			Data: expectedSettings,
		},
	}

	expectedSettingsSchema, _ := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(expectedSettings))

	expected := &LoadedProviderConfig{
		Config: &providerconfig.ProviderConfig{
			Title: "minecraft",
			Image: "itzg/minecraft-server",
			Volumes: []*providerconfig.Volume{
				{
					Name:  "derp",
					Mount: &providerconfig.VolumeMount{Path: "/path/to/file"},
					Source: &providerconfig.VolumeSource{
						FileInput: &providerconfig.FileInputSpec{
							AcceptProps: []string{".jar"},
							HelpText:    "A *.zip file of your minecraft world data.",
							Destination: &providerconfig.FileInputDestination{BucketFolder: "saves"},
							Description: "The save file that will be used by your server. We will automatically backup any new saves to [your save]_autosave.",
						},
					},
				},
			},
			Forms: []*providerconfig.SettingsForm{
				{
					Steps: []*steps.Step{
						{
							Id: "settings",
							Components: []*steps.Component{
								{
									JsonSchema: &steps.JSONSchemaComponent{
										Path:   ptr.String("settings.json"),
										Schema: ptr.String(string(expectedSettings)),
									},
								},
							},
						},
					},
				},
			},
		},
		Filename: "provider.json",
	}

	expected.Forms = []Form{
		{
			raw: expected.Config.Forms[0],
			Steps: map[string]*CompiledStep{
				"settings": {
					components: []*CompiledComponent{
						{
							component:  expected.Config.Forms[0].Steps[0].Components[0],
							jsonSchema: expectedSettingsSchema,
						},
					},
					step: expected.Config.Forms[0].Steps[0],
				},
			},
		},
	}

	// -- When
	//
	actual, err := DefaultClient.Load(testFs, "provider.json")

	// -- Then
	//
	if c.NoError(err) {
		c.EqualProviderConfig(expected, actual)
	}
}

func (c *ClientTestSuite) TestLoadMissingFile() {
	// -- Given
	//
	expectedSettings := []byte(`{}`)
	testFs := fstest.MapFS{
		"provider.yaml": {},
		"settings.json": {
			Data: expectedSettings,
		},
	}

	// -- When
	//
	actual, err := c.Target.Load(testFs, "provider.yml")

	// -- Then
	//
	c.Nil(actual)
	c.EqualError(err, "open provider.yml: file does not exist")
}

func (c *ClientTestSuite) TestLoadInvalidYaml() {
	// -- Given
	//
	expectedSettings := []byte(`{}`)
	testFs := fstest.MapFS{
		"provider.yaml": {
			Data: []byte(`invalid file.`),
		},
		"settings.json": {
			Data: expectedSettings,
		},
	}

	// -- When
	//
	actual, err := c.Target.Load(testFs, "provider.yaml")

	// -- Then
	//
	c.Nil(actual)
	c.EqualError(err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `invalid...` into map[string]interface {}")
}

func (c *ClientTestSuite) TestLoadInvalidJson() {
	// -- Given
	//
	testFs := fstest.MapFS{
		"provider.json": {
			Data: []byte(`invalid file.`),
		},
		"settings.json": {
			Data: []byte(`{}`),
		},
	}

	// -- When
	//
	actual, err := c.Target.Load(testFs, "provider.json")

	// -- Then
	//
	c.Nil(actual)
	c.EqualError(err, "invalid character 'i' looking for beginning of value")
}

func (c *ClientTestSuite) TestLoadMissingSettingsJson() {
	// -- Given
	//
	testFs := fstest.MapFS{
		"provider.json": {
			Data: []byte(`{
  "forms": [
			{
    "steps": [
      {
        "id": "settings",
        "components": [
          {
            "json_schema": {
              "path": "settings.json"
            }
          }
        ]
      }
    ]
			}
			]
  }
			}`),
		},
	}

	// -- When
	//
	actual, err := c.Target.Load(testFs, "provider.json")

	// -- Then
	//
	c.Nil(actual)
	c.EqualError(err, "problem loading step \"settings\": failed to load json schema: open settings.json: file does not exist")
}

func (c *ClientTestSuite) TestLoadInvalidSettingsJson() {
	// -- Given
	//
	testFs := fstest.MapFS{
		"provider.json": {
			Data: []byte(`
			{
  "forms": [
			{
    "steps": [
      {
        "id": "settings",
        "components": [
          {
            "json_schema": {
              "path": "settings.json"
            }
          }
        ]
      }
    ]
			}
  ]
			}
			`),
		},
		"settings.json": {
			Data: []byte(`invalid file.`),
		},
	}

	// -- When
	//
	actual, err := c.Target.Load(testFs, "provider.json")

	// -- Then
	//
	c.Nil(actual)
	c.EqualError(err, "problem loading step \"settings\": invalid json schema: invalid character 'i' looking for beginning of value")
}

func (c *ClientTestSuite) TestLoadInvalidFileExt() {
	// -- Given
	//
	testFs := fstest.MapFS{
		"provider.txt": {
			Data: []byte(`invalid file.`),
		},
	}

	// -- When
	//
	actual, err := c.Target.Load(testFs, "provider.txt")

	// -- Then
	//
	c.Nil(actual)
	c.EqualError(err, ".txt is not a supported extension for provider configs")
}

func (c *ClientTestSuite) TestLoadAll() {
	// -- Given
	//
	expectedSettings := []byte(`{}`)
	testFs := fstest.MapFS{
		"minecraft/provider.yaml": {
			Data: []byte(`
title: minecraft
forms:
  - steps:
      - id: settings
        components:
          - json_schema:
              path: settings.json
      `),
		},
		"minecraft/settings.json": {
			Data: expectedSettings,
		},
		"factorio/provider.yaml": {
			Data: []byte(`
title: factorio
forms:
  - steps:
    - id: settings
      components:
        - json_schema:
            path: settings.json
      `),
		},
		"factorio/settings.json": {
			Data: expectedSettings,
		},
		"valheim/provider.yaml": {
			Data: []byte(`
title: valheim
forms:
  - steps:
    - id: settings
      components:
        - json_schema:
            path: settings.json
      `),
		},
		"valheim/settings.json": {
			Data: expectedSettings,
		},
		"random_file.txt": {},
	}

	expectedComps := []*providerconfig.SettingsForm{
		{
			Steps: []*steps.Step{
				{
					Id: "settings",
					Components: []*steps.Component{
						{
							JsonSchema: &steps.JSONSchemaComponent{Path: ptr.String("settings.json")},
						},
					},
				},
			},
		},
	}

	f, _ := testFs.Sub("factorio")
	exp, _ := CompileStep(f, expectedComps[0].Steps[0])
	expForms := []Form{
		{
			Steps: map[string]*CompiledStep{"settings": exp},
			raw:   expectedComps[0],
		},
	}

	expected := []*LoadedProviderConfig{
		{
			Config: &providerconfig.ProviderConfig{
				Title: "factorio",
				Forms: expectedComps,
			},
			Filename: "factorio/provider.yaml",
			Forms:    expForms,
		},
		{
			Config: &providerconfig.ProviderConfig{
				Title: "minecraft",
				Forms: expectedComps,
			},
			Filename: "minecraft/provider.yaml",
			Forms:    expForms,
		},
		{
			Config: &providerconfig.ProviderConfig{
				Title: "valheim",
				Forms: expectedComps,
			},
			Filename: "valheim/provider.yaml",
			Forms:    expForms,
		},
	}

	// -- When
	//
	actual, err := c.Target.LoadAll(testFs)

	// -- Then
	//
	if c.NoError(err) {
		c.EqualProviderConfigs(expected, actual)
	}
}

func (c *ClientTestSuite) TestMatches() {
	// -- Given
	//
	given := []Form{
		{
			raw: &providerconfig.SettingsForm{
				Screens: []providerconfig.Screen_Enum{
					providerconfig.Screen_start,
				},
				Steps: []*steps.Step{
					{
						Id:    "step1",
						Title: ptr.Ptr("step 1"),
					},
				},
			},
		},
		{
			raw: &providerconfig.SettingsForm{
				Screens: []providerconfig.Screen_Enum{
					providerconfig.Screen_update,
				},
				Steps: []*steps.Step{
					{
						Id:    "step2",
						Title: ptr.Ptr("step 2"),
					},
				},
			},
		},
		{
			raw: &providerconfig.SettingsForm{
				Screens: []providerconfig.Screen_Enum{
					providerconfig.Screen_start,
				},
				Steps: []*steps.Step{
					{
						Id:    "step3",
						Title: ptr.Ptr("step 3"),
					},
				},
			},
		},
	}

	query := FormQuery{Screen: providerconfig.Screen_start}

	// -- When
	//
	actual := collection.Filter(given, func(t Form) bool {
		return query.Matches(&t)
	})

	// -- Then
	//
	if c.Len(actual, 2) {
		c.Equal(actual[0].raw.Steps[0].Id, "step1")
		c.Equal(actual[1].raw.Steps[0].Id, "step3")
	}
}

type EqualProviderConfigsOpt func(o equalProviderConfigOpts) *equalProviderConfigOpts

type equalProviderConfigOpts struct {
}

func (c *ClientTestSuite) EqualProviderConfig(expected, actual *LoadedProviderConfig, o ...EqualProviderConfigsOpt) bool {
	opts := &equalProviderConfigOpts{}
	for _, v := range o {
		opts = v(*opts)
	}

	c.Len(actual.Forms, len(expected.Forms))

	for i, form := range expected.Forms {
		c.Len(actual.Forms[i].Steps, len(form.Steps))

		c.Empty(cmp.Diff(actual.Forms[i].raw, form.raw, protocmp.Transform()))

		for k, v := range form.Steps {
			a, ok := actual.Forms[i].Steps[k]
			if !c.True(ok) {
				continue
			}

			if c.Len(a.components, len(v.components)) {
				for i, expectedComp := range v.components {
					actualComp := a.components[i]

					if expectedComp.Component() != nil {
						c.Empty(cmp.Diff(expectedComp.Component(), actualComp.component, protocmp.Transform()))
					}

					if expectedComp.jsonSchema != nil {
						c.NotNil(actualComp.jsonSchema)
					} else {
						c.Nil(actualComp.jsonSchema)
					}

					c.Equal(expectedComp.component.GetJsonSchema().GetSchema(), actualComp.component.GetJsonSchema().GetSchema())
				}
			}

			c.Equal(v.validation, a.validation)

			if v.Step() != nil {
				c.Empty(cmp.Diff(v.Step(), a.Step(), protocmp.Transform()))
			}
		}
	}

	c.Empty(cmp.Diff(expected.Config, actual.Config, protocmp.Transform()))
	return c.Equal(expected.Filename, actual.Filename)
}

func (c *ClientTestSuite) EqualProviderConfigs(expected, actual []*LoadedProviderConfig, o ...EqualProviderConfigsOpt) bool {
	if !c.Equal(len(expected), len(actual)) {
		return false
	}

	for idx, v := range expected {
		if !c.EqualProviderConfig(v, actual[idx], o...) {
			fmt.Printf("element %d is not equal.", idx)
			return false
		}
	}
	return true
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
