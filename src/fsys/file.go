package fsys

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Unbewohnte/ftu/checksum"
)

// A struct that represents the necessary file information for transportation through node
type File struct {
	ID         uint64 // Set manually
	Name       string
	Path       string
	ParentPath string
	Size       uint64
	Checksum   string
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
		Name:       stats.Name(),
		Path:       absPath,
		ParentPath: filepath.Dir(absPath),
		Size:       uint64(stats.Size()),
		Handler:    nil,
	}

	// get checksum
	err = file.Open()
	if err != nil {
		return nil, err
	}
	defer file.Handler.Close()

	checksum, err := checksum.GetPartialCheckSum(file.Handler)
	if err != nil {
		return nil, err
	}

	file.Checksum = checksum

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
