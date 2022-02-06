/*
ftu - file transferring utility.
Copyright (C) 2021,2022  Kasyanov Nikolay Alexeyevich (Unbewohnte (https://unbewohnte.xyz/))

This file is a part of ftu

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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

	//(id in binary)(filename length in binary)(filename)(filesize)(checksum length in binary)(checksum)(relative path to the upper directory size in binary if present)(relative path)

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

	// relative path
	var relPathLength uint64
	err = binary.Read(packetReader, binary.BigEndian, &relPathLength)
	if err != nil {
		return nil, err
	}
	relPathBytes := make([]byte, relPathLength)
	_, err = packetReader.Read(relPathBytes)
	if err != nil {
		return nil, err
	}
	relPath := string(relPathBytes)

	return &fsys.File{
		ID:                 fileID,
		Name:               filename,
		Size:               filesize,
		Checksum:           checksum,
		RelativeParentPath: relPath,
		Handler:            nil,
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
