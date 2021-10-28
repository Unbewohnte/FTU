package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

// A struct that represents the main file information
type File struct {
	Name       string
	Path       string
	ParentPath string
	Size       uint64
}

var ErrorNotFile error = fmt.Errorf("not a file")

func GetFile(path string) (*File, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	stats, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	// check if it is a directory
	if stats.IsDir() {
		return nil, ErrorNotFile
	}

	file := File{
		Name:       stats.Name(),
		Path:       absPath,
		ParentPath: filepath.Dir(absPath),
		Size:       uint64(stats.Size()),
	}

	return &file, nil
}
