/*
ftu - file transferring utility.
Copyright (C) 2021,2022  Kasyanov Nikolay Alexeevich (Unbewohnte (https://unbewohnte.xyz/))

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

	// fmt.Printf("[SEND] packet %+s; len: %d\n", packetBytes[:30], len(packetBytes))

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

// Sends a piece of file to the connection; The next calls will send
// another piece util the file has been fully sent. If encrKey is not nil - encrypts each packet with
// this key. Returns amount of filebytes written to the connection
func SendPiece(file *fsys.File, connection net.Conn, encrKey []byte) (uint64, error) {
	var sentBytes uint64 = 0

	err := file.Open()
	if err != nil {
		return sentBytes, err
	}
	defer file.Close()

	if file.SentBytes == 0 {
		file.Handler.Seek(0, io.SeekStart)
	}

	if file.Size == file.SentBytes {
		return sentBytes, ErrorSentAll
	}

	fileBytesPacket := Packet{
		Header: HeaderFileBytes,
	}

	packetBodyBuff := new(bytes.Buffer)

	// write file ID first
	err = binary.Write(packetBodyBuff, binary.BigEndian, file.ID)
	if err != nil {
		return sentBytes, err
	}

	// fill the remaining space of packet with the contents of a file
	canSendBytes := uint64(MAXPACKETSIZE) - fileBytesPacket.Size() - uint64(packetBodyBuff.Len())

	if encrKey != nil {
		// account for padding
		canSendBytes -= 48
	}

	if (file.Size - file.SentBytes) < canSendBytes {
		canSendBytes = (file.Size - file.SentBytes)
	}

	fileBytes := make([]byte, canSendBytes)

	read, err := file.Handler.ReadAt(fileBytes, int64(file.SentBytes))
	if err != nil {
		return sentBytes, err
	}
	file.SentBytes += uint64(read)
	sentBytes += uint64(canSendBytes)

	packetBodyBuff.Write(fileBytes)

	fileBytesPacket.Body = packetBodyBuff.Bytes()

	if encrKey != nil {
		err = fileBytesPacket.EncryptBody(encrKey)
		if err != nil {
			return sentBytes, err
		}
	}

	// send it to the other side
	err = SendPacket(connection, fileBytesPacket)
	if err != nil {
		return 0, err
	}

	return sentBytes, nil
}

// Sends a symlink to the other side. If encrKey is not nil - encrypts the packet with this key
func SendSymlink(symlink *fsys.Symlink, connection net.Conn, encrKey []byte) error {
	symlinkPacket := Packet{
		Header: HeaderSymlink,
	}

	symlinkPacketBodyBuff := new(bytes.Buffer)

	// SYMLINK~(string size in binary)(location in the filesystem)(string size in binary)(location of a target)

	binary.Write(symlinkPacketBodyBuff, binary.BigEndian, uint64(len(symlink.Path)))
	symlinkPacketBodyBuff.Write([]byte(symlink.Path))

	binary.Write(symlinkPacketBodyBuff, binary.BigEndian, uint64(len(symlink.TargetPath)))
	symlinkPacketBodyBuff.Write([]byte(symlink.TargetPath))

	symlinkPacket.Body = symlinkPacketBodyBuff.Bytes()

	if encrKey != nil {
		err := symlinkPacket.EncryptBody(encrKey)
		if err != nil {
			return err
		}
	}

	err := SendPacket(connection, symlinkPacket)
	if err != nil {
		return err
	}

	return nil
}
