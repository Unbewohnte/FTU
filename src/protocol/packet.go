// This file describes the general packet structure and provides methods to work with them before|after the transportation

// General packet structure:
// (size of the whole packet in binary)(packet header)(header delimeter (~))(packet contents)

package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"github.com/Unbewohnte/ftu/encryption"
)

// Internal representation of packet before|after the transportation
type Packet struct {
	Header Header
	Body   []byte
}

// Returns a size of the given packet as if it would be sent and presented in bytes.
// ie: FILE~bytes_here
func (packet *Packet) Size() uint64 {
	packetBytes := new(bytes.Buffer)
	packetBytes.Write([]byte(packet.Header))
	packetBytes.Write([]byte(HEADERDELIMETER))
	packetBytes.Write(packet.Body)

	return uint64(packetBytes.Len())
}

var ErrorNotPacketBytes error = fmt.Errorf("not packet bytes")

// Converts packet bytes into Packet struct
func BytesToPacket(packetbytes []byte) (*Packet, error) {
	// check if there`s a header delimiter present
	pString := string(packetbytes)
	if !strings.Contains(pString, HEADERDELIMETER) {
		return nil, ErrorNotPacketBytes
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

var ErrorExceededMaxPacketsize error = fmt.Errorf("the packet is too big")

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

// Reads a packet from given connection, returns its bytes.
// ASSUMING THAT THE PACKETS ARE SENT BY `SendPacket` function !!!!
func ReadFromConn(connection net.Conn) ([]byte, error) {
	var packetSize uint64
	err := binary.Read(connection, binary.BigEndian, &packetSize)
	if err != nil {
		return nil, err
	}

	// have a packetsize, now reading the whole packet
	packetBuffer := new(bytes.Buffer)

	// splitting a big-sized packet into chunks and constructing it from pieces
	left := packetSize
	for {
		if left == 0 {
			break
		}

		buff := make([]byte, 8192)
		if left < uint64(len(buff)) {
			buff = make([]byte, left)
		}

		read, _ := connection.Read(buff)
		left -= uint64(read)

		packetBuffer.Write(buff[:read])
	}

	// fmt.Printf("[RECV] read from connection: %s; length: %d\n", packetBuffer.Bytes()[:30], packetBuffer.Len())

	return packetBuffer.Bytes(), nil
}
