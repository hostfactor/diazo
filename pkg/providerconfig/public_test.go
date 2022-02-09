package providerconfig

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/stretchr/testify/suite"
	"github.com/xeipuuv/gojsonschema"
	"google.golang.org/protobuf/testing/protocmp"
	"testing"
	"testing/fstest"
)

type PublicTestSuite struct {
	suite.Suite

	Target *client
}

func (p *PublicTestSuite) TestLoad() {
	// -- Given
	//
	expectedSettings := []byte(`{}`)
	testFs := fstest.MapFS{
		"provider.yaml": {
			Data: []byte(`
title: minecraft
version_sync:
  source:
    container_registry: 
      matches_tag: '^(latest|java8|java9|java11|java15)$'
image: itzg/minecraft-server
file_inputs:
  - accept_props:
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
	}

	expected := &ProviderConfig{
		Config: &providerconfig.ProviderConfig{
			Title: "minecraft",
			VersionSync: &providerconfig.VersionSyncSpec{
				Source: &providerconfig.RemoteVersionSource{
					ContainerRegistry: &providerconfig.ContainerRegistryVersionSource{
						MatchesTag: "^(latest|java8|java9|java11|java15)$",
					},
				},
			},
			Image: "itzg/minecraft-server",
			FileInputs: []*providerconfig.FileInputParam{
				{
					AcceptProps: []string{".jar"},
					HelpText:    "A *.zip file of your minecraft world data.",
					Destination: &providerconfig.FileInputParamDestination{BucketFolder: "saves"},
					Description: "The save file that will be used by your server. We will automatically backup any new saves to [your save]_autosave.",
				},
			},
		},
		Filename:       "provider.yaml",
		SettingsSchema: &gojsonschema.Schema{},
		RawSettings:    expectedSettings,
	}

	// -- When
	//
	actual, err := DefaultClient.Load(testFs, "provider.yaml", "settings.json")

	// -- Then
	//
	if p.NoError(err) {
		p.EqualProviderConfig(expected, actual)
	}
}

func (p *PublicTestSuite) TestLoadJson() {
	// -- Given
	//
	expectedSettings := []byte(`{}`)
	testFs := fstest.MapFS{
		"provider.json": {
			Data: []byte(`
{
  "title": "minecraft",
	"version_sync": {
		"source": {
			"container_registry": {
				"matches_tag": "^(latest|java8|java9|java11|java15)$"
			}
		}
	},
	"image": "itzg/minecraft-server",
	"file_inputs": [
		{
			"accept_props": [".jar"],
			"help_text": "A *.zip file of your minecraft world data.",
			"description": "The save file that will be used by your server. We will automatically backup any new saves to [your save]_autosave.",
			"destination": {
				"bucket_folder": "saves"
			}
		}
	]
}
`),
		},
		"settings.json": {
			Data: expectedSettings,
		},
	}

	expected := &ProviderConfig{
		Config: &providerconfig.ProviderConfig{
			Title: "minecraft",
			VersionSync: &providerconfig.VersionSyncSpec{
				Source: &providerconfig.RemoteVersionSource{
					ContainerRegistry: &providerconfig.ContainerRegistryVersionSource{
						MatchesTag: "^(latest|java8|java9|java11|java15)$",
					},
				},
			},
			Image: "itzg/minecraft-server",
			FileInputs: []*providerconfig.FileInputParam{
				{
					AcceptProps: []string{".jar"},
					HelpText:    "A *.zip file of your minecraft world data.",
					Destination: &providerconfig.FileInputParamDestination{BucketFolder: "saves"},
					Description: "The save file that will be used by your server. We will automatically backup any new saves to [your save]_autosave.",
				},
			},
		},
		Filename:       "provider.json",
		SettingsSchema: &gojsonschema.Schema{},
		RawSettings:    expectedSettings,
	}

	// -- When
	//
	actual, err := DefaultClient.Load(testFs, "provider.json", "settings.json")

	// -- Then
	//
	if p.NoError(err) {
		p.EqualProviderConfig(expected, actual)
	}
}

func (p *PublicTestSuite) TestLoadMissingFile() {
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
	actual, err := p.Target.Load(testFs, "provider.yml", "settings.json")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, "open provider.yml: file does not exist")
}

func (p *PublicTestSuite) TestLoadInvalidYaml() {
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
	actual, err := p.Target.Load(testFs, "provider.yaml", "settings.json")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `invalid...` into map[string]interface {}")
}

func (p *PublicTestSuite) TestLoadInvalidJson() {
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
	actual, err := p.Target.Load(testFs, "provider.json", "settings.json")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, "invalid character 'i' looking for beginning of value")
}

func (p *PublicTestSuite) TestLoadMissingSettingsJson() {
	// -- Given
	//
	testFs := fstest.MapFS{
		"provider.json": {
			Data: []byte(`invalid file.`),
		},
	}

	// -- When
	//
	actual, err := p.Target.Load(testFs, "provider.json", "settings.json")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, "settings.json was not found. A JSON schema file is required for every provider")
}

func (p *PublicTestSuite) TestLoadInvalidSettingsJson() {
	// -- Given
	//
	testFs := fstest.MapFS{
		"provider.json": {
			Data: []byte(`{}`),
		},
		"settings.json": {
			Data: []byte(`invalid file.`),
		},
	}

	// -- When
	//
	actual, err := p.Target.Load(testFs, "provider.json", "settings.json")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, "invalid settings.json JSON schema file: invalid character 'i' looking for beginning of value")
}

func (p *PublicTestSuite) TestLoadInvalidFileExt() {
	// -- Given
	//
	testFs := fstest.MapFS{
		"provider.txt": {
			Data: []byte(`invalid file.`),
		},
		"settings.json": {
			Data: []byte(`{}`),
		},
	}

	// -- When
	//
	actual, err := p.Target.Load(testFs, "provider.txt", "settings.json")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, ".txt is not a supported extension for provider configs")
}

func (p *PublicTestSuite) TestLoadAll() {
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

	expected := []*ProviderConfig{
		{
			Config:         &providerconfig.ProviderConfig{Title: "factorio"},
			Filename:       "factorio/provider.yaml",
			SettingsSchema: &gojsonschema.Schema{},
			RawSettings:    expectedSettings,
		},
		{
			Config:         &providerconfig.ProviderConfig{Title: "minecraft"},
			Filename:       "minecraft/provider.yaml",
			SettingsSchema: &gojsonschema.Schema{},
			RawSettings:    expectedSettings,
		},
		{
			Config:         &providerconfig.ProviderConfig{Title: "valheim"},
			Filename:       "valheim/provider.yaml",
			SettingsSchema: &gojsonschema.Schema{},
			RawSettings:    expectedSettings,
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

func (p *PublicTestSuite) EqualProviderConfig(expected, actual *ProviderConfig, o ...EqualProviderConfigsOpt) bool {
	opts := &equalProviderConfigOpts{}
	for _, v := range o {
		opts = v(*opts)
	}

	if expected.SettingsSchema != nil {
		p.NotNil(actual.SettingsSchema)
	} else {
		p.Nil(actual.SettingsSchema)
	}

	p.Equal(expected.RawSettings, actual.RawSettings)
	p.Empty(cmp.Diff(expected.Config, actual.Config, protocmp.Transform()))
	return p.Equal(expected.Filename, actual.Filename)
}

func (p *PublicTestSuite) EqualProviderConfigs(expected, actual []*ProviderConfig, o ...EqualProviderConfigsOpt) bool {
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

func TestPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PublicTestSuite))
}
