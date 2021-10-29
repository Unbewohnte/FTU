package fs

import (
	"io"
	"testing"
)

func Test_GetFile(t *testing.T) {
	filepath := "../testfiles/testfile.txt"

	file, err := GetFile(filepath)
	if err != nil {
		t.Fatalf("GetFile error: %s", err)
	}

	expectedFilename := "testfile.txt"
	if file.Name != expectedFilename {
		t.Fatalf("GetFile error: filenames do not match: expected filename to be %s; got %s", expectedFilename, file.Name)
	}
}

func Test_GetFileOpen(t *testing.T) {
	filepath := "../testfiles/testfile.txt"

	file, err := GetFile(filepath)
	if err != nil {
		t.Fatalf("GetFile error: %s", err)
	}

	err = file.Open()
	if err != nil {
		t.Fatalf("GetFile error: could not open file: %s", err)
	}

	_, err = io.ReadAll(file.Handler)
	if err != nil {
		t.Fatalf("GetFile error: could not read from file: %s", err)
	}
}
