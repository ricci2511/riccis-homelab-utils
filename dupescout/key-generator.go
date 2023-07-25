package dupescout

import (
	"encoding/hex"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// KeyGenerator is a function for generating keys for a given path
// which will be used to compare against the keys of other paths to
// determine if they are duplicates.
//
// Feel free to implement your custom KeyGenerator.
type KeyGeneratorFunc func(path string) (string, error)

func crc32HexString(path string, full bool) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer file.Close()

	hash := crc32.NewIEEE()

	// Either copy the entire file contents or just the first 16KB.
	if full {
		_, err = io.Copy(hash, file)
	} else {
		_, err = io.CopyN(hash, file, 1024*16)
	}

	if err != nil && err != io.EOF {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// HashKeyGenerator is the default if no KeyGenerator is specified.
//
// Generates a crc32 hash of the first 16KB of the file contents as the key,
// which should be enough to achieve a good balance of uniqueness, collision
// resistance, and performance for most files.
func HashKeyGenerator(path string) (string, error) {
	return crc32HexString(path, false)
}

// Generates a crc32 hash of the entire file contents as the key, which
// is a lot slower than HashKeyGenerator but should be more accurate.
func FullHashKeyGenerator(path string) (string, error) {
	return crc32HexString(path, true)
}

// MovieTvFileNamesKeyGenerator patterns and variables
var (
	tvPattern    = regexp.MustCompile(`(.+?)\s*(S\d+E\d+|\d+x\d+|S\d+E\d+-E\d+)`)
	moviePattern = regexp.MustCompile(`(?:.+?)(?:\s*[-.]\s*|\s+)(\d{4})`)
	ilegalChars  = ".- ()[]{},:;_"
)

// MovieTvFileNamesKeyGenerator generates a key based on the movie or series title.
// Discarding the quality, resolution, codec, release group, etc.
//
// Example: "Avatar - 2009 - Bluray-1080p" and "Avatar.2009.Bluray.720p" will
// both generate the key: "Avatar2009", and will be considered possible duplicates.
//
// Extensions other than .mkv, .mp4, .avi, .wmv, .mov, .flv, .webm, .mpeg are ignored.
func MovieTvFileNamesKeyGenerator(path string) (string, error) {
	fileName := filepath.Base(path)

	if !validVideoExt(filepath.Ext(fileName)) {
		return "", nil
	}

	if matches := tvPattern.FindStringSubmatch(fileName); matches != nil && len(matches) > 2 {
		// "Breaking Bad - S01E01 - Bluray-1080p" -> "BreakingBadS01E01"
		// matches[1] = "Breaking Bad", matches[2] = "S01E01"
		tvKey := matches[1] + matches[2]
		return removeChars(tvKey, ilegalChars), nil
	}

	if matches := moviePattern.FindStringSubmatch(fileName); matches != nil && len(matches) > 0 {
		// "Avatar - 2009 - Bluray-1080p" -> "Avatar2009"
		// matches[0] = "Avatar - 2009", matches[1] = "2009"
		return removeChars(matches[0], ilegalChars), nil
	}

	return fileName, nil
}

func removeChars(s, chars string) string {
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(chars, r) {
			return -1
		}

		return r
	}, s)
}

func validVideoExt(ext string) bool {
	switch ext {
	case ".mkv", ".mp4", ".avi", ".wmv", ".mov", ".flv", ".webm", ".mpeg":
		return true
	}

	return false
}
