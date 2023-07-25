# dupescout
A tiny Go package to concurrently find and return duplicate file paths within the given directory and including its subdirectories.

## Installation
```bash
go get github.com/ricci2511/riccis-homelab-utils/dupescout
```

## Usage
The package exposes a single function `Find` which takes a `dupescout.Cfg` struct and returns a slice of duplicate file paths.
Check out [dedupsc](https://github.com/ricci2511/riccis-homelab-utils/tree/main/dedupsc) for another example of how this package can be used.

```go
package main

import (
    "fmt"
    "github.com/ricci2511/riccis-homelab-utils/dupescout"
)

func main() {
    filters:= dupescout.Filters{
        HiddenInclude: true,
        DirsExclude: []string{"node_modules"},
        ExtInclude: []string{".txt", ".json", ".go"}, // only search for .txt, .json and .go files
    }
    cfg := dupescout.Cfg{
        Path: "~/Dev",
        Filters: filters,
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
	KeyGenerator KeyGeneratorFunc // key generator function to use
	Workers      int              // number of workers (defaults to GOMAXPROCS)
}
```

The `KeyGenerator` field allows you to specify a custom function to generate a key for a given file path that maps to a slice of duplicate file paths.

Some functions are already provided, the default one being `dupescout.Crc32HashKeyGenerator` which simply hashes the first 16KB of file contents with `crc32`. The functions prefixed with `Full` hash the entire file contents instead of just the first 16KB, which is way slower but should be more accurate for rare cases where the first 16KB are not enough. Available `KeyGenerator` functions are:

- `dupescout.Crc32HashKeyGenerator`
- `dupescout.FullCrc32HashKeyGenerator`
- `dupescout.Sha256HashKeyGenerator`
- `dupescout.FullSha256HashKeyGenerator`

In case you want to use custom logic to generate keys, you simply pass a function that satisfies the `dupescout.KeyGeneratorFunc`. An example can be found [here](https://github.com/ricci2511/riccis-homelab-utils/blob/main/dedupsc/movie-tv-key-generator.go).
