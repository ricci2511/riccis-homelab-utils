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

// The `filters` type satisfies the flag.Value interface, therefore it can be used
// as follows:
//
// `flag.Var(&cfg.ExtInclude, "include", "extensions to include")`
//
// Extensions can be provided as a csv or space separated list.
type Filters []string

func (f *Filters) String() string {
	return ""
}

func (f *Filters) Set(val string) error {
	vals := strings.FieldsFunc(val, func(r rune) bool {
		return r == ' ' || r == ','
	})
	*f = append(*f, vals...)
	return nil
}

type Cfg struct {
	Path         string           // path to search for duplicates
	IgnoreHidden bool             // ignore hidden files and directories
	ExtInclude   Filters          // file extensions to include (higher priority than exclude)
	ExtExclude   Filters          // file extensions to exclude
	KeyGenerator keyGeneratorFunc // key generator function to use
	Workers      int              // number of workers (defaults to GOMAXPROCS)
}

// Beauty prints the cfg struct.
func (c *Cfg) String() string {
	keygenFn := runtime.FuncForPC(reflect.ValueOf(c.KeyGenerator).Pointer())
	keygenFnName := filepath.Base(keygenFn.Name())

	return fmt.Sprintf(
		"\n{\n\tPath: %s\n\tIgnoreHidden: %t\n\tExtInclude: %s\n\tExtExclude: %s\n\tKeyGenerator: %s\n}",
		c.Path,
		c.IgnoreHidden,
		c.ExtInclude,
		c.ExtExclude,
		keygenFnName,
	)
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

	if c.Workers == 0 {
		c.Workers = runtime.GOMAXPROCS(0)
	}
}

// Checks if the provided path satisfies the extension filter.
// The path is expected to be a file.
func (c *Cfg) satisfiesExtFilter(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	// Include takes precedence over exclude.
	if len(c.ExtInclude) > 0 {
		for _, filter := range c.ExtInclude {
			if ext == filter {
				return true
			}
		}

		return false
	}

	if len(c.ExtExclude) > 0 {
		for _, filter := range c.ExtExclude {
			if ext == filter {
				return false
			}
		}
	}

	return true
}
