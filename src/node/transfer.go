package node

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"os"

	"github.com/Unbewohnte/ftu/checksum"
	"github.com/Unbewohnte/ftu/fsys"
	"github.com/Unbewohnte/ftu/protocol"
)

// sends a notification about the file
func sendFilePacket(connection net.Conn, file *fsys.File) error {
	if connection == nil {
		return ErrorNotConnected
	}

	err := file.Open()
	if err != nil {
		return err
	}

	// FILE~(idInBinary)(filenameLengthInBinary)(filename)(filesize)(checksumLengthInBinary)checksum

	// send file packet with file description
	filePacket := protocol.Packet{
		Header: protocol.HeaderFile,
	}
	fPacketBodyBuff := new(bytes.Buffer)

	// file id
	binary.Write(fPacketBodyBuff, binary.BigEndian, &file.ID)

	// filename
	filenameLen := uint64(len([]byte(file.Name)))
	binary.Write(fPacketBodyBuff, binary.BigEndian, &filenameLen)
	fPacketBodyBuff.Write([]byte(file.Name))

	// size
	binary.Write(fPacketBodyBuff, binary.BigEndian, &file.Size)

	// checksum
	fileChecksum, err := checksum.GetPartialCheckSum(file.Handler)
	if err != nil {
		return err
	}

	checksumLen := uint64(len([]byte(fileChecksum)))
	binary.Write(fPacketBodyBuff, binary.BigEndian, &checksumLen)
	fPacketBodyBuff.Write([]byte(fileChecksum))

	filePacket.Body = fPacketBodyBuff.Bytes()

	err = protocol.SendPacket(connection, filePacket)
	if err != nil {
		return err
	}

	return nil
}

// sends a notification about the directory
func sendDirectoryPacket(connection net.Conn, dir *fsys.Directory) error {
	if connection == nil {
		return ErrorNotConnected
	}

	return nil
}

// sends a piece of file to the connection; The next calls will send
// another piece util the file has been fully sent
func sendPiece(file *fsys.File, connection net.Conn) error {
	if file.Handler == nil {
		fHandler, err := os.Open(file.Path)
		if err != nil {
			return err
		}

		file.Handler = fHandler
	}

	if file.SentBytes == 0 {
		file.Handler.Seek(0, io.SeekStart)
	}

	if file.Size == file.SentBytes {
		return ErrorSentAll
	}

	fileBytesPacket := protocol.Packet{
		Header: protocol.HeaderFileBytes,
	}

	packetBodyBuff := new(bytes.Buffer)

	// write file ID first
	err := binary.Write(packetBodyBuff, binary.BigEndian, &file.ID)
	if err != nil {
		return err
	}

	// fill the remaining space of packet with the contents of a file
	canSendBytes := uint64(protocol.MAXPACKETSIZE) - fileBytesPacket.Size() - uint64(packetBodyBuff.Len())

	if (file.Size - file.SentBytes) < canSendBytes {
		canSendBytes = (file.Size - file.SentBytes)
	}
	fileBytes := make([]byte, canSendBytes)

	read, err := file.Handler.ReadAt(fileBytes, int64(file.SentBytes))
	if err != nil {
		return err
	}

	packetBodyBuff.Write(fileBytes)

	fileBytesPacket.Body = packetBodyBuff.Bytes()

	// send it to the other side
	err = protocol.SendPacket(connection, fileBytesPacket)
	if err != nil {
		return err
	}

	file.SentBytes += uint64(read)

	return nil
}
