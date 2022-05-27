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

func FindFileSelection(volName string, fs []*filesystem.FileSelection) (*filesystem.FileSelection, int) {
	for i, v := range fs {
		if v.GetVolumeName() == volName {
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

func MapByVolumeName(selectedFiles []*filesystem.FileSelection) map[string][]*filesystem.FileSelection {
	selectedFilesByVolName := map[string][]*filesystem.FileSelection{}
	for idx, v := range selectedFiles {
		val := selectedFilesByVolName[v.VolumeName]
		if len(val) == 0 {
			val = []*filesystem.FileSelection{}
		}
		val = append(val, selectedFiles[idx])
		selectedFilesByVolName[v.VolumeName] = val
	}

	return selectedFilesByVolName
}

// Clean compacts' filesystem.FileSelection by aggregating all files in the same volume into a singular slice and performs
// any necessary transformations
func Clean(fs []*filesystem.FileSelection) []*filesystem.FileSelection {
	out := make([]*filesystem.FileSelection, 0, len(fs))

	for _, v := range fs {
		found, idx := FindFileSelection(v.VolumeName, out)
		if found != nil {
			out[idx].Locations = CleanFileLocations(MergeFileLocations(found.GetLocations(), v.GetLocations()))
		} else {
			out = append(out, v)
		}
	}

	return out
}

func CleanFileLocations(f []*filesystem.FileLocation) []*filesystem.FileLocation {
	out := make([]*filesystem.FileLocation, 0, len(f))
	for _, v := range f {
		if !InvalidFileLocation(v) {
			out = append(out, v)
		}
	}
	return out
}

func InvalidFileLocation(f *filesystem.FileLocation) bool {
	if f == nil {
		return true
	}
	return f.BucketFile == nil || f.GetBucketFile().GetFolder() == "" || f.GetBucketFile().GetName() == ""
}

// MergeFileLocations merges filesystem.FileLocation by their destination ensuring there aren't any duplicates. If there
// is a duplicate, l2 takes precedence.
func MergeFileLocations(l1, l2 []*filesystem.FileLocation) []*filesystem.FileLocation {
	out := make([]*filesystem.FileLocation, 0, len(l1)+len(l2))
	for i := range l1 {
		out = append(out, l1[i])
	}

	for i, v := range l2 {
		idx := FindFileLocation(v, out)
		if idx < 0 {
			out = append(out, l2[i])
		} else {
			out[idx] = l2[i]
		}
	}

	return out
}

func EqualFileLocations(l1, l2 *filesystem.FileLocation) bool {
	return l1.GetBucketFile().GetName() == l2.GetBucketFile().GetName() &&
		l1.GetBucketFile().GetFolder() == l2.GetBucketFile().GetFolder()
}

func FindFileLocation(l *filesystem.FileLocation, sl []*filesystem.FileLocation) int {
	for i, v := range sl {
		if EqualFileLocations(l, v) {
			return i
		}
	}
	return -1
}
