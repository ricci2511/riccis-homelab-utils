package dupescout

import (
	"encoding/hex"
	"hash/crc32"
	"io"
	"os"
)

// KeyGenerator is a function for generating keys for a given path
// which will be used to compare against the keys of other paths to
// determine if they are duplicates.
//
// Feel free to implement your custom KeyGenerator.
type keyGeneratorFunc func(path string) (string, error)

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

// MovieTvFileNamesKeyGenerator generates a key based on the movie or series title.
// Discarding the quality, resolution, codec, release group, etc.
//
// Example: "Avatar - 2009 - Bluray-1080p" and "Avatar - 2009 - Bluray-720p" will
// both generate the same key: "Avatar-2009", and will be considered possible duplicates.
//
// Extensions other than .mkv, .mp4, .avi, .wmv, .mov, .flv, .webm, .mpeg are ignored.
func MovieTvFileNamesKeyGenerator(path string) (string, error) {
	// TODO
	return "", nil
}
