package provideractions

import (
	"github.com/hostfactor/api/go/blueprint/filesystem"
)

// NameMatcher creates an filesystem.FileMatcher_Name which matches the exact name of a file.
func NameMatcher(name string) *filesystem.FileMatcher {
	return &filesystem.FileMatcher{
		Expression: &filesystem.FileMatcher_Expression{Name: name},
	}
}

// GlobMatcher creates an filesystem.FileMatcher_Glob which matches all files with the same glob expression.
func GlobMatcher(globs ...string) *filesystem.FileMatcher {
	return &filesystem.FileMatcher{
		Expression: &filesystem.FileMatcher_Expression{Glob: &filesystem.GlobMatcher{Value: globs}},
	}
}

// RegexMatcher creates an filesystem.FileMatcher_Regex which matches all files that match the regex.
func RegexMatcher(regex string) *filesystem.FileMatcher {
	return &filesystem.FileMatcher{Expression: &filesystem.FileMatcher_Expression{Regex: regex}}
}
