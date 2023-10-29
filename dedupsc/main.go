package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ricci2511/riccis-homelab-utils/dupescout"
)

func main() {
	cfg := dupescout.Cfg{}
	cfg.KeyGenerator = keyGeneratorSelect()
	flag.Var(&cfg.Paths, "p", "paths to search for duplicates")
	flag.BoolVar(&cfg.Filters.SkipSubdirs, "sd", false, "skip directories traversal")
	flag.BoolVar(&cfg.Filters.HiddenInclude, "ih", false, "ignore hidden files and directories")
	flag.Var(&cfg.Filters.ExtInclude, "ie", "extensions to include")
	flag.Var(&cfg.Filters.ExtExclude, "ee", "extensions to exclude")
	flag.Var(&cfg.Filters.DirsExclude, "ed", "directories or subdirectories to exclude")
	flag.IntVar(&cfg.Workers, "w", 0, "number of workers (defaults to GOMAXPROCS)")
	logPaths := flag.Bool("l", false, "duplicate results will be logged to stdout")
	flag.Parse()

	// When logging, loading spinner is redundant.
	var done chan struct{}
	if !*logPaths {
		done = make(chan struct{})
		go loadingSpinner(done)
	}

	dupes := []string{}
	dupesChan := make(chan []string, 10)

	// Start the duplicate search in its own goroutine.
	go func(cfg dupescout.Cfg, dupesChan chan []string) {
		err := dupescout.StreamResults(cfg, dupesChan)
		if err != nil {
			log.Println(err)
		}
	}(cfg, dupesChan)

	// Append a human readable size to each received duplicate path.
	for dupePaths := range dupesChan {
		for _, path := range dupePaths {
			fi, err := os.Stat(path)
			if err != nil {
				log.Println(err)
				continue
			}
			s := fmt.Sprintf("%s (%s)", path, humanReadableSize(fi.Size()))
			if *logPaths {
				fmt.Println(s)
			}
			dupes = append(dupes, s)
		}
	}

	// Close done channel right after all duplicates have been found.
	if done != nil {
		close(done)
	}

	if len(dupes) == 0 {
		fmt.Printf("\nNo duplicates found with the provided configuration: %s\n", cfg.String())
		os.Exit(0)
	}

	prompt := &survey.MultiSelect{
		Message:  "Delete selected files:",
		Options:  dupes,
		PageSize: 10,
	}

	selectedDupes := []string{}
	err := survey.AskOne(prompt, &selectedDupes)
	if err != nil {
		log.Fatal(err)
	}

	sizeSuffixRegex := regexp.MustCompile(` \(.+\)$`)
	for _, path := range selectedDupes {
		path = sizeSuffixRegex.ReplaceAllString(path, "")
		err := os.Remove(path)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Deleted: %s\n", path)
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
		description: "Detects and groups the same movie/tv shows based on the file name.",
		fn:          movieTvFileNamesKeyGenerator, // custom key generator function
	},
	"AudioCodecKeyGenerator": {
		description: "Groups video files together based on their audio codec.",
		fn:          audioCodecKeyGenerator(""), // custom key generator function (closure)
	},
	"Crc32HashKeyGenerator": {
		description: "Generates a crc32 hash of the first 16KB of the file contents.",
		fn:          dupescout.Crc32HashKeyGenerator,
	},
	"FullCrc32HashKeyGenerator": {
		description: "Generates a crc32 hash of the entire file contents. Slower, but more accurate.",
		fn:          dupescout.FullCrc32HashKeyGenerator,
	},
	"Sha256HashKeyGenerator": {
		description: "Generates a sha256 hash of the first 16KB of the file contents.",
		fn:          dupescout.Sha256HashKeyGenerator,
	},
	"FullSha256HashKeyGenerator": {
		description: "Generates a sha256 hash of the entire file contents. Slower, but more accurate.",
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

	if keygenFnName == "AudioCodecKeyGenerator" {
		// Make the user input the audio codec to group files by.
		prompt := &survey.Input{
			Message: "Enter the audio codec to group files by:",
			Help:    "Examples: aac, ac3, dts, mp3, vorbis, flac, opus, etc.",
		}
		var audioCodec string
		err := survey.AskOne(prompt, &audioCodec)
		if err != nil {
			log.Fatal(err)
		}
		// Return a closure that returns a KeyGeneratorFunc with the provided audio codec.
		return audioCodecKeyGenerator(audioCodec)
	}

	return keygenMap[keygenFnName].fn
}
