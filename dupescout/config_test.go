package dupescout

import (
	"os"
	"reflect"
	"runtime"
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

	if reflect.ValueOf(cfg.KeyGenerator).Pointer() != reflect.ValueOf(Crc32HashKeyGenerator).Pointer() {
		t.Error("Expected key generator to be set to default: Crc32HashKeyGenerator")
	}

	defaultWorkers := runtime.GOMAXPROCS(0)
	if cfg.Workers != defaultWorkers {
		t.Errorf("Expected workers to be set to default: %d", defaultWorkers)
	}

	cfg = &Cfg{Workers: 5}
	cfg.defaults()

	if cfg.Workers != 5 {
		t.Errorf("Expected workers to be 5")
	}

	cfg.KeyGenerator = Sha256HashKeyGenerator

	if reflect.ValueOf(cfg.KeyGenerator).Pointer() != reflect.ValueOf(Sha256HashKeyGenerator).Pointer() {
		t.Errorf("Expected key generator to be set to Sha256HashKeyGenerator")
	}
}
