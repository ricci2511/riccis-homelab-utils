package dupescout

import (
	"fmt"
	"os"
	"testing"
)

// Helper to create a temp file with the given string content and a function to remove it when done.
func createTempFile(content string) (os.File, func()) {
	f, err := os.CreateTemp("", "dupescout")
	if err != nil {
		fmt.Println(err)
	}

	f.WriteString(content)

	clean := func() {
		f.Close()
		os.Remove(f.Name())
	}

	return *f, clean
}

// Helper to test the equality of two files with the same content using the given KeyGeneratorFunc.
func hashKeyGeneratorEquality(content string, keyGenFunc KeyGeneratorFunc) (bool, error) {
	file1, clean := createTempFile(content)
	defer clean()

	key1, err := keyGenFunc(file1.Name())
	if err != nil {
		return false, err
	}

	file2, clean := createTempFile(content)
	defer clean()

	key2, err := keyGenFunc(file2.Name())
	if err != nil {
		return false, err
	}

	return key1 == key2, nil
}

// Helper to test the inequality of two files with different content using the given KeyGeneratorFunc.
func hashKeyGeneratorInequality(content1 string, content2 string, keyGenFunc KeyGeneratorFunc) (bool, error) {
	file1, clean := createTempFile(content1)
	defer clean()

	key1, err := keyGenFunc(file1.Name())
	if err != nil {
		return false, err
	}

	file2, clean := createTempFile(content2)
	defer clean()

	key2, err := keyGenFunc(file2.Name())
	if err != nil {
		return false, err
	}

	return key1 != key2, nil
}

func TestCrc32HashKeyGeneratorEquality(t *testing.T) {
	content := "Hello, World!"

	equal, err := hashKeyGeneratorEquality(content, Crc32HashKeyGenerator)
	if err != nil {
		t.Error(err)
	}

	if !equal {
		t.Errorf("Expected %s to equal %s", content, content)
	}
}

func TestCrc32HashKeyGeneratorInequality(t *testing.T) {
	content1 := "Go rocks!"
	content2 := "JavaScript rocks!"

	inequal, err := hashKeyGeneratorInequality(content1, content2, Crc32HashKeyGenerator)
	if err != nil {
		t.Error(err)
	}

	if !inequal {
		t.Errorf("Expected %s to not equal %s", content1, content2)
	}
}

func TestSha256HashKeyGeneratorEquality(t *testing.T) {
	content := "Hello, World!"

	equal, err := hashKeyGeneratorEquality(content, Sha256HashKeyGenerator)
	if err != nil {
		t.Error(err)
	}

	if !equal {
		t.Errorf("Expected %s to equal %s", content, content)
	}
}

func TestSha256HashKeyGeneratorInequality(t *testing.T) {
	content1 := "Go rocks!"
	content2 := "JavaScript rocks!"

	inequal, err := hashKeyGeneratorInequality(content1, content2, Sha256HashKeyGenerator)
	if err != nil {
		t.Error(err)
	}

	if !inequal {
		t.Errorf("Expected %s to not equal %s", content1, content2)
	}
}
