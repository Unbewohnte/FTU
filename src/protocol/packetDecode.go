// Methods to decode read from connection packets defined in protocol
package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/Unbewohnte/ftu/fsys"
)

var ErrorWrongPacket error = fmt.Errorf("wrong type of packet header")

// decodes packet with the header FILE into the fsys.File struct
func DecodeFilePacket(filePacket *Packet) (*fsys.File, error) {
	if filePacket.Header != HeaderFile {
		return nil, ErrorWrongPacket
	}
	// FILE~(idInBinary)(filenameLengthInBinary)(filename)(filesize)(checksumLengthInBinary)checksum

	// retrieve data from packet body

	// id
	packetReader := bytes.NewBuffer(filePacket.Body)

	var fileID uint64
	err := binary.Read(packetReader, binary.BigEndian, &fileID)
	if err != nil {
		return nil, err
	}

	// filename
	var filenameLength uint64
	err = binary.Read(packetReader, binary.BigEndian, &filenameLength)
	if err != nil {
		return nil, err
	}

	filenameBytes := make([]byte, filenameLength)
	_, err = packetReader.Read(filenameBytes)
	if err != nil {
		return nil, err
	}

	filename := string(filenameBytes)

	// filesize
	var filesize uint64
	err = binary.Read(packetReader, binary.BigEndian, &filesize)
	if err != nil {
		return nil, err
	}

	// checksum
	var checksumLength uint64
	err = binary.Read(packetReader, binary.BigEndian, &checksumLength)
	if err != nil {
		return nil, err
	}
	checksumBytes := make([]byte, checksumLength)
	_, err = packetReader.Read(checksumBytes)
	if err != nil {
		return nil, err
	}
	checksum := string(checksumBytes)

	return &fsys.File{
		ID:       fileID,
		Name:     filename,
		Size:     filesize,
		Checksum: checksum,
		Handler:  nil,
	}, nil
}

// decodes DIRECTORY packet into fsys.Directory struct
func DecodeDirectoryPacket(dirPacket *Packet) (*fsys.Directory, error) {
	if dirPacket.Header != HeaderDirectory {
		return nil, ErrorWrongPacket
	}

	// DIRECTORY~(dirname size in binary)(dirname)(dirsize)

	packetReader := bytes.NewReader(dirPacket.Body)

	// name
	var dirNameSize uint64
	err := binary.Read(packetReader, binary.BigEndian, &dirNameSize)
	if err != nil {
		return nil, err
	}
	dirName := make([]byte, dirNameSize)
	_, err = packetReader.Read(dirName)
	if err != nil {
		return nil, err
	}

	// size
	var dirSize uint64
	err = binary.Read(packetReader, binary.BigEndian, &dirSize)
	if err != nil {
		return nil, err
	}

	dir := fsys.Directory{
		Name: string(dirName),
		Size: dirSize,
	}

	return &dir, nil
}

// decodes TRANSFERINFO packet into either fsys.File or fsys.Directory struct.
// decodeTransferPacket cannot return 2 nils or both non-nils as 2 first return values in case
// of a successfull decoding
func DecodeTransferPacket(transferPacket *Packet) (*fsys.File, *fsys.Directory, error) {
	if transferPacket.Header != HeaderTransferOffer {
		return nil, nil, ErrorWrongPacket
	}

	var file *fsys.File = nil
	var dir *fsys.Directory = nil
	var err error

	// determine if it`s a file or a directory
	switch string(transferPacket.Body[0]) {
	case FILECODE:
		filePacket := Packet{
			Header: HeaderFile,
			Body:   transferPacket.Body[1:],
		}

		file, err = DecodeFilePacket(&filePacket)
		if err != nil {
			return nil, nil, err
		}

	case DIRCODE:
		dirPacket := Packet{
			Header: HeaderDirectory,
			Body:   transferPacket.Body[1:],
		}

		dir, err = DecodeDirectoryPacket(&dirPacket)
		if err != nil {
			return nil, nil, err
		}

	default:
		return nil, nil, ErrorInvalidPacket
	}

	return file, dir, nil
}
