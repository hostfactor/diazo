package providerconfig

import (
	"github.com/google/go-cmp/cmp"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/stretchr/testify/suite"
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
		Filename: "provider.yaml",
	}

	// -- When
	//
	actual, err := DefaultClient.Load(testFs, "provider.yaml")

	// -- Then
	//
	if p.NoError(err) {
		p.Empty(cmp.Diff(expected, actual, protocmp.Transform()))
		p.Equal(expected, actual)
	}
}

func (p *PublicTestSuite) TestLoadJson() {
	// -- Given
	//
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
		Filename: "provider.json",
	}

	// -- When
	//
	actual, err := DefaultClient.Load(testFs, "provider.json")

	// -- Then
	//
	if p.NoError(err) {
		p.Empty(cmp.Diff(expected, actual, protocmp.Transform()))
		p.Equal(expected, actual)
	}
}

func (p *PublicTestSuite) TestLoadMissingFile() {
	// -- Given
	//
	testFs := fstest.MapFS{"provider.yaml": {}}

	// -- When
	//
	actual, err := p.Target.Load(testFs, "provider.yml")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, "open provider.yml: file does not exist")
}

func (p *PublicTestSuite) TestLoadInvalidYaml() {
	// -- Given
	//
	testFs := fstest.MapFS{"provider.yaml": {
		Data: []byte(`
invalid file.
`),
	}}

	// -- When
	//
	actual, err := p.Target.Load(testFs, "provider.yaml")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, "yaml: unmarshal errors:\n  line 2: cannot unmarshal !!str `invalid...` into map[string]interface {}")
}

func (p *PublicTestSuite) TestLoadInvalidJson() {
	// -- Given
	//
	testFs := fstest.MapFS{"provider.json": {
		Data: []byte(`
invalid file.
`),
	}}

	// -- When
	//
	actual, err := p.Target.Load(testFs, "provider.json")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, "invalid character 'i' looking for beginning of value")
}

func (p *PublicTestSuite) TestLoadInvalidFileExt() {
	// -- Given
	//
	testFs := fstest.MapFS{"provider.txt": {
		Data: []byte(`
invalid file.
`),
	}}

	// -- When
	//
	actual, err := p.Target.Load(testFs, "provider.txt")

	// -- Then
	//
	p.Nil(actual)
	p.EqualError(err, ".txt is not a supported extension for provider configs")
}

func (p *PublicTestSuite) TestLoadAll() {
	// -- Given
	//
	testFs := fstest.MapFS{
		"minecraft/provider.yaml": {
			Data: []byte(`
title: minecraft
`),
		},
		"factorio/provider.yaml": {
			Data: []byte(`
title: factorio
`),
		},
		"valheim/provider.yaml": {
			Data: []byte(`
title: valheim
`),
		},
		"random_file.txt": {},
	}

	expected := []*ProviderConfig{
		{
			Config:   &providerconfig.ProviderConfig{Title: "factorio"},
			Filename: "factorio/provider.yaml",
		},
		{
			Config:   &providerconfig.ProviderConfig{Title: "minecraft"},
			Filename: "minecraft/provider.yaml",
		},
		{
			Config:   &providerconfig.ProviderConfig{Title: "valheim"},
			Filename: "valheim/provider.yaml",
		},
	}

	// -- When
	//
	actual, err := p.Target.LoadAll(testFs)

	// -- Then
	//
	if p.NoError(err) {
		p.Empty(cmp.Diff(expected, actual, protocmp.Transform()))
	}
}

func TestPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PublicTestSuite))
}
