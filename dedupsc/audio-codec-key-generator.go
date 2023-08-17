package main

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ricci2511/riccis-homelab-utils/dupescout"
)

// Groups video files together based on their audio codec.
// Only files with the provided audio codec argument will be processed,
// others will be skipped.
//
// Dependencies: ffprobe
func audioCodecKeyGenerator(audioCodec string) dupescout.KeyGeneratorFunc {
	return func(path string) (string, error) {
		fileName := filepath.Base(path)
		if !validVideoExt(filepath.Ext(fileName)) {
			return "", nil
		}

		execPath, err := exec.LookPath("ffprobe")
		if err != nil {
			return "", err
		}

		cmd := exec.Command(execPath, "-v", "error", "-select_streams", "a:0", "-show_entries", "stream=codec_name", "-of", "default=noprint_wrappers=1:nokey=1", path)
		output, err := cmd.Output()
		if err != nil {
			return "", err
		}

		codec := string(output)
		if strings.TrimSpace(codec) == string(audioCodec) {
			return codec, nil
		}

		return "", dupescout.ErrSkipFile
	}
}
