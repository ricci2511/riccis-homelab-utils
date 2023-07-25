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

	flag.StringVar(&cfg.Path, "path", "", "path to search for duplicates")
	flag.BoolVar(&cfg.Filters.SkipSubdirs, "skip-subdirs", false, "skip subdirectories traversal")
	flag.BoolVar(&cfg.Filters.HiddenInclude, "incl-hidden", false, "ignore hidden files and directories")
	flag.Var(&cfg.Filters.ExtInclude, "incl-exts", "extensions to include")
	flag.Var(&cfg.Filters.ExtExclude, "excl-exts", "extensions to exclude")
	flag.Var(&cfg.Filters.DirsExclude, "excl-dirs", "directories or subdirectories to exclude")
	flag.IntVar(&cfg.Workers, "workers", 0, "number of workers (defaults to GOMAXPROCS)")
	flag.Parse()

	cfg.KeyGenerator = keyGeneratorSelect()

	done := make(chan struct{})
	go loadingSpinner(done)

	// Start the search for dupes, blocks until all of them have been processed.
	dupes, err := dupescout.Find(cfg)
	if err != nil {
		log.Println(err)
	}

	// Close done channel right after the search is done, triggering the spinner to stop.
	close(done)

	if len(dupes) == 0 {
		fmt.Printf("No duplicates found with the provided configuration: %s\n", cfg.String())
		os.Exit(0)
	}

	selectedDupes := []string{}

	prompt := &survey.MultiSelect{
		Message:  "Select duplicates to delete:",
		Options:  dupes,
		PageSize: 10,
	}

	err = survey.AskOne(prompt, &selectedDupes)
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

	// Break when done channel is closed, otherwise keep spinning.
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

type keyGeneratorPair struct {
	description string
	fn          dupescout.KeyGeneratorFunc
}

var keygenMap = map[string]keyGeneratorPair{
	"HashKeyGenerator": {
		description: "Generates a crc32 hash of the first 16KB of the file contents, which should be enough to achieve a good balance of uniqueness, collision resistance, and performance for most files.",
		fn:          dupescout.HashKeyGenerator,
	},
	"FullHashKeyGenerator": {
		description: "Generates a crc32 hash of the entire file contents. A lot slower than HashKeyGenerator but should be more accurate.",
		fn:          dupescout.FullHashKeyGenerator,
	},
	"MovieTvFileNamesKeyGenerator": {
		description: "Detects movie/tv show files based on the file name. Useful for detecting repeated movies/tv episodes even if they are different files.",
		fn:          dupescout.MovieTvFileNamesKeyGenerator,
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
