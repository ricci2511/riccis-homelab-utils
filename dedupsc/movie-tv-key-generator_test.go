package main

import "testing"

func TestMovieTvFileNamesKeyGenerator(t *testing.T) {
	tcs := []struct {
		path     string
		expected string
	}{
		{"Alien - 1979 - Bluray-1080p.mkv", "Alien1979"},
		{"Alien.1979.1080p.BluRay.x265-RelGroup.mp4", "Alien1979"},
		{"8 Mile - 2002 - Bluray-1080p.mkv", "8Mile2002"},
		{"8.Mile.2002.1080p.BluRay.x265-RelGroup.mp4", "8Mile2002"},
		{"Breaking Bad - S01E01 - Pilot.mkv", "BreakingBadS01E01"},
		{"Breaking Bad - 1x01 - Pilot.mkv", "BreakingBad1x01"},
		{"Breaking.Bad.S01E01.1080p.BluRay.x265-RelGroup.mkv", "BreakingBadS01E01"},
		{"Star Trek Deep Space Nine - S01E01-E02 - Emissary.mkv", "StarTrekDeepSpaceNineS01E01E02"},
		{"Star Trek Deep Space Nine - 1x01-02 - Emissary.mkv", "StarTrekDeepSpaceNine1x0102"},
		{"Breaking.Bad.S01E01.1080p.BluRay.x265-RelGroup.mp3", ""}, // Unsupported file extension returns empty key
	}

	for _, tc := range tcs {
		t.Run(tc.path, func(t *testing.T) {
			key, err := movieTvFileNamesKeyGenerator(tc.path)
			if err != nil {
				t.Error(err)
			}

			if key != tc.expected {
				t.Errorf("Expected key to be '%s', got '%s'", tc.expected, key)
			}
		})
	}
}
