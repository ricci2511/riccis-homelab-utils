package dupescout

import (
	"os"
	"testing"
)

func TestSanitizePath(t *testing.T) {
	cfg := &Cfg{Path: "~/Dev"}
	cfg.sanitizePath()

	home, _ := os.UserHomeDir()
	isWindows := os.PathSeparator == '\\'

	if isWindows && cfg.Path != home+"\\Dev" {
		t.Errorf("Expected %s, got %s", home+"\\Dev", cfg.Path)
	}

	if !isWindows && cfg.Path != home+"/Dev" {
		t.Errorf("Expected %s, got %s", home+"/Dev", cfg.Path)
	}

	cfg = &Cfg{Path: "~/Dev/../dupescout"}
	cfg.sanitizePath()

	if isWindows && cfg.Path != home+"\\dupescout" {
		t.Errorf("Expected %s, got %s", home+"\\dupescout", cfg.Path)
	}

	if !isWindows && cfg.Path != home+"/dupescout" {
		t.Errorf("Expected %s, got %s", home+"/dupescout", cfg.Path)
	}
}

func TestDefaults(t *testing.T) {
	cfg := &Cfg{}
	cfg.defaults()

	if cfg.KeyGenerator == nil {
		t.Errorf("Expected key generator to be set")
	}

	if cfg.Workers == 0 {
		t.Errorf("Expected workers to be set")
	}

	cfg = &Cfg{Workers: 5}
	cfg.defaults()

	if cfg.Workers != 5 {
		t.Errorf("Expected workers to be 5")
	}
}
