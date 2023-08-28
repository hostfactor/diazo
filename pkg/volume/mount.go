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
	MountFile(vol *providerconfig.Volume, fp string, op ...opts.Opt[MountFileOpts]) (int64, error)

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

type MountFileOpts struct {
	DestPath string
}

func (m MountFileOpts) DefaultOptions() MountFileOpts {
	return MountFileOpts{}
}

// WithDestPath overwrite the default destination path that the file is mounted to.
func WithDestPath(dst string) opts.Opt[MountFileOpts] {
	return func(m *MountFileOpts) {
		m.DestPath = filepath.Clean(dst)
	}
}

type mounter struct {
	UserfilesClient userfiles.Client
	BasePath        string
}

func (m *mounter) MountFile(vol *providerconfig.Volume, fp string, op ...opts.Opt[MountFileOpts]) (int64, error) {
	o := opts.DefaultApply(op...)
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
	target := o.DestPath
	if target == "" {
		target = filepath.Join(vol.GetMount().GetPath(), filename)
	} else if filepath.Ext(target) == "" {
		target = filepath.Join(vol.GetMount().GetPath(), target, filename)
	} else {
		target = filepath.Join(vol.GetMount().GetPath(), target)
	}
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
