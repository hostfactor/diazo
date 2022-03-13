package provideractions

import (
	"github.com/hostfactor/api/go/blueprint/filesystem"
)

type FileMatcher interface {
	FileMatcher() *filesystem.FileMatcher
}

type nameMatcher filesystem.FileMatcher_Name

func (n *nameMatcher) FileMatcher() *filesystem.FileMatcher {
	out := filesystem.FileMatcher_Name(*n)
	return &filesystem.FileMatcher{File: &out}
}

type regexMatcher filesystem.FileMatcher_Regex

func (n *regexMatcher) FileMatcher() *filesystem.FileMatcher {
	out := filesystem.FileMatcher_Regex(*n)
	return &filesystem.FileMatcher{File: &out}
}

type globMatcher filesystem.FileMatcher_Glob

func (n *globMatcher) FileMatcher() *filesystem.FileMatcher {
	out := filesystem.FileMatcher_Glob(*n)
	return &filesystem.FileMatcher{File: &out}
}

// NameMatcher creates an filesystem.FileMatcher_Name which matches the exact name of a file.
func NameMatcher(name string) FileMatcher {
	return &nameMatcher{
		Name: name,
	}
}

// GlobMatcher creates an filesystem.FileMatcher_Glob which matches all files with the same glob expression.
func GlobMatcher(globs ...string) FileMatcher {
	return &globMatcher{
		Glob: &filesystem.GlobMatcher{Value: globs},
	}
}

// RegexMatcher creates an filesystem.FileMatcher_Regex which matches all files that match the regex.
func RegexMatcher(regex string) FileMatcher {
	return &regexMatcher{Regex: regex}
}
