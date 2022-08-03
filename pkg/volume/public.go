package volume

import (
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/fileloc"
	"github.com/hostfactor/diazo/pkg/userfiles"
)

type Mounter interface {
	MountFileSelection(vol *providerconfig.Volume, sel *filesystem.FileSelection) (int64, error)
}

// NewMounter creates a new Mounter. Ultimately, it acts as a wrapper to convert filesystem definitions to userfiles.Client
// calls. The basePath is a prefix path used for all keys with the userfiles.Client.
func NewMounter(cl userfiles.Client, basePath string) Mounter {
	return &mounter{
		UserfilesClient: cl,
		BasePath:        basePath,
	}
}

type mounter struct {
	UserfilesClient userfiles.Client
	BasePath        string
}

func (m *mounter) MountFileSelection(vol *providerconfig.Volume, sel *filesystem.FileSelection) (int64, error) {
	mountPath := vol.GetMount().GetPath()

	bytesDownloaded := int64(0)
	cl := fileloc.New(m.UserfilesClient, m.BasePath)
	for _, loc := range sel.GetLocations() {
		w, err := cl.Download(loc, mountPath)
		if err != nil {
			return bytesDownloaded, err
		}
		bytesDownloaded += w.Size
	}

	return bytesDownloaded, nil
}

func VolumesToMap(vols []*providerconfig.Volume) map[string]*providerconfig.Volume {
	m := map[string]*providerconfig.Volume{}
	for i, vol := range vols {
		m[vol.GetName()] = vols[i]
	}

	return m
}
