// Methonds to send packets in various ways defined in protocol
package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/Unbewohnte/ftu/encryption"
	"github.com/Unbewohnte/ftu/fsys"
)

// Sends given packet to connection.
// ALL packets MUST be sent by this method
func SendPacket(connection net.Conn, packet Packet) error {
	packetBytes, err := packet.ToBytes()
	if err != nil {
		return err
	}

	fmt.Printf("[SEND] packet %+s; len: %d\n", packetBytes[:30], len(packetBytes))

	// write the result (ie: (packetsize)(header)~(bodybytes))
	connection.Write(packetBytes)

	return nil
}

// sends an encryption key to the other side
func SendEncryptionKey(connection net.Conn, encrKey []byte) error {
	encrKeyPacketBuffer := new(bytes.Buffer)

	encrKeyLength := uint64(len(encrKey))

	err := binary.Write(encrKeyPacketBuffer, binary.BigEndian, &encrKeyLength)
	if err != nil {
		return err
	}

	encrKeyPacketBuffer.Write(encrKey)

	err = SendPacket(connection, Packet{
		Header: HeaderEncryptionKey,
		Body:   encrKeyPacketBuffer.Bytes(),
	})
	if err != nil {
		return err
	}

	return nil
}

// sends a TRANSFEROFFER packet to connection with information about either file or directory.
// If file is the only thing that the sender is going to send - leave dir arg as nil, the same
// applies if directory is the only thing that the sender is going to send - leave file as nil.
// sendTransferOffer PANICS if both file and dir are present or nil. If encrKey != nil - encrypts
// constructed packet
func SendTransferOffer(connection net.Conn, file *fsys.File, dir *fsys.Directory, encrKey []byte) error {
	if file == nil && dir == nil {
		panic("either file or dir must be specified")
	} else if file != nil && dir != nil {
		panic("only one either file or dir must be specified")
	}

	transferOfferPacket := Packet{
		Header: HeaderTransferOffer,
	}

	if file != nil {
		filePacket, err := CreateFilePacket(file)
		if err != nil {
			return err
		}

		transferOfferBody := append([]byte(FILECODE), filePacket.Body...)

		// if encrKey is present - encrypt
		if encrKey != nil {
			encryptedBody, err := encryption.Encrypt(encrKey, transferOfferBody)
			if err != nil {
				return err
			}
			transferOfferBody = encryptedBody
		}

		transferOfferPacket.Body = transferOfferBody

	} else if dir != nil {
		dirPacket, err := CreateDirectoryPacket(dir)
		if err != nil {
			return err
		}

		transferOfferBody := append([]byte(DIRCODE), dirPacket.Body...)
		// if encrKey is present - encrypt
		if encrKey != nil {
			encryptedBody, err := encryption.Encrypt(encrKey, transferOfferBody)
			if err != nil {
				return err
			}
			transferOfferBody = encryptedBody
		}

		transferOfferPacket.Body = transferOfferBody
	}

	// send packet
	err := SendPacket(connection, transferOfferPacket)
	if err != nil {
		return err
	}

	return nil
}

var ErrorSentAll error = fmt.Errorf("sent the whole file")

// sends a piece of file to the connection; The next calls will send
// another piece util the file has been fully sent. If encrKey is not nil - encrypts each packet with
// this key
func SendPiece(file *fsys.File, connection net.Conn, encrKey []byte) error {
	err := file.Open()
	if err != nil {
		return err
	}
	defer file.Handler.Close()

	if file.SentBytes == 0 {
		file.Handler.Seek(0, io.SeekStart)
	}

	if file.Size == file.SentBytes {
		return ErrorSentAll
	}

	fileBytesPacket := Packet{
		Header: HeaderFileBytes,
	}

	packetBodyBuff := new(bytes.Buffer)

	// write file ID first
	err = binary.Write(packetBodyBuff, binary.BigEndian, &file.ID)
	if err != nil {
		return err
	}

	// fill the remaining space of packet with the contents of a file
	canSendBytes := uint64(MAXPACKETSIZE) - fileBytesPacket.Size() - uint64(packetBodyBuff.Len())

	if encrKey != nil {
		// account for padding
		canSendBytes -= 32
	}

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

	if encrKey != nil {
		err = fileBytesPacket.EncryptBody(encrKey)
		if err != nil {
			return err
		}
	}

	// send it to the other side
	err = SendPacket(connection, fileBytesPacket)
	if err != nil {
		return err
	}

	file.SentBytes += uint64(read)

	return nil
}