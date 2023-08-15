package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ricci2511/riccis-homelab-utils/dupescout"
)

func main() {
	cfg := dupescout.Cfg{}
	cfg.KeyGenerator = keyGeneratorSelect()
	flag.StringVar(&cfg.Path, "path", "", "path to search for duplicates")
	flag.BoolVar(&cfg.Filters.SkipSubdirs, "skip-subdirs", false, "skip subdirectories traversal")
	flag.BoolVar(&cfg.Filters.HiddenInclude, "incl-hidden", false, "ignore hidden files and directories")
	flag.Var(&cfg.Filters.ExtInclude, "incl-exts", "extensions to include")
	flag.Var(&cfg.Filters.ExtExclude, "excl-exts", "extensions to exclude")
	flag.Var(&cfg.Filters.DirsExclude, "excl-dirs", "directories or subdirectories to exclude")
	flag.IntVar(&cfg.Workers, "workers", 0, "number of workers (defaults to GOMAXPROCS)")
	flag.Parse()

	done := make(chan struct{})
	go loadingSpinner(done)

	dupes := []string{}
	dupesChan := make(chan string)

	// Start the duplicate search in its own goroutine.
	go func(cfg dupescout.Cfg, dupesChan chan<- string) {
		err := dupescout.StreamResults(cfg, dupesChan)
		if err != nil {
			log.Println(err)
		}
	}(cfg, dupesChan)

	// Append a human readable size to each received duplicate path.
	for dupePath := range dupesChan {
		fi, err := os.Stat(dupePath)
		if err != nil {
			log.Println(err)
			continue
		}
		s := fmt.Sprintf("%s (%s)", dupePath, humanReadableSize(fi.Size()))
		dupes = append(dupes, s)
	}

	// Close done channel right after all duplicates have been found.
	close(done)

	if len(dupes) == 0 {
		fmt.Printf("\nNo duplicates found with the provided configuration: %s\n", cfg.String())
		os.Exit(0)
	}

	prompt := &survey.MultiSelect{
		Message:  "Select duplicates to delete:",
		Options:  dupes,
		PageSize: 10,
	}

	selectedDupes := []string{}
	err := survey.AskOne(prompt, &selectedDupes)
	if err != nil {
		log.Fatal(err)
	}

	for _, path := range selectedDupes {
		fmt.Printf("Deleting: %s\n", path)
		err := os.Remove(path)
		if err != nil {
			log.Fatal(err)
		}
	}
}

var earthSpinner = []string{"🌍", "🌎", "🌏"}

func loadingSpinner(done <-chan struct{}) {
	i := 0
	l := len(earthSpinner)

	// Return when done channel is closed, otherwise keep spinning.
	for {
		select {
		case <-done:
			return
		default:
			fmt.Printf("\rScanning... %s", earthSpinner[i])
			i = (i + 1) % l
			time.Sleep(150 * time.Millisecond)
		}
	}
}

func humanReadableSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	// Calculate the divisor and exponent of the unit symbol to use (KiB, MiB, GiB, etc).
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

type keyGeneratorPair struct {
	description string
	fn          dupescout.KeyGeneratorFunc
}

var keygenMap = map[string]keyGeneratorPair{
	"MovieTvFileNamesKeyGenerator": {
		description: "Detects movie/tv show files based on the file name. Useful for detecting repeated movies/tv episodes even if they are different files.",
		fn:          movieTvFileNamesKeyGenerator, // custom key generator function
	},
	"Crc32HashKeyGenerator": {
		description: "Generates a crc32 hash of the first 16KB of the file contents. Should be enough to achieve a good balance of uniqueness, collision resistance, and performance for most files.",
		fn:          dupescout.Crc32HashKeyGenerator,
	},
	"FullCrc32HashKeyGenerator": {
		description: "Generates a crc32 hash of the entire file contents. A lot slower than HashKeyGenerator but should be more accurate.",
		fn:          dupescout.FullCrc32HashKeyGenerator,
	},
	"Sha256HashKeyGenerator": {
		description: "Generates a sha256 hash of the first 16KB of the file contents.",
		fn:          dupescout.Sha256HashKeyGenerator,
	},
	"FullSha256HashKeyGenerator": {
		description: "Generates a sha256 hash of the entire file contents. Again, a lot slower but should be more accurate.",
		fn:          dupescout.FullSha256HashKeyGenerator,
	},
}

// Prompts the user to select a key generator function and returns it.
func keyGeneratorSelect() dupescout.KeyGeneratorFunc {
	var keygenFnNames []string
	for fnName := range keygenMap {
		keygenFnNames = append(keygenFnNames, fnName)
	}

	prompt := &survey.Select{
		Message: "Select a key generator function:",
		Options: keygenFnNames,
		Description: func(val string, _ int) string {
			return keygenMap[val].description
		},
	}

	var keygenFnName string
	err := survey.AskOne(prompt, &keygenFnName)
	if err != nil {
		log.Fatal(err)
	}

	return keygenMap[keygenFnName].fn
}
