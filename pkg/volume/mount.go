package volume

import (
	"fmt"
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/actions/fileactions"
	"github.com/hostfactor/diazo/pkg/collection"
	"github.com/hostfactor/diazo/pkg/filesys"
	"io/fs"
	"path/filepath"
	"strings"
)

type Mount interface {
	fs.ReadFileFS
	fs.ReadDirFS
	fmt.Stringer
}

func NewMount(vol *providerconfig.VolumeMount, f fs.FS) Mount {
	policies := collection.Map(vol.GetAccess(), func(f *filesystem.FileAccessPolicy) *filesys.AccessPolicy {
		return filesys.NewAccessPolicy(f)
	})

	recurs, other := collection.Split(policies, func(t *filesys.AccessPolicy) bool {
		return t.Wrapped.GetRecursive()
	})

	return &mount{
		FS:                f,
		Mount:             vol,
		RootPolicies:      other,
		RecursivePolicies: recurs,
	}
}

type mount struct {
	FS    fs.FS
	Mount *providerconfig.VolumeMount

	RootPolicies      []*filesys.AccessPolicy
	RecursivePolicies []*filesys.AccessPolicy
}

func (m *mount) String() string {
	return m.Mount.GetPath()
}

func (m *mount) ReadDir(fp string) ([]fs.DirEntry, error) {
	fp, policies := m.cleanFp(fp)

	if len(policies) == 0 {
		return []fs.DirEntry{}, nil
	}

	de, err := fs.ReadDir(m.FS, fp)
	if err != nil {
		return nil, err
	}

	de = collection.Filter(de, func(t fs.DirEntry) bool {
		return m.hasAccess(filepath.Join(fp, t.Name()), filesys.FilePermRead, policies)
	})

	return collection.Map(de, func(f fs.DirEntry) fs.DirEntry {
		return filesys.NewDirEntry(fp, f)
	}), nil
}

func (m *mount) Open(fp string) (fs.File, error) {
	fp, policies := m.cleanFp(fp)

	if !m.hasAccess(fp, filesys.FilePermRead, policies) {
		return nil, fs.ErrNotExist
	}

	f, err := m.FS.Open(fp)
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(fp)

	return filesys.NewFile(filepath.Join(m.String(), dir), f), nil
}

func (m *mount) ReadFile(fp string) ([]byte, error) {
	fp, policies := m.cleanFp(fp)

	if !m.hasAccess(fp, filesys.FilePermRead, policies) {
		return nil, fs.ErrNotExist
	}

	return fs.ReadFile(m.FS, fp)
}

func (m *mount) hasAccess(fp string, op filesys.FilePerm, policies []*filesys.AccessPolicy) bool {
	lvl := filesys.FilePerm(0)
	for _, ap := range policies {
		if ap.Perm == 0 {
			return false
		}
		if ap.Wrapped.GetMatches() == nil || fileactions.MatchPath(fp, ap.Wrapped.GetMatches()) {
			lvl |= ap.Perm
		}
	}
	return lvl&op != 0
}

func (m *mount) cleanFp(fp string) (string, []*filesys.AccessPolicy) {
	fp = strings.Trim(fp, "/ ")

	policies := m.RecursivePolicies
	if !strings.Contains(fp, "/") {
		policies = append(policies, m.RootPolicies...)
	}
	return fp, policies
}
