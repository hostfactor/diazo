package doccache

import (
	"fmt"
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/providerconfig"
	"io/fs"
	"os"
	"path/filepath"
)

type DocCache interface {
	Get() *blueprint.Docs
}

func New(f fs.FS, pr *providerconfig.ProviderConfig) (DocCache, error) {
	out := &blueprint.Docs{
		Entries: make([]*blueprint.Docs_Entry, 0, len(pr.GetDocs().GetEntries())),
	}

	for _, v := range pr.GetDocs().GetEntries() {
		var content []byte
		var err error
		if filepath.IsAbs(v.GetPath()) {
			content, err = os.ReadFile(v.GetPath())
		} else {
			content, err = fs.ReadFile(f, v.GetPath())
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read doc entry %s: %w", v.GetPath(), err)
		}

		out.Entries = append(out.Entries, &blueprint.Docs_Entry{
			Title:   v.GetTitle(),
			Content: string(content),
		})
	}

	return &docCache{
		Docs: out,
	}, nil
}

type docCache struct {
	Docs *blueprint.Docs
}

func (d *docCache) Get() *blueprint.Docs {
	return d.Docs
}
