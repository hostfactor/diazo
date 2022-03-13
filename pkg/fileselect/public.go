package fileselect

import "github.com/hostfactor/api/go/blueprint/filesystem"

// MergeFileSelect merges two filesystems into a single slice. If there is a name conflict, f2 takes precedence.
func MergeFileSelect(f1, f2 []*filesystem.FileSelection) []*filesystem.FileSelection {
	out := make([]*filesystem.FileSelection, 0, len(f1)+len(f2))
	for i := range f1 {
		out = append(out, f1[i])
	}

	for i, v := range f2 {
		found, idx := FindFileSelection(v.GetVolumeName(), out)
		if found == nil {
			out = append(out, f2[i])
		} else {
			out[idx] = f2[i]
		}
	}

	return out
}

func FindFileSelection(name string, fs []*filesystem.FileSelection) (*filesystem.FileSelection, int) {
	for i, v := range fs {
		if v.GetVolumeName() == name {
			return v, i
		}
	}
	return nil, -1
}
