package dupescout

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/sync/semaphore"
)

type Cfg struct {
	Path         string           // path to search for duplicates
	IgnoreHidden bool             // ignore hidden files and directories
	KeyGenerator keyGeneratorFunc // key generator function to use
}

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

type pair struct {
	key  string // depends on the KeyGeneratorFunc§
	path string
}

// Map of keys to paths.
type results map[string][]string

type Dupes []string

func Start(c Cfg) Dupes {
	// Set defaults on provided config as needed.
	c.defaults()

	workers := runtime.GOMAXPROCS(0) * 2
	sem := semaphore.NewWeighted(int64(workers))
	pairs := make(chan pair)
	result := make(chan results)
	wg := new(sync.WaitGroup)

	go consumePairs(pairs, result)

	wg.Add(1)

	err := search(c.Path, pairs, sem, wg, &c)
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()

	close(pairs)

	return processDupes(result)
}

func consumePairs(pairs <-chan pair, result chan<- results) {
	m := make(results)

	for p := range pairs {
		m[p.key] = append(m[p.key], p.path)
	}

	result <- m
}

func search(dir string, pairs chan<- pair, sem *semaphore.Weighted, wg *sync.WaitGroup, c *Cfg) error {
	defer wg.Done()
	defer sem.Release(1)

	if err := sem.Acquire(context.Background(), 1); err != nil {
		return err
	}

	return filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if c.IgnoreHidden && strings.HasPrefix(fi.Name(), ".") {
			if fi.Mode().IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		if fi.Mode().IsDir() && path != dir {
			wg.Add(1)
			go search(path, pairs, sem, wg, c)
			return filepath.SkipDir
		}

		if fi.Mode().IsRegular() && fi.Size() > 0 {
			wg.Add(1)
			go processFile(path, pairs, sem, wg, c)
		}

		return nil
	})
}

func processFile(path string, pairs chan<- pair, sem *semaphore.Weighted, wg *sync.WaitGroup, c *Cfg) {
	defer wg.Done()
	defer sem.Release(1)

	if err := sem.Acquire(context.Background(), 1); err != nil {
		log.Fatal(err)
	}

	key, err := c.KeyGenerator(path)
	if err != nil {
		log.Fatal(err)
	}

	if key == "" {
		return
	}

	pairs <- pair{key, path}
}

func processDupes(result <-chan results) Dupes {
	dupes := Dupes{}

	for _, paths := range <-result {
		if len(paths) > 1 {
			dupes = append(dupes, paths...)
		}
	}

	return dupes
}
