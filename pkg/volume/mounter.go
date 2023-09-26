package volume

import (
	"github.com/eddieowens/opts"
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/fileutils"
	"github.com/hostfactor/diazo/pkg/userfiles"
	"os"
	"path"
	"path/filepath"
)

type Mounter interface {
	MountVolume(vol *providerconfig.Volume, selections []string, op ...opts.Opt[MountOpts]) (Mount, error)

	MountFile(vol *providerconfig.Volume, fp string) (int64, error)

	// MountFileSelection Deprecated. Use MountFile instead.
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

type MountOpts struct {
	HandleErr func(err error)
}

func (m MountOpts) DefaultOptions() MountOpts {
	return MountOpts{}
}

func WithHandleErr(f func(err error)) opts.Opt[MountOpts] {
	return func(m *MountOpts) {
		m.HandleErr = f
	}
}

type mounter struct {
	UserfilesClient userfiles.Client
	BasePath        string
}

func (m *mounter) MountVolume(vol *providerconfig.Volume, selections []string, op ...opts.Opt[MountOpts]) (Mount, error) {
	o := opts.DefaultApply(op...)
	mnt := NewMount(vol.GetMount(), os.DirFS(vol.GetMount().GetPath()))

	for _, v := range selections {
		_, err := m.MountFile(vol, v)
		if err != nil {
			if o.HandleErr != nil {
				o.HandleErr(err)
				continue
			} else {
				return nil, err
			}
		}
	}

	return mnt, nil
}

func (m *mounter) MountFile(vol *providerconfig.Volume, fp string) (int64, error) {
	name := vol.Name
	if name == "save" {
		// maintain backwards compat.
		name = "saves"
	}

	reader, err := m.UserfilesClient.FetchFileReader(path.Join(m.BasePath, name, fp))
	if err != nil {
		return 0, err
	}

	defer func() {
		_ = reader.Reader.Close()
	}()

	_, filename := path.Split(reader.Key)
	target := filepath.Join(vol.GetMount().GetPath(), filename)
	_ = os.MkdirAll(filepath.Dir(target), os.ModePerm)

	return fileutils.WriteFileFromReader(target, reader.Reader)
}

func (m *mounter) MountFileSelection(vol *providerconfig.Volume, sel *filesystem.FileSelection) (int64, error) {
	bytesDownloaded := int64(0)
	for _, loc := range sel.GetLocations() {
		w, err := m.MountFile(vol, loc.GetBucketFile().GetName())
		if err != nil {
			return bytesDownloaded, err
		}
		bytesDownloaded += w
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
