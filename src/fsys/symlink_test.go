package fsys

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_IsSymlink(t *testing.T) {
	dirpath := "../testfiles/"

	symlinkPath := filepath.Join(dirpath, "testsymlink.txt")
	os.Symlink(filepath.Join(dirpath, "testfile.txt"), symlinkPath)

	isSymlink, err := IsSymlink(symlinkPath)
	if err != nil {
		t.Fatalf("%s\n", err)
	}
	if !isSymlink {
		t.Fatalf("%s expected to be a symlink\n", symlinkPath)
	}
}
