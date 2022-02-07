package diazo

import (
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/containers"
	"github.com/hostfactor/diazo/pkg/version/remoteversion"
	"github.com/hostfactor/diazo/pkg/version/versiontransformer"
)

var DefaultClient Client

type Client interface {
	SyncVersions(syncer VersionSyncer, spec *providerconfig.VersionSyncSpec) error
}

type VersionSyncer interface {
	SyncVersions(rv *remoteversion.RemoteVersion) error
}

type VersionSyncerFunc func(rv *remoteversion.RemoteVersion) error

func (v VersionSyncerFunc) SyncVersions(rv *remoteversion.RemoteVersion) error { return v(rv) }

func NewClient(conf *providerconfig.ProviderConfig, conClient containers.Client) Client {
	cl := &client{
		ContainerClient: conClient,
		Config:          conf,
	}

	return cl
}

type client struct {
	ContainerClient containers.Client
	Config          *providerconfig.ProviderConfig
}

func (c *client) SyncVersions(syncer VersionSyncer, spec *providerconfig.VersionSyncSpec) error {
	transformers, err := versiontransformer.CompileTransformers(spec)
	if err != nil {
		return err
	}

	if spec.GetSource().GetContainerRegistry() != nil {
		registry := remoteversion.NewContainerRegistrySource(c.Config.Image, c.ContainerClient)
		vers, err := registry.FetchVersions(spec.GetSource())
		if err != nil {
			return err
		}

		for _, ver := range vers {
			for _, transformer := range transformers {
				ver = transformer.TransformRemoteVersion(ver)
			}
			if err := syncer.SyncVersions(ver); err != nil {
				return err
			}
		}
	}

	return nil
}
