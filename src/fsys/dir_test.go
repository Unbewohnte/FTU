package fsys

import (
	"testing"
)

func Test_GetDir(t *testing.T) {
	dirpath := "../testfiles/"

	_, err := GetDir(dirpath, false)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func Test_GetDirRecursive(t *testing.T) {
	dirpath := "../testfiles/"

	dir, err := GetDir(dirpath, true)
	if err != nil {
		t.Fatalf("%s", err)
	}

	expectedAmountOfUpperDirectories := 3
	if len(dir.Directories) != expectedAmountOfUpperDirectories {
		t.Fatalf("expected to have %d inner directories; got %d", expectedAmountOfUpperDirectories, len(dir.Directories))
	}

	for _, innerDir := range dir.Directories {
		if innerDir.Size > dir.Size {
			t.Fatalf("inner dir cannot have a bigger size (%d B) than its parent`s total size (%d B)", innerDir.Size, dir.Size)
		}
	}

}

func Test_GetFiles(t *testing.T) {
	dirpath := "../testfiles/"

	dir, err := GetDir(dirpath, true)
	if err != nil {
		t.Fatalf("%s", err)
	}

	// recursive
	files := dir.GetAllFiles(true)

	fileCount := 6
	if len(files) != fileCount {
		t.Fatalf("expected to get %d files; got %d\n", fileCount, len(files))
	}

	// not recursive
	files = dir.GetAllFiles(false)
	fileCount = 1
	if len(files) != fileCount {
		t.Fatalf("expected to get %d files; got %d\n", fileCount, len(files))
	}

}

// func Test_SetRelativePaths(t *testing.T) {
// 	dirpath := "../testfiles/"

// 	dir, err := GetDir(dirpath, true)
// 	if err != nil {
// 		t.Fatalf("%s", err)
// 	}

// 	absDirPath, err := filepath.Abs(dirpath)
// 	if err != nil {
// 		t.Fatalf("%s", err)
// 	}

// 	err = dir.SetRelativePaths(absDirPath, true)
// 	if err != nil {
// 		t.Fatalf("%s", err)
// 	}

// 	for count, file := range dir.GetAllFiles(true) {
// 		t.Errorf("[%d] %v\n", count, file.RelativeParentPath)
// 	}
// }
