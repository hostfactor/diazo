package fileactions

import (
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/diazo/pkg/fileutils"
	"github.com/hostfactor/diazo/pkg/userfiles"
	"github.com/mattn/go-zglob"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func MatchBucketFiles(cli userfiles.Client, key userfiles.Key, folder string, matcher *filesystem.FileMatcher) ([]*userfiles.FileReader, error) {
	if matcher.GetExpression().GetName() != "" {
		r, err := cli.FetchFileReader(path.Join(userfiles.GenFolderKey(folder, key), matcher.GetExpression().GetName()))
		if err != nil {
			return nil, err
		}
		return []*userfiles.FileReader{r}, nil
	} else if matcher.GetExpression().GetGlob() != nil || matcher.GetExpression().GetRegex() != "" {
		handles, err := cli.ListUserFolder(folder, key)
		if err != nil {
			return nil, err
		}

		out := make([]*userfiles.FileReader, 0, len(handles))

		for _, v := range handles {
			r, err := cli.FetchFileReader(v.Key)
			if err != nil {
				return nil, err
			}

			if MatchGlob(v.Name, matcher.GetExpression().GetGlob()) || MatchRegex(v.Name, matcher.GetExpression().GetRegex()) {
				out = append(out, r)
			}
		}
		return out, nil
	}
	return []*userfiles.FileReader{}, nil
}

// MatchPath matches a relative or absolute path to a file matcher.
func MatchPath(p string, matcher *filesystem.FileMatcher) (matched bool) {
	if matcher.GetExpression().GetRegex() != "" {
		matched = MatchRegex(p, matcher.GetExpression().GetRegex())
		if matched {
			return
		}
	}

	if matcher.GetExpression().GetName() != "" {
		_, filename := filepath.Split(p)
		matched = filename == matcher.GetExpression().GetName()
		if matched {
			return
		}
	}

	if matcher.GetExpression().GetGlob() != nil {
		matched = MatchGlob(p, matcher.GetExpression().GetGlob())
	}

	return
}

func MatchGlob(fp string, matcher *filesystem.GlobMatcher) (matched bool) {
	for _, g := range matcher.GetValue() {
		m, _ := zglob.Match(strings.TrimSpace(g), fp)
		matched = matched || m
	}
	return
}

func MatchRegex(fp string, regex string) (matched bool) {
	if regex == "" {
		return
	}
	re, err := regexp.Compile(regex)
	if err == nil {
		matched = re.MatchString(fp)
	}
	return
}

func MatchDirectoryFile(f fs.FS, matcher *filesystem.DirectoryFileMatcher) (matched bool) {
	d := fileutils.FsPath(f)
	if d == matcher.GetDirectory() {
		m := GetFsMatches(f, matcher.GetMatches())
		return len(m) > 0
	}
	return
}

func GetFsMatches(f fs.FS, matcher *filesystem.FileMatcher) []string {
	matched := make([]string, 0)
	if matcher.GetExpression().GetGlob() != nil {
		for _, v := range matcher.GetExpression().GetGlob().GetValue() {
			globs, err := fs.Glob(f, strings.TrimSpace(v))
			if err != nil {
				continue
			}

			matched = append(matched, globs...)
		}
	}

	if matcher.GetExpression().GetName() != "" || matcher.GetExpression().GetRegex() != "" {
		de, _ := fs.ReadDir(f, ".")

		if matcher.GetExpression().GetName() != "" {
			for _, v := range de {
				if v.Name() == matcher.GetExpression().GetName() {
					matched = append(matched, matcher.GetExpression().GetName())
				}
			}
		}

		if matcher.GetExpression().GetRegex() != "" {
			re, err := regexp.Compile(matcher.GetExpression().GetRegex())
			if err == nil {
				for _, v := range de {
					if re.MatchString(v.Name()) {
						matched = append(matched, v.Name())
					}
				}
			}
		}
	}

	return matched
}
