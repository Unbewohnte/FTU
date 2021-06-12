package receiver

import (
	"fmt"
	"os"

	"github.com/Unbewohnte/FTU/checksum"
)

// Receiver`s file struct. Used internally by receiver
type File struct {
	Filename string
	Filesize uint64
	CheckSum checksum.CheckSum
}

// Goes through all files in the downloads directory and compares their
// names with the name of the file that is about to be downloaded
func (r *Receiver) CheckIfFileAlreadyExists() (bool, error) {
	contents, err := os.ReadDir(r.DownloadsFolder)
	if err != nil {
		return false, fmt.Errorf("could not get contents of the downloads` directory: %s", err)
	}
	for _, file := range contents {
		if file.Name() == r.FileToDownload.Filename {
			return true, nil
		}
	}
	return false, nil
}
