package dupescout

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/puzpuzpuz/xsync/v2"
	"golang.org/x/sync/errgroup"
)

type pair struct {
	key  string // depends on the KeyGeneratorFunc
	path string
}

type dupescout struct {
	g           *errgroup.Group  // "wait group" to limit the num of concurrent search workers
	pairs       chan *pair       // channel to send pairs to, which are processed and sent to the caller
	shutdown    chan os.Signal   // channel to receive shutdown signals on
	generatorFn KeyGeneratorFunc // function that generates a key for a given path to identify duplicates
	filters     Filters          // filters to apply when searching for duplicates
}

func newDupeScout(c Cfg) *dupescout {
	g := new(errgroup.Group)
	g.SetLimit(c.Workers)

	return &dupescout{
		g:           g,
		pairs:       make(chan *pair, c.Workers),
		shutdown:    make(chan os.Signal, 1),
		generatorFn: c.KeyGenerator,
		filters:     c.Filters,
	}
}

// Starts the search for duplicates which can be customized by the provided Cfg struct.
func run(c Cfg, dupesChan chan []string, stream bool) error {
	c.defaults()
	dup := newDupeScout(c)

	go dup.consumePairs(dupesChan, stream)
	go gracefulShutdown(dup.shutdown)

	for _, path := range c.Paths {
		p := path
		dup.g.Go(func() error {
			return dup.search(p)
		})
	}

	err := dup.g.Wait()
	close(dup.pairs) // Trigger pair consumer to process the results.
	return err
}

// Runs the duplicate search and returns a slice of all duplicate paths.
func GetResults(c Cfg) ([]string, error) {
	dupesChan := make(chan []string, 1)
	err := run(c, dupesChan, false)
	return <-dupesChan, err
}

// Runs the duplicate search and streams the duplicate paths to the provided channel
// as they are found.
func StreamResults(c Cfg, dupesChan chan []string) error {
	return run(c, dupesChan, true)
}

// Processes the produced pairs and sends the results to the provided channel.
// Depending on the stream bool, results are either sent in chunks or all at once.
func (dup *dupescout) consumePairs(dupesChan chan []string, stream bool) {
	defer close(dupesChan)

	// key -> last encountered path
	m := xsync.NewMapOf[string]()

	for p := range dup.pairs {
		storedPath, ok := m.Load(p.key)
		if !ok {
			m.Store(p.key, p.path)
			continue
		}
		// When storedPath is not empty, it indicates that we have found the first duplicate,
		// so we send both the stored path and the current path.
		if storedPath != "" {
			m.Store(p.key, "")
			paths := []string{storedPath, p.path}
			if stream {
				// Send in chunks.
				dupesChan <- paths
				continue
			}
			// Not streaming, collect all duplicate paths.
			select {
			case currDupes := <-dupesChan:
				dupesChan <- append(currDupes, paths...)
			default:
				dupesChan <- paths
			}
			continue
		}
		// Previous duplicate paths have already been sent, so just send the current path.
		if stream {
			dupesChan <- []string{p.path}
			continue
		}
		// Select is not needed since at this point the dupesChan can't be empty or full.
		dupesChan <- append(<-dupesChan, p.path)
	}
}

// Produces a pair with the key which is generated by `dup.generatorFn` and the path
// which is then sent to the pairs channel.
func (dup *dupescout) producePair(path string) error {
	if dup.shuttingDown() {
		return nil // Stop pair production if shutdown is in progress.
	}

	key, err := dup.generatorFn(path)
	if err != nil {
		if errors.Is(err, ErrSkipFile) {
			return nil // Don't collect ErrSkipFile errors
		}
		return err
	}

	if key == "" {
		return fmt.Errorf("\nkey generator returned an empty key for path: %s", path)
	}

	dup.pairs <- &pair{key, path}
	return nil
}

// Walks the tree of the provided dir and triggers the production of pairs for each valid file.
func (dup *dupescout) search(dir string) error {
	return filepath.WalkDir(dir, func(path string, de os.DirEntry, err error) error {
		if dup.shuttingDown() {
			return nil
		}

		if err != nil {
			return err
		}

		if de.IsDir() && dup.filters.skipDir(path) {
			return filepath.SkipDir
		}

		if de.Type().IsRegular() && !dup.filters.skipFile(path) {
			fi, err := de.Info()
			if err != nil || fi.Size() == 0 {
				return nil
			}

			dup.g.Go(func() error {
				return dup.producePair(path)
			})
		}

		return nil
	})
}

// Helper to check if a shutdown signal has been received.
func (dup *dupescout) shuttingDown() bool {
	select {
	case <-dup.shutdown:
		return true
	default:
		return false
	}
}

// Sets up a signal handler worker for graceful shutdown.
func gracefulShutdown(shutdown chan os.Signal) {
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	log.Println("\nReceived signal, shutting down after current workers are done...")
	close(shutdown)
}
