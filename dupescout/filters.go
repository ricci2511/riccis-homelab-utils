package dupescout

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"
)

// Satisfies the flag.Value interface, string values can be provided as a csv or space separated list.
//
// `flag.Var(&cfg.DirsExclude "exclude-dirs", "exclude directories or subdirectories")
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

	if len(f.ExtInclude) > 0 {
		return !slices.Contains(f.ExtInclude, ext) // Skip files not in include list
	}

	return slices.Contains(f.ExtExclude, ext) // Skip files in exclude list
}

// Checks if the provided path should be skipped based on dir filters.
//
// Assumes that the path is a directory.
func (f *Filters) skipDir(path string) bool {
	dirName := filepath.Base(path)
	if f.SkipSubdirs || skipHidden(dirName, f.HiddenInclude) {
		return true
	}

	return slices.Contains(f.DirsExclude, dirName) // Skip dirs in exclude list
}

// Helper to check if the provided dir or file name is hidden and should be skipped
// based on the HiddenInclude filter.
func skipHidden(name string, hiddenInclude bool) bool {
	return !hiddenInclude && strings.HasPrefix(name, ".")
}
