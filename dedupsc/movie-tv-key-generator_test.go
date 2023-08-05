package main

import "testing"

func TestMovieFileName(t *testing.T) {
	path := "Alien - 1979 - Bluray-1080p.mkv"

	movieKey, err := movieTvFileNamesKeyGenerator(path)
	if err != nil {
		t.Error(err)
	}

	if movieKey != "Alien1979" {
		t.Errorf("Expected key to be 'Alien1979', got '%s'", movieKey)
	}
}

func TestMovieFileNameReleaseGroup(t *testing.T) {
	path := "Alien.1979.1080p.BluRay.x265-RelGroup.mkv"

	movieKey, err := movieTvFileNamesKeyGenerator(path)
	if err != nil {
		t.Error(err)
	}

	if movieKey != "Alien1979" {
		t.Errorf("Expected key to be 'Alien1979', got '%s'", movieKey)
	}
}

func TestTvFileName(t *testing.T) {
	path := "Breaking Bad - S01E01 - Pilot.mkv"

	tvKey, err := movieTvFileNamesKeyGenerator(path)
	if err != nil {
		t.Error(err)
	}

	if tvKey != "BreakingBadS01E01" {
		t.Errorf("Expected key to be 'BreakingBadS01E01', got '%s'", tvKey)
	}
}

func TestTvFileNameReleaseGroup(t *testing.T) {
	path := "Breaking.Bad.S01E01.1080p.BluRay.x265-RelGroup.mkv"

	tvKey, err := movieTvFileNamesKeyGenerator(path)
	if err != nil {
		t.Error(err)
	}

	if tvKey != "BreakingBadS01E01" {
		t.Errorf("Expected key to be 'BreakingBadS01E01', got '%s'", tvKey)
	}
}

func TestTvMultiEpisodeFileName(t *testing.T) {
	path := "Star Trek Deep Space Nine - S01E01-E02 - Emissary.mkv"

	tvKey, err := movieTvFileNamesKeyGenerator(path)
	if err != nil {
		t.Error(err)
	}

	if tvKey != "StarTrekDeepSpaceNineS01E01E02" {
		t.Errorf("Expected key to be 'StarTrekDeepSpaceNineS01E01E02', got '%s'", tvKey)
	}
}

func TestAlternateTvFileNamePattern(t *testing.T) {
	path := "Breaking Bad - 1x01 - Pilot.mkv"

	tvKey, err := movieTvFileNamesKeyGenerator(path)
	if err != nil {
		t.Error(err)
	}

	if tvKey != "BreakingBad1x01" {
		t.Errorf("Expected key to be 'BreakingBad1x01', got '%s'", tvKey)
	}
}

func TestInvalidFileExtension(t *testing.T) {
	path := "Breaking.Bad.S01E01.1080p.BluRay.x265-RelGroup.mp3"

	tvKey, err := movieTvFileNamesKeyGenerator(path)
	if err != nil {
		t.Error(err)
	}

	if tvKey != "" {
		t.Errorf("Expected key to be '', got '%s'", tvKey)
	}
}
