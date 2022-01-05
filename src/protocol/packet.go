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

// General packet structure and methods to work with them before|after the transportation

// Packet structure during transportation:
// (size of the whole packet in binary (big endian uint64))(packet header)(header delimeter (~))(packet contents)

package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/Unbewohnte/ftu/encryption"
)

// Internal representation of packet before|after the transportation
type Packet struct {
	Header Header
	Body   []byte
}

var ErrorInvalidPacket error = fmt.Errorf("invalid packet header or body")

// Returns a size of the given packet as if it would be sent and presented in bytes.
// ie: FILE~bytes_here
func (packet *Packet) Size() uint64 {
	packetBytes := new(bytes.Buffer)
	packetBytes.Write([]byte(packet.Header))
	packetBytes.Write([]byte(HEADERDELIMETER))
	packetBytes.Write(packet.Body)

	return uint64(packetBytes.Len())
}

// Converts packet bytes into Packet struct
func BytesToPacket(packetbytes []byte) (*Packet, error) {
	// check if there`s a header delimiter present
	pString := string(packetbytes)
	if !strings.Contains(pString, HEADERDELIMETER) {
		return nil, ErrorInvalidPacket
	}

	var header Header
	var body []byte

	for counter, b := range packetbytes {
		if string(b) == HEADERDELIMETER {
			header = Header(packetbytes[0:counter])
			body = packetbytes[counter+1:]
			break
		}
	}

	return &Packet{
		Header: header,
		Body:   body,
	}, nil
}

var ErrorExceededMaxPacketsize error = fmt.Errorf("packet is too big")

// Converts given packet struct into ready-to-transfer bytes, constructed by following the protocol
func (packet *Packet) ToBytes() ([]byte, error) {
	packetSize := packet.Size()

	if packetSize > uint64(MAXPACKETSIZE) {
		return nil, ErrorExceededMaxPacketsize
	}

	// creating a buffer and writing the whole packet into it
	packetBuffer := new(bytes.Buffer)

	// packet size bytes
	err := binary.Write(packetBuffer, binary.BigEndian, &packetSize)
	if err != nil {
		return nil, err
	}

	// header, delimeter and body ie: FILENAME~file.txt
	packetBuffer.Write([]byte(packet.Header))
	packetBuffer.Write([]byte(HEADERDELIMETER))
	packetBuffer.Write(packet.Body)

	return packetBuffer.Bytes(), nil
}

// Encrypts packet`s BODY with AES encryption
func (packet *Packet) EncryptBody(key []byte) error {
	// encrypting packet`s body
	encryptedBody, err := encryption.Encrypt(key, packet.Body)
	if err != nil {
		return err
	}
	packet.Body = encryptedBody

	return nil
}

// Decrypts packet`s BODY with AES decryption
func (packet *Packet) DecryptBody(key []byte) error {
	if len(packet.Body) == 0 {
		return nil
	}

	decryptedBody, err := encryption.Decrypt(key, packet.Body)
	if err != nil {
		return err
	}

	packet.Body = decryptedBody

	return nil
}
