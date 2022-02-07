package versiontransformer

import (
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/version/remoteversion"
)

type Transformer interface {
	TransformRemoteVersion(ver *remoteversion.RemoteVersion) *remoteversion.RemoteVersion
}

func CompileTransformers(spec *providerconfig.VersionSyncSpec) ([]Transformer, error) {
	transformers := make([]Transformer, 0)
	if v := spec.GetNameTransformer().GetSubstringReplace(); v != nil {
		replacer, err := NewSubstringReplace(v)
		if err != nil {
			return nil, err
		}
		transformers = append(transformers, replacer)
	}

	return transformers, nil
}
