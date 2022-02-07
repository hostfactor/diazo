package remoteversion

import "github.com/hostfactor/api/go/providerconfig"

type VersionFetcher interface {
	FetchVersions(cr *providerconfig.RemoteVersionSource) ([]*RemoteVersion, error)
}

type RemoteVersion struct {
	Name string
}
