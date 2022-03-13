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

func NewMounter(cl userfiles.Client) Mounter {
	return &mounter{
		UserfilesClient: cl,
	}
}

type mounter struct {
	UserfilesClient userfiles.Client
}

func (m *mounter) MountFileSelection(vol *providerconfig.Volume, sel *filesystem.FileSelection) (int64, error) {
	mountPath := vol.GetMount().GetPath()

	bytesDownloaded := int64(0)
	cl := fileloc.New(m.UserfilesClient)
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
