package server

import (
	"fmt"
	"os"
)

// Struct that represents the served file. Used internally in the server
type File struct {
	path        string
	Filename    string
	Filesize    uint64
	SentBytes   uint64
	LeftBytes   uint64
	SentPackets uint64
	Handler     *os.File
}

// Prepares a file for serving. Used for preparing info before sending a handshake
func getFile(path string) (*File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("could not get a fileinfo: %s", err)
	}
	handler, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("couldn`t be able to open the file: %s", err)
	}
	return &File{
		path:      path,
		Filename:  info.Name(),
		Filesize:  uint64(info.Size()),
		SentBytes: 0,
		LeftBytes: uint64(info.Size()),
		Handler:   handler,
	}, nil
}
