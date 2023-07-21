package dupescout

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

// ExtFilter is a slice of file extensions to filter by.
//
// This type satisfies the flag.Value interface, therefore it can be used
// as follows: flag.Var(&extFilter, "ext", "filter by file extension")
//
// Extensions can be provided as a csv or space separated list.
type ExtFilter []string

func (e *ExtFilter) String() string {
	return ""
}

func (e *ExtFilter) Set(val string) error {
	vals := strings.FieldsFunc(val, func(r rune) bool {
		return r == ' ' || r == ','
	})
	*e = append(*e, vals...)
	return nil
}

type Cfg struct {
	Path         string           // path to search for duplicates
	IgnoreHidden bool             // ignore hidden files and directories
	ExtFilter                     // filter by file extension
	KeyGenerator keyGeneratorFunc // key generator function to use
}

// Beauty prints the cfg struct.
func (c *Cfg) String() string {
	keygenFn := runtime.FuncForPC(reflect.ValueOf(c.KeyGenerator).Pointer())
	keygenFnName := filepath.Base(keygenFn.Name())

	return fmt.Sprintf("\n{\n\tPath: %s\n\tIgnoreHidden: %t\n\tExtFilter: %s\n\tKeyGenerator: %s\n}", c.Path, c.IgnoreHidden, c.ExtFilter, keygenFnName)
}

// Sanitizes the provided path, supports ~ and ~username.
func (c *Cfg) sanitizePath() {
	if strings.HasPrefix(c.Path, "~") {
		firstSlash := strings.Index(c.Path, "/")

		if firstSlash == 1 {
			home, err := os.UserHomeDir()
			if err != nil {
				log.Fatal(err)
			}

			c.Path = strings.Replace(c.Path, "~", home, 1)
		} else if firstSlash > 1 {
			username := c.Path[1:firstSlash]
			userAccount, err := user.Lookup(username)
			if err != nil {
				log.Fatal(err)
			}

			c.Path = strings.Replace(c.Path, c.Path[:firstSlash], userAccount.HomeDir, 1)
		}
	}

	path, err := filepath.Abs(c.Path)
	if err != nil {
		log.Fatal(err)
	}

	c.Path = path
}

// Sets default values for the cfg struct as needed.
func (c *Cfg) defaults() {
	if c.Path == "" {
		c.Path = "."
		fmt.Println("No path specified, scanning from current directory")
	} else {
		c.sanitizePath()
	}

	// Default to hash key generator.
	if c.KeyGenerator == nil {
		c.KeyGenerator = HashKeyGenerator
	}
}

// Checks if the provided path satisfies the extension filter.
// The path is expected to be a file.
func (c *Cfg) satisfiesExtFilter(path string) bool {
	// Just return early if no filter is set.
	if len(c.ExtFilter) == 0 {
		return true
	}

	ext := strings.ToLower(filepath.Ext(path))

	for _, filter := range c.ExtFilter {
		if ext == filter {
			return true
		}
	}

	return false
}
