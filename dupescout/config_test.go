package dupescout

import (
	"os"
	"reflect"
	"runtime"
	"testing"
)

func TestSanitizePath(t *testing.T) {
	cfg := &Cfg{Paths: []string{"~/Dev", "~/Dev/../dupescout"}}
	path := sanitizePath(cfg.Paths[0])

	home, _ := os.UserHomeDir()
	isWindows := os.PathSeparator == '\\'

	if isWindows && path != home+"\\Dev" {
		t.Errorf("Expected %s, got %s", home+"\\Dev", path)
	}

	if !isWindows && path != home+"/Dev" {
		t.Errorf("Expected %s, got %s", home+"/Dev", path)
	}

	path = sanitizePath(cfg.Paths[1]) // Use second path now

	if isWindows && path != home+"\\dupescout" {
		t.Errorf("Expected %s, got %s", home+"\\dupescout", path)
	}

	if !isWindows && path != home+"/dupescout" {
		t.Errorf("Expected %s, got %s", home+"/dupescout", path)
	}
}

func TestDefaults(t *testing.T) {
	cfg := &Cfg{}
	cfg.defaults()

	if reflect.ValueOf(cfg.KeyGenerator).Pointer() != reflect.ValueOf(Crc32HashKeyGenerator).Pointer() {
		t.Error("Expected key generator to be set to default: Crc32HashKeyGenerator")
	}

	defaultWorkers := runtime.GOMAXPROCS(0) / 2
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
