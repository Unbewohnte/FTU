package fs

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

	expectedAmountOfUpperDirectories := 2
	if len(dir.Directories) != expectedAmountOfUpperDirectories {
		t.Fatalf("GetDir error: expected to have %d inner directories; got %d", expectedAmountOfUpperDirectories, len(dir.Directories))
	}

	innerDir1 := dir.Directories[0]
	if innerDir1 == nil || innerDir1.Name != "testdir" {
		t.Fatalf("GetDir error: expected to have the first inner directory to be \"%s\"; got \"%s\"", "testdir", innerDir1.Name)
	}
}
