package dupescout

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Satisfies the flag.Value interface, string values can be provided as a csv or space separated list.
//
// `flag.Var(&cfg.HiddenInclude "include-hidden", "include hidden files and directories")`
type FiltersList []string

func (fl *FiltersList) String() string {
	return ""
}

func (fl *FiltersList) Set(val string) error {
	vals := strings.FieldsFunc(val, func(r rune) bool {
		return r == ' ' || r == ','
	})
	*fl = append(*fl, vals...)
	return nil
}

type Filters struct {
	SkipSubdirs   bool        // skip subdirectories traversal
	HiddenInclude bool        // ignore hidden files and directories
	ExtInclude    FiltersList // file extensions to include (higher priority than exclude)
	ExtExclude    FiltersList // file extensions to exclude
	DirsExclude   FiltersList // directories or subdirectories to exclude
}

// Beauty stringifies the Filters struct.
func (f *Filters) String() string {
	return fmt.Sprintf(
		"\t{\n\t\tSkipSubdirs: %t\n\t\tHiddenInclude: %t\n\t\tExtInclude: %s\n\t\tExtExclude: %s\n\t\tDirsExclude: %s\n\t}",
		f.SkipSubdirs,
		f.HiddenInclude,
		f.ExtInclude,
		f.ExtExclude,
		f.DirsExclude,
	)
}

// Checks if the provided path satisfies the extension filter.
//
// Assumes that the path is a file.
func (f *Filters) satisfiesExtFilter(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	// Include takes precedence over exclude.
	if len(f.ExtInclude) > 0 {
		for _, filter := range f.ExtInclude {
			if ext == filter {
				return true
			}
		}

		return false
	}

	if len(f.ExtExclude) > 0 {
		for _, filter := range f.ExtExclude {
			if ext == filter {
				return false
			}
		}
	}

	return true
}

// Checks if the provided path should be skipped based on dir filters.
//
// Assumes that the path is a subdirectory.
func (f *Filters) skipDir(path string) bool {
	if f.SkipSubdirs {
		return true
	}

	if len(f.DirsExclude) == 0 {
		return false
	}

	folder := filepath.Base(path)

	for _, filter := range f.DirsExclude {
		if folder == filter {
			return true
		}
	}

	return false
}
