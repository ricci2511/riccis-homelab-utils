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

// Dupescout is the main struct that holds the state of the search.
type dupescout struct {
	g        *errgroup.Group
	pairs    chan *pair
	shutdown chan os.Signal
}

func newDupeScout(workers int) *dupescout {
	g := new(errgroup.Group)
	g.SetLimit(workers)

	return &dupescout{
		g:        g,
		pairs:    make(chan *pair, 500),
		shutdown: make(chan os.Signal, 1),
	}
}

// Starts the search for duplicates which can be customized by the provided Cfg struct.
func run(c Cfg, dupesChan chan []string, stream bool) error {
	c.defaults()
	dup := newDupeScout(c.Workers)

	go dup.consumePairs(dupesChan, stream)
	go gracefulShutdown(dup.shutdown)

	dup.g.Go(func() error {
		return dup.search(c.Path, &c)
	})

	err := dup.g.Wait()
	close(dup.pairs) // Trigger the pair consumer to process the results.
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
// Depending on the stream bool, results are either sent as they are found or
// once the search is done.
func (dup *dupescout) consumePairs(dupesChan chan []string, stream bool) {
	defer close(dupesChan)
	m := xsync.NewMapOf[[]string]()

	for p := range dup.pairs {
		paths, ok := m.Load(p.key)
		if !ok {
			m.Store(p.key, []string{p.path})
			continue
		}
		// We can save memory by just keeping the key.
		// Since paths are being sent, we don't need to keep them all in the map.
		m.Store(p.key, nil)
		// At this point, paths slice is either nil or contains one element. If it's nil,
		// the previous paths related to this key have already been sent, otherwise we
		// send the stored path along with the current path.
		if len(paths) == 1 {
			paths = []string{paths[0], p.path}
			if stream {
				// Streaming, send in chunks.
				dupesChan <- paths
				continue
			}
			// Not streaming, we collect all duplicate paths.
			select {
			case currDupes := <-dupesChan:
				dupesChan <- append(currDupes, paths...)
			default:
				dupesChan <- paths
			}
			continue
		}
		if stream {
			dupesChan <- []string{p.path}
			continue
		}
		// Select is not needed since at this point the dupesChan can't be empty or full.
		dupesChan <- append(<-dupesChan, p.path)
	}
}

// Produces a pair with the key which is generated by the KeyGeneratorFunc and the path
// which is then sent to the pairs channel.
func (dup *dupescout) producePair(path string, keyGen KeyGeneratorFunc) error {
	// Stop pair production if a shutdown signal has been received.
	if dup.shuttingDown() {
		return nil
	}

	key, err := keyGen(path)
	if err != nil {
		if errors.Is(err, ErrSkipFile) {
			return nil // don't collect ErrSkipFile errors
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
func (dup *dupescout) search(dir string, c *Cfg) error {
	return filepath.WalkDir(dir, func(path string, de os.DirEntry, err error) error {
		// Stop searching if a shutdown signal has been received.
		if dup.shuttingDown() {
			return nil
		}

		if err != nil {
			return err
		}

		if de.IsDir() && c.skipDir(path) {
			return filepath.SkipDir
		}

		if de.Type().IsRegular() && !c.skipFile(path) {
			fi, err := de.Info()
			if err != nil || fi.Size() == 0 {
				return nil
			}

			dup.g.Go(func() error {
				return dup.producePair(path, c.KeyGenerator)
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
