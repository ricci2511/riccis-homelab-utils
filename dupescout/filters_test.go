package dupescout

import (
	"reflect"
	"testing"
)

func TestFiltersListSet(t *testing.T) {
	var fl FiltersList

	fl.Set("test1,test2")
	if !reflect.DeepEqual(fl, FiltersList{"test1", "test2"}) {
		t.Errorf("Expected FiltersList to be [test1 test2], got %v", fl)
	}

	fl.Set("test3 test4")
	if !reflect.DeepEqual(fl, FiltersList{"test1", "test2", "test3", "test4"}) {
		t.Errorf("Expected FiltersList to be [test1 test2 test3 test4], got %v", fl)
	}
}

func TestSkipFile(t *testing.T) {
	f := Filters{
		ExtExclude: []string{".jpg", ".png"},
	}

	if !f.skipFile("test.jpg") || !f.skipFile("test.png") {
		t.Error("Expected true, got false")
	}

	// Hidden files must be skipped by default
	if !f.skipFile(".vimrc") {
		t.Error("Expected true, got false")
	}

	f.HiddenInclude = true

	if f.skipFile(".vimrc") {
		t.Error("Expected false, got true")
	}

	f.ExtInclude = []string{".txt", ".docx"}

	if f.skipFile("test.txt") || f.skipFile("test.docx") {
		t.Error("Expected false, got true")
	}

	// Even though ".js" files are not explicitly excluded, they will be
	// skipped if include filters are set and the file extension is not
	// in the ExtInclude slice.
	if !f.skipFile("test.js") {
		t.Error("Expected true, got false")
	}
}

func TestSkipDir(t *testing.T) {
	f := Filters{
		DirsExclude: []string{"node_modules"},
	}

	if !f.skipDir("node_modules") {
		t.Error("Expected true, got false")
	}

	if f.skipDir("test") {
		t.Error("Expected false, got true")
	}

	// Hidden directories must be skipped by default
	if !f.skipDir(".git") {
		t.Error("Expected true, got false")
	}

	f.HiddenInclude = true

	if f.skipDir(".git") {
		t.Error("Expected false, got true")
	}

	f.SkipSubdirs = true

	// Even though "test" is not explicitly excluded, it will be
	// skipped if SkipSubdirs is true, because it prevents the
	// recursive traversal of any subdirectories.
	if !f.skipDir("test") {
		t.Error("Expected true, got false")
	}
}
