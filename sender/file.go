package sender

import (
	"fmt"
	"os"

	"github.com/Unbewohnte/FTU/checksum"
)

// Struct that represents the served file. Used internally in the sender
type File struct {
	path        string
	Filename    string
	Filesize    uint64
	SentBytes   uint64
	LeftBytes   uint64
	SentPackets uint64
	Handler     *os.File
	CheckSum    checksum.CheckSum
}

// Prepares a file for serving. Used for preparing info before sending a fileinfo packet by sender
func getFile(path string) (*File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("could not get a fileinfo: %s", err)
	}
	handler, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("wasn`t able to open the file: %s", err)
	}
	checksum, err := checksum.GetPartialCheckSum(handler)
	if err != nil {
		return nil, fmt.Errorf("could not get a partial file checksum: %s", err)
	}

	return &File{
		path:      path,
		Filename:  info.Name(),
		Filesize:  uint64(info.Size()),
		SentBytes: 0,
		LeftBytes: uint64(info.Size()),
		Handler:   handler,
		CheckSum:  checksum,
	}, nil
}
