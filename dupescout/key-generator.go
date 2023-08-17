package dupescout

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"os"
)

var (
	// Used to skip a file during key generation.
	//
	// These kind of errors are ignored and not returned to the caller
	// of dupescout.GetResults() or dupescout.StreamResults().
	ErrSkipFile = fmt.Errorf("skip file")
)

// KeyGenerator generates a key for a given file path, which then is mapped to
// a list of file paths that share the same key (duplicates).
//
// The provided KeyGeneratorFuncs hash the file contents to generate the key, but
// the logic can be anything as long as it's deterministic. For example, you could
// generate a key based on the file name, size, etc.
type KeyGeneratorFunc func(path string) (string, error)

func generateFileHash(path string, hash hash.Hash, full bool) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer file.Close()

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

// Crc32HashKeyGenerator is the default if no KeyGenerator is specified.
//
// Generates a crc32 hash of the first 16KB of the file contents as the key,
// which should be enough to achieve a good balance of uniqueness, collision
// resistance, and performance for most files.
func Crc32HashKeyGenerator(path string) (string, error) {
	return generateFileHash(path, crc32.NewIEEE(), false)
}

// Generates a crc32 hash of the entire file contents as the key, which
// is a lot slower than HashKeyGenerator but should be more accurate.
func FullCrc32HashKeyGenerator(path string) (string, error) {
	return generateFileHash(path, crc32.NewIEEE(), true)
}

// Generates a sha256 hash of the first 16KB of the file contents as the key
func Sha256HashKeyGenerator(path string) (string, error) {
	return generateFileHash(path, sha256.New(), false)
}

// Generates a sha256 hash of the entire file contents as the key
func FullSha256HashKeyGenerator(path string) (string, error) {
	return generateFileHash(path, sha256.New(), true)
}
