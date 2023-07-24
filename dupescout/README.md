# dupescout
A tiny Go package to concurrently find and return duplicate file paths within the given directory and including its subdirectories.

## Installation
```bash
go get github.com/ricci2511/riccis-homelab-utils/dupescout
```

## Usage
The package exposes a single function `Start` which takes a `dupescout.Cfg` struct and returns a slice of duplicate file paths.

```go
package main

import (
    "fmt"
    "github.com/ricci2511/riccis-homelab-utils/dupescout"
)

func main() {
    cfg := dupescout.Cfg{
        Path: "~/Dev",
		Filters: dupescout.Filters{
			HiddenInclude: true,
			DirsExclude: []string{"node_modules"},
			ExtInclude: []string{".txt", ".json", ".go"}, // only search for .txt, .json and .go files
		},
    }

    fmt.Println("Searching...")

    // Blocks until the search is complete
    dupes := dupescout.Find(cfg)

    fmt.Println("Search complete")

    for _, path := range selectedDupes {
        fmt.Println(path)
    }
}
```

The `dupescout.Cfg` struct has the following fields as of now (more to come probably):

```go
type Cfg struct {
	Path         string           // path to search for duplicates
	Filters                       // various filters for the search (see filters.go)
	KeyGenerator keyGeneratorFunc // key generator function to use
	Workers      int              // number of workers (defaults to GOMAXPROCS)
}
```

The `KeyGenerator` field allows you to specify a custom function to generate a key for a given file path that maps to a slice of duplicate file paths.
Some `KeyGenerator` functions are already provided, the default one being `dupescout.HashKeyGenerator` which simply hashes the first 16KB of file contents with `crc32` and returns it encoded as a hex string. Other provided functions are:

- `dupescout.FullHashKeyGenerator` is similar to the regular one, except it hashes the entire file contents instead of just the first 16KB. This will be much slower for large files, but should be more accurate for rare cases where the first 16KB are not enough.
- `dupescout.MovieTvFileNamesKeyGenerator` returns the movie or tv show name, along with the season and episode number and year if applicable as the key. This is useful for finding duplicate movies that have different qualities, codecs, etc. Example: `Interstellar - 2014 - Bluray-2160p.mkv` and `Interstellar.2014.1080p.BluRay.x264.mkv` will result in the same key `Interstellar-2014`, and thus will be considered duplicates. STILL WORK IN PROGRESS
- Maybe more to come if I find any other useful cases.
