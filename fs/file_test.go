package fs

import "testing"

func Test_GetFile(t *testing.T) {
	filepath := "../testfiles/testfile.txt"

	file, err := GetFile(filepath)
	if err != nil {
		t.Fatalf("GetFile error: %s", err)
	}

	if file.Name != "testfile.txt" {
		t.Fatalf("GetFile error: filenames do not match")
	}
}
