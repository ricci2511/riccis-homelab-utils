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

// Checks if the provided path should be skipped based on file filters.
//
// Assumes that the path is a file.
func (f *Filters) skipFile(path string) bool {
	fileName := filepath.Base(path)
	if skipHidden(fileName, f.HiddenInclude) {
		return true
	}

	ext := strings.ToLower(filepath.Ext(fileName))

	// Include takes precedence over exclude.
	if len(f.ExtInclude) > 0 {
		for _, filter := range f.ExtInclude {
			if ext == filter {
				return false
			}
		}

		return true
	}

	if len(f.ExtExclude) > 0 {
		for _, filter := range f.ExtExclude {
			if ext == filter {
				return true
			}
		}
	}

	return false
}

// Checks if the provided path should be skipped based on dir filters.
//
// Assumes that the path is a directory.
func (f *Filters) skipDir(path string) bool {
	dirName := filepath.Base(path)
	if f.SkipSubdirs || skipHidden(dirName, f.HiddenInclude) {
		return true
	}

	if len(f.DirsExclude) == 0 {
		return false
	}

	for _, filter := range f.DirsExclude {
		if dirName == filter {
			return true
		}
	}

	return false
}

// Helper to check if the provided dir or file name is hidden and should be skipped
// based on the HiddenInclude filter.
func skipHidden(name string, hiddenInclude bool) bool {
	if !hiddenInclude && strings.HasPrefix(name, ".") {
		return true
	}

	return false
}
