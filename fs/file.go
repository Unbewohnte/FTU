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
	Handler    *os.File
}

var ErrorNotFile error = fmt.Errorf("not a file")

// Get general information about a file with the
// future ability to open it.
// NOTE that Handler field is nil BY DEFAULT until you
// manually call a (file *File) Open() function to open it !
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
		Handler:    nil,
	}

	return &file, nil
}

// Opens file for read/write operations
func (file *File) Open() error {
	handler, err := os.OpenFile(file.Path, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	file.Handler = handler

	return nil
}
