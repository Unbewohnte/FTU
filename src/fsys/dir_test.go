package fsys

import "testing"

func Test_GetDir(t *testing.T) {
	dirpath := "../testfiles/"

	_, err := GetDir(dirpath, false)
	if err != nil {
		t.Fatalf("GetDir error: %s", err)
	}
}

func Test_GetDirRecursive(t *testing.T) {
	dirpath := "../testfiles/"

	dir, err := GetDir(dirpath, true)
	if err != nil {
		t.Fatalf("GetDir error: %s", err)
	}

	expectedAmountOfUpperDirectories := 3
	if len(dir.Directories) != expectedAmountOfUpperDirectories {
		t.Fatalf("GetDir error: expected to have %d inner directories; got %d", expectedAmountOfUpperDirectories, len(dir.Directories))
	}
}
