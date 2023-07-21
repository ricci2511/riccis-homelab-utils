package dupescout

import (
	"crypto/md5"
	"io"
	"os"
)

// KeyGenerator is a function for generating keys for a given path
// which will be used to compare against the keys of other paths to
// determine if they are duplicates.
//
// Feel free to implement your custom KeyGenerator.
type keyGeneratorFunc func(path string) (string, error)

// HashKeyGenerator is the default if no KeyGenerator is specified.
//
// Generates a md5 hash of the file contents and uses it as the key.
func HashKeyGenerator(path string) (string, error) {
	file, err := os.Open(path)

	if err != nil {
		return "", err
	}

	defer file.Close()

	hash := md5.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return string(hash.Sum(nil)), nil
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
