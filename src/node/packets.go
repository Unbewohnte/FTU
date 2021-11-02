// node-specific packets and packet handling
package node

import (
	"bytes"
	"encoding/binary"
	"net"

	"github.com/Unbewohnte/ftu/fsys"
	"github.com/Unbewohnte/ftu/protocol"
)

// Reads packets from connection in an endless loop, sends them to the channel
func receivePackets(connection net.Conn, packetPipe chan *protocol.Packet) error {
	for {
		if connection == nil {
			return ErrorNotConnected
		}

		packetBytes, err := protocol.ReadFromConn(connection)
		if err != nil {
			close(packetPipe)
			return err
		}

		incomingPacket, err := protocol.BytesToPacket(packetBytes)
		if err != nil {
			close(packetPipe)
			return err
		}

		packetPipe <- incomingPacket
	}
}

// decodes packet with the header FILE into the fsys.File struct
func decodeFilePacket(filePacket *protocol.Packet) (*fsys.File, error) {
	// FILE~(idInBinary)(filenameLengthInBinary)(filename)(filesize)(checksumLengthInBinary)checksum

	// retrieve data from packet body

	// id
	packetReader := bytes.NewBuffer(filePacket.Body)

	var fileID uint64
	err := binary.Read(packetReader, binary.BigEndian, &fileID)
	if err != nil {
		panic(err)
	}

	// filename
	var filenameLength uint64
	err = binary.Read(packetReader, binary.BigEndian, &filenameLength)
	if err != nil {
		panic(err)
	}

	filenameBytes := make([]byte, filenameLength)
	_, err = packetReader.Read(filenameBytes)
	if err != nil {
		panic(err)
	}

	filename := string(filenameBytes)

	// filesize
	var filesize uint64
	err = binary.Read(packetReader, binary.BigEndian, &filesize)
	if err != nil {
		panic(err)
	}

	// checksum
	var checksumLength uint64
	err = binary.Read(packetReader, binary.BigEndian, &checksumLength)
	if err != nil {
		panic(err)
	}
	checksumBytes := make([]byte, checksumLength)
	_, err = packetReader.Read(checksumBytes)
	if err != nil {
		panic(err)
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
