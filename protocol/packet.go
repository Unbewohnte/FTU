// This file describes the general packet structure and provides methods to work with them before|after the transportation

// Examples of packets, ready for transportation in pseudo-code:
// []byte(|34|FILEDATA~fe2[gkr3j930f]fwpglkrt[o])
// []byte(|57|FILENAME~theBestFileNameEver_Existed_in'''theUniverse.txt)
// general structure:
// PACKETSIZEDELIMETER packetsize PACKETSIZEDELIMETER packet.Header HEADERDELIMETER packet.Body (without spaces between)
package protocol

import (
	"bytes"
	"fmt"
	"net"
	"strconv"

	"github.com/Unbewohnte/ftu/encryption"
)

// Internal representation of packet before|after the transportation
type Packet struct {
	Header Header
	Body   []byte
}

// Returns a size of the given packet as if it would be sent and presented in bytes.
// ie: FILESIZE~[49 49 56 55 56 53 50 49 54]
// DOES COUNT THE PACKETSIZEDELIMETER
func MeasurePacketSize(packet Packet) uint64 {
	packetBytes := new(bytes.Buffer)
	packetBytes.Write([]byte(packet.Header))
	packetBytes.Write([]byte(HEADERDELIMETER))
	packetBytes.Write(packet.Body)

	return uint64(packetBytes.Len())
}

// Converts packet bytes into Packet struct
func BytesToPacket(packetbytes []byte) Packet {
	var header Header
	var body []byte

	for counter, b := range packetbytes {
		if string(b) == HEADERDELIMETER {
			header = Header(packetbytes[0:counter])
			body = packetbytes[counter+1:]
			break
		}
	}
	return Packet{
		Header: header,
		Body:   body,
	}
}

// Converts given packet struct into ready-to-transfer bytes, constructed by following the protocol
func PacketToBytes(packet Packet) ([]byte, error) {
	packetSize := MeasurePacketSize(packet)

	if packetSize > uint64(MAXPACKETSIZE) {
		return nil, fmt.Errorf("invalid packet!: EXCEEDED MAX PACKETSIZE")
	}

	packetSizeBytes := []byte(strconv.Itoa(int(packetSize)))

	// creating a buffer and writing the whole packet into it
	packetBuffer := new(bytes.Buffer)

	// packetsize between delimeters (ie: |17|)
	packetBuffer.Write([]byte(PACKETSIZEDELIMETER))
	packetBuffer.Write(packetSizeBytes)
	packetBuffer.Write([]byte(PACKETSIZEDELIMETER))

	// ie: FILENAME~file.txt
	packetBuffer.Write([]byte(packet.Header))
	packetBuffer.Write([]byte(HEADERDELIMETER))
	packetBuffer.Write(packet.Body)

	// for debug purposes (ᗜˬᗜ)
	// fmt.Printf("SENDING PACKET: %s%s%s%s%s%s\n",
	// 	[]byte(PACKETSIZEDELIMETER), packetSizeBytes, []byte(PACKETSIZEDELIMETER),
	// 	[]byte(packetToSend.Header), []byte(HEADERDELIMETER), packetToSend.Body)

	return packetBuffer.Bytes(), nil
}

// Sends given packet to connection, following all the protocol`s rules.
// ALL packets MUST be sent by this method
func SendPacket(connection net.Conn, packetToSend Packet) error {
	packetBytes, err := PacketToBytes(packetToSend)
	if err != nil {
		return fmt.Errorf("could not convert given packet to bytes: %s", err)
	}
	// write the result (ie: |17|FILENAME~file.png)
	connection.Write(packetBytes)

	return nil
}

// Sends given packet to connection, as the normal `SendPacket` method, but
// encodes given packet`s BODY with AES encryption
func SendEncryptedPacket(connection net.Conn, packetToSend Packet, key []byte) error {
	// encrypting packet`s body
	encryptedBody, err := encryption.Encrypt(key, packetToSend.Body)
	if err != nil {
		return fmt.Errorf("could not encrypt packet`s body: %s", err)
	}
	packetToSend.Body = encryptedBody

	// sending the encrypted packet
	err = SendPacket(connection, packetToSend)
	if err != nil {
		return fmt.Errorf("could not send packet: %s", err)
	}

	return nil
}

// Reads a packet from given connection, returns its bytes.
// ASSUMING THAT THE PACKETS ARE SENT BY `SendPacket` function !!!!
func ReadFromConn(connection net.Conn) ([]byte, error) {
	var err error
	var delimeterCounter int = 0
	var packetSizeStrBuffer string = ""
	var packetSize int = 0

	for {
		buffer := make([]byte, 1)
		connection.Read(buffer)

		if string(buffer) == PACKETSIZEDELIMETER {
			delimeterCounter++

			// the first delimeter has been found, skipping the rest of the loop
			if delimeterCounter == 1 {
				continue
			}
		}

		// the last delimeter, the next read will be the packet itself, so breaking
		if delimeterCounter == 2 {
			break
		}

		packetSizeStrBuffer += string(buffer)
	}

	packetSize, err = strconv.Atoi(packetSizeStrBuffer)
	if err != nil {
		return nil, fmt.Errorf("could not convert packetsizeStr into int: %s", err)
	}

	// have a packetsize, now reading the whole packet
	packetBuffer := new(bytes.Buffer)

	// splitting big-sized packet into chunks and constructing it from pieces
	left := packetSize
	for {
		if left == 0 {
			break
		}
		buff := make([]byte, 1024)
		if left < len(buff) {
			buff = make([]byte, left)
		}

		read, _ := connection.Read(buff)
		left -= read

		packetBuffer.Write(buff[:read])
	}

	return packetBuffer.Bytes(), nil
}
