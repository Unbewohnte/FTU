package fsys

import (
	"fmt"
	"os"
	"path/filepath"
)

var FileIDsCounter uint64 = 1

// A struct that represents the necessary file information for transportation through node
type File struct {
	ID         uint64
	Name       string
	Path       string
	ParentPath string
	Size       uint64
	Checksum   string   // Set manually
	Handler    *os.File // Set when .Open() is called
	SentBytes  uint64   // Set manually during transportation
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
		ID:         FileIDsCounter,
		Name:       stats.Name(),
		Path:       absPath,
		ParentPath: filepath.Dir(absPath),
		Size:       uint64(stats.Size()),
		Handler:    nil,
	}

	// increment ids counter so the next file will have a different ID
	FileIDsCounter++

	return &file, nil
}

// Opens file for read/write operations
func (file *File) Open() error {
	handler, err := os.OpenFile(file.Path, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	file.Handler = handler

	return nil
}
