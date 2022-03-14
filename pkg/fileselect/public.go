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

// GetLastFilename gets the last filename from the list of filesystem.FileLocation. The filename will only consist of
// the name and the extension e.g. "save.zip".
func GetLastFilename(locs []*filesystem.FileLocation) string {
	if len(locs) == 0 {
		return ""
	}

	return GetLocationFilename(locs[len(locs)-1])
}

// GetFilenames gathers all filenames from a slice of locations. The filenames will only consist of the name and the extension e.g.
// "save.zip".
func GetFilenames(fs []*filesystem.FileLocation) []string {
	out := make([]string, 0, len(fs))
	for _, v := range fs {
		name := GetLocationFilename(v)
		if name != "" {
			out = append(out, name)
		}
	}

	return out
}

// GetLocationFilename retrieves the filename from the filesystem.FileLocation. The filename will consist of the name
// and the extension e.g. "save.zip".
func GetLocationFilename(loc *filesystem.FileLocation) string {
	if f := loc.GetBucketFile(); f != nil {
		return f.Name
	}

	return ""
}
