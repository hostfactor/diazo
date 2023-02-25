package providerconfig

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/hostfactor/api/go/blueprint/steps"
	"github.com/hostfactor/api/go/providerconfig"
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

func (p *ClientTestSuite) TestLoad() {
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
app_settings:
  steps:
    - id: settings
      components:
        - json_schema:
            path: settings.json
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
			AppSettings: &providerconfig.AppSettingsSchema{
				Steps: []*steps.Step{
					{
						Id: "settings",
						Components: []*steps.Component{
							{
								JsonSchema: &steps.JSONSchemaComponent{
									Path: "settings.json",
								},
							},
						},
					},
				},
			},
		},
		Filename: "provider.yaml",
	}

	expected.SettingsSchema = map[string]*CompiledStep{
		"settings": {
			Step: expected.Config.AppSettings.Steps[0],
			Components: []*CompiledComponent{
				{
					RawJSONSchema: string(expectedSettings),
					JSONSchema:    expectedSettingsSchema,
				},
			},
		},
	}

	// -- When
	//
	actual, err := DefaultClient.Load(testFs, "provider.yaml")

	// -- Then
	//
	if p.NoError(err) {
		p.EqualProviderConfig(expected, actual)
	}
}

func (p *ClientTestSuite) TestLoadJson() {
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
  "app_settings": {
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
			AppSettings: &providerconfig.AppSettingsSchema{
				Steps: []*steps.Step{
					{
						Id: "settings",
						Components: []*steps.Component{
							{
								JsonSchema: &steps.JSONSchemaComponent{
									Path: "settings.json",
								},
							},
						},
					},
				},
			},
		},
		Filename: "provider.json",
	}

	expected.SettingsSchema = map[string]*CompiledStep{
		"settings": {
			Components: []*CompiledComponent{
				{
					JSONSchema:    expectedSettingsSchema,
					RawJSONSchema: string(expectedSettings),
				},
			},
			Step: expected.Config.AppSettings.Steps[0],
		},
	}

	// -- When
	//
	actual, err := DefaultClient.Load(testFs, "provider.json")

	// -- Then
	//
	if p.NoError(err) {
		p.EqualProviderConfig(expected, actual)
	}
}

func (p *ClientTestSuite) TestLoadMissingFile() {
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
	actual, err := p.Target.Load(testFs, "provider.yml")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, "open provider.yml: file does not exist")
}

func (p *ClientTestSuite) TestLoadInvalidYaml() {
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
	actual, err := p.Target.Load(testFs, "provider.yaml")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `invalid...` into map[string]interface {}")
}

func (p *ClientTestSuite) TestLoadInvalidJson() {
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
	actual, err := p.Target.Load(testFs, "provider.json")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, "invalid character 'i' looking for beginning of value")
}

func (p *ClientTestSuite) TestLoadMissingSettingsJson() {
	// -- Given
	//
	testFs := fstest.MapFS{
		"provider.json": {
			Data: []byte(`{
  "app_settings": {
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
			}`),
		},
	}

	// -- When
	//
	actual, err := p.Target.Load(testFs, "provider.json")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, "failed to load json schema for step \"settings\": open settings.json: file does not exist")
}

func (p *ClientTestSuite) TestLoadInvalidSettingsJson() {
	// -- Given
	//
	testFs := fstest.MapFS{
		"provider.json": {
			Data: []byte(`
			{
  "app_settings": {
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
			}
			`),
		},
		"settings.json": {
			Data: []byte(`invalid file.`),
		},
	}

	// -- When
	//
	actual, err := p.Target.Load(testFs, "provider.json")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, "invalid json schema for step \"settings\": invalid character 'i' looking for beginning of value")
}

func (p *ClientTestSuite) TestLoadInvalidFileExt() {
	// -- Given
	//
	testFs := fstest.MapFS{
		"provider.txt": {
			Data: []byte(`invalid file.`),
		},
	}

	// -- When
	//
	actual, err := p.Target.Load(testFs, "provider.txt")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, ".txt is not a supported extension for provider configs")
}

func (p *ClientTestSuite) TestLoadAll() {
	// -- Given
	//
	expectedSettings := []byte(`{}`)
	testFs := fstest.MapFS{
		"minecraft/provider.yaml": {
			Data: []byte(`title: minecraft`),
		},
		"minecraft/settings.json": {
			Data: expectedSettings,
		},
		"factorio/provider.yaml": {
			Data: []byte(`title: factorio`),
		},
		"factorio/settings.json": {
			Data: expectedSettings,
		},
		"valheim/provider.yaml": {
			Data: []byte(`title: valheim`),
		},
		"valheim/settings.json": {
			Data: expectedSettings,
		},
		"random_file.txt": {},
	}

	expected := []*LoadedProviderConfig{
		{
			Config:         &providerconfig.ProviderConfig{Title: "factorio"},
			Filename:       "factorio/provider.yaml",
			SettingsSchema: map[string]*CompiledStep{},
		},
		{
			Config:         &providerconfig.ProviderConfig{Title: "minecraft"},
			Filename:       "minecraft/provider.yaml",
			SettingsSchema: map[string]*CompiledStep{},
		},
		{
			Config:         &providerconfig.ProviderConfig{Title: "valheim"},
			Filename:       "valheim/provider.yaml",
			SettingsSchema: map[string]*CompiledStep{},
		},
	}

	// -- When
	//
	actual, err := p.Target.LoadAll(testFs)

	// -- Then
	//
	if p.NoError(err) {
		p.EqualProviderConfigs(expected, actual)
	}
}

type EqualProviderConfigsOpt func(o equalProviderConfigOpts) *equalProviderConfigOpts

type equalProviderConfigOpts struct {
}

func (p *ClientTestSuite) EqualProviderConfig(expected, actual *LoadedProviderConfig, o ...EqualProviderConfigsOpt) bool {
	opts := &equalProviderConfigOpts{}
	for _, v := range o {
		opts = v(*opts)
	}

	p.Len(actual.SettingsSchema, len(expected.SettingsSchema))

	for k, v := range expected.SettingsSchema {
		a, ok := actual.SettingsSchema[k]
		if !p.True(ok) {
			continue
		}

		if p.Len(a.Components, len(v.Components)) {
			for i, expectedComp := range v.Components {
				actualComp := a.Components[i]
				if expectedComp.JSONSchema != nil {
					p.NotNil(actualComp.JSONSchema)
				} else {
					p.Nil(actualComp.JSONSchema)
				}

				p.Equal(expectedComp.RawJSONSchema, actualComp.RawJSONSchema)
			}
		}

		p.Equal(v.Validation, a.Validation)

		p.Empty(cmp.Diff(v.Step, a.Step, protocmp.Transform()))
	}

	p.Empty(cmp.Diff(expected.Config, actual.Config, protocmp.Transform()))
	return p.Equal(expected.Filename, actual.Filename)
}

func (p *ClientTestSuite) EqualProviderConfigs(expected, actual []*LoadedProviderConfig, o ...EqualProviderConfigsOpt) bool {
	if !p.Equal(len(expected), len(actual)) {
		return false
	}

	for idx, v := range expected {
		if !p.EqualProviderConfig(v, actual[idx], o...) {
			fmt.Printf("element %d is not equal.", idx)
			return false
		}
	}
	return true
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
