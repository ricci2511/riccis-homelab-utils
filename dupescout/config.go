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

type Cfg struct {
	Path         string           // path to search for duplicates
	Filters                       // various filters for the search
	KeyGenerator KeyGeneratorFunc // key generator function to use
	Workers      int              // number of workers (defaults to GOMAXPROCS)
}

// Beauty stringifies the Cfg struct.
func (c *Cfg) String() string {
	keygenFn := runtime.FuncForPC(reflect.ValueOf(c.KeyGenerator).Pointer())
	keygenFnName := filepath.Base(keygenFn.Name())

	return fmt.Sprintf(
		"\n{\n\tPath: %s\n\tFilters: \n%s\n\tKeyGenerator: %s\n}",
		c.Path,
		c.Filters.String(),
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
		c.KeyGenerator = Crc32HashKeyGenerator
	}

	if c.Workers == 0 {
		c.Workers = runtime.GOMAXPROCS(0)
	}
}
