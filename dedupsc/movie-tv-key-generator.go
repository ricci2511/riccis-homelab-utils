package main

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Supported episode naming patterns: S01E01, 1x01, S01E01-E02, 1x01-02
	tvPattern    = regexp.MustCompile(`(.+?)\s*(S\d+E\d+(?:-E\d+)?|\d+x\d+(?:-\d+)?|S\d+E\d+-\d+)`)
	moviePattern = regexp.MustCompile(`(?:.+?)(?:\s*[-.]\s*|\s+)(\d{4})`)
	ilegalChars  = ".- ()[]{},:;_"
)

// Custom KeyGenerator function to generate a key based on the movie or series title.
// Discarding the quality, resolution, codec, release group, etc.
//
// Example: "Avatar - 2009 - Bluray-1080p" and "Avatar.2009.Bluray.720p" will
// both generate the key: "Avatar2009", and will be considered possible duplicates.
func movieTvFileNamesKeyGenerator(path string) (string, error) {
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
