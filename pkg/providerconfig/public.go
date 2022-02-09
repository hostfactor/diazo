package providerconfig

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/hostfactor/api/go/providerconfig"
	jsoniter "github.com/json-iterator/go"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
	"io/fs"
	"path/filepath"
)

var DefaultClient = NewClient()

var (
	DefaultProviderFilename = "provider.yaml"
	DefaultSettingsFilename = "settings.json"
)

type ProviderConfig struct {
	Config       *providerconfig.ProviderConfig
	SettingsFile *gojsonschema.Schema
	Filename     string
}

type Client interface {
	// Load loads a single ProviderConfig from the specified directory.
	Load(f fs.FS, providerFilename, settingsFilename string) (*ProviderConfig, error)

	// LoadAll loads all provider configs within the directory. Assumes that every Provider directory is housed as a child
	// directory of the specified fs.FS.
	LoadAll(f fs.FS) ([]*ProviderConfig, error)
}

func NewClient() Client {
	return &client{}
}

func Load(f fs.FS) (*ProviderConfig, error) {
	return DefaultClient.Load(f, DefaultProviderFilename, DefaultSettingsFilename)
}

type client struct {
}

func (c *client) LoadAll(f fs.FS) ([]*ProviderConfig, error) {
	entries, err := fs.ReadDir(f, ".")
	if err != nil {
		return nil, err
	}

	confs := make([]*ProviderConfig, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sub, err := fs.Sub(f, entry.Name())
		if err != nil {
			return nil, err
		}

		conf, err := c.Load(sub, DefaultProviderFilename, DefaultSettingsFilename)
		if err != nil {
			return nil, err
		}
		conf.Filename = filepath.Join(entry.Name(), conf.Filename)

		confs = append(confs, conf)
	}

	return confs, nil
}

func (c *client) Load(f fs.FS, providerFilename, settingsFilename string) (*ProviderConfig, error) {
	settingsContent, err := fs.ReadFile(f, settingsFilename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%s was not found. A JSON schema file is required for every provider", settingsFilename)
		}
		return nil, err
	}

	out := &ProviderConfig{
		Filename: providerFilename,
	}

	loader := gojsonschema.NewBytesLoader(settingsContent)
	out.SettingsFile, err = gojsonschema.NewSchema(loader)
	if err != nil {
		return nil, fmt.Errorf("invalid %s JSON schema file: %s", settingsFilename, err.Error())
	}

	providerContent, err := fs.ReadFile(f, providerFilename)
	if err != nil {
		return nil, err
	}

	out.Config, err = c.unmarshalProviderFile(providerFilename, providerContent)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *client) unmarshalProviderFile(filename string, content []byte) (*providerconfig.ProviderConfig, error) {
	ext := filepath.Ext(filename)
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
	err := jsonpb.UnmarshalString(string(content), conf)
	if err != nil {
		return nil, err
	}

	return conf, err
}
