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

// Satisfies the flag.Value interface, string values can be provided as a csv or space separated list.
//
// `flag.Var(&cfg.Paths, "p", "list of paths to search in for duplicates")`
type Paths []string

func (p *Paths) String() string {
	return ""
}

func (p *Paths) Set(val string) error {
	vals := strings.FieldsFunc(val, func(r rune) bool {
		return r == ' ' || r == ','
	})
	*p = append(*p, vals...)
	return nil
}

type Cfg struct {
	KeyGenerator KeyGeneratorFunc // Function to generate a key based on the file path.
	Paths                         // List of paths to search in for duplicates.
	Filters                       // Filters to apply when searching for duplicates.
	Workers      int              // Number of workers to use when searching for duplicates.
}

// Beauty stringifies the Cfg struct.
func (c *Cfg) String() string {
	keygenFn := runtime.FuncForPC(reflect.ValueOf(c.KeyGenerator).Pointer())
	keygenFnName := filepath.Base(keygenFn.Name())

	return fmt.Sprintf(
		"\n{\n\tPath: %s\n\tFilters: \n%s\n\tKeyGenerator: %s\n}",
		c.Paths,
		c.Filters.String(),
		keygenFnName,
	)
}

// Sanitizes the provided path, supports ~ and ~username.
func sanitizePath(path string) string {
	if strings.HasPrefix(path, "~") {
		firstSlash := strings.Index(path, "/")

		if firstSlash == 1 {
			home, err := os.UserHomeDir()
			if err != nil {
				log.Fatal(err)
			}

			path = strings.Replace(path, "~", home, 1)
		} else if firstSlash > 1 {
			username := path[1:firstSlash]
			userAccount, err := user.Lookup(username)
			if err != nil {
				log.Fatal(err)
			}

			path = strings.Replace(path, path[:firstSlash], userAccount.HomeDir, 1)
		}
	}

	path, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	return path
}

// Sets default values for the cfg struct as needed.
func (c *Cfg) defaults() {
	for i, path := range c.Paths {
		if path == "" {
			c.Paths[i] = "." // Default to current directory
			continue
		}

		c.Paths[i] = sanitizePath(path)
	}

	if c.KeyGenerator == nil {
		c.KeyGenerator = Crc32HashKeyGenerator // Default to CRC32 (fast and sufficient for most cases)
	}

	if c.Workers == 0 {
		c.Workers = runtime.GOMAXPROCS(0) / 2
	}
}
