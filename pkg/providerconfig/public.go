package providerconfig

import (
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/hostfactor/api/go/providerconfig"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v3"
	"io/fs"
	"path/filepath"
)

var DefaultClient = NewClient()

var (
	DefaultFilename = "provider.yaml"
)

type ProviderConfig struct {
	Config   *providerconfig.ProviderConfig
	Filename string
}

type Client interface {
	// Load loads a single ProviderConfig from the specified directory.
	Load(f fs.FS, filename string) (*ProviderConfig, error)

	// LoadAll loads all provider configs within the directory. Assumes that every Provider directory is housed as a child
	// directory of the specified fs.FS.
	LoadAll(f fs.FS) ([]*ProviderConfig, error)
}

func NewClient() Client {
	return &client{}
}

func Load(f fs.FS, filename string) (*ProviderConfig, error) {
	return DefaultClient.Load(f, filename)
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

		conf, err := c.Load(sub, DefaultFilename)
		if err != nil {
			return nil, err
		}
		conf.Filename = filepath.Join(entry.Name(), conf.Filename)

		confs = append(confs, conf)
	}

	return confs, nil
}

func (c *client) Load(f fs.FS, filename string) (*ProviderConfig, error) {
	content, err := fs.ReadFile(f, filename)
	if err != nil {
		return nil, err
	}

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
	err = jsonpb.UnmarshalString(string(content), conf)
	if err != nil {
		return nil, err
	}

	out := &ProviderConfig{
		Config:   conf,
		Filename: filename,
	}

	return out, nil
}
