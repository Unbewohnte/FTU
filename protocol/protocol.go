package protocol

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// a package that describes how server and client should communicate

const MAXPACKETSIZE int = 2048  // whole packet
const MAXFILEDATASIZE int = 512 // only `FileData` | MUST be less than `MAXPACKETSIZE`
const PACKETSIZEDELIMETER string = "|"

// Headers
type Header string

const HeaderFileData Header = "FILEDATA"
const HeaderFileInfo Header = "FILEINFO"
const HeaderReject Header = "FILE_REJECT"
const HeaderAccept Header = "FILE_ACCEPT"
const HeaderReady Header = "READY"
const HeaderDisconnecting Header = "BYE!"

// Packet structure.
// A Packet without a header is an invalid packet
type Packet struct {
	Header   Header `json:"Header"`
	Filename string `json:"Filename"`
	Filesize uint64 `json:"Filesize"`
	FileData []byte `json:"Filedata"`
}

// converts valid packet bytes into `Packet` struct
func ReadPacketBytes(packetBytes []byte) (Packet, error) {
	var packet Packet
	err := json.Unmarshal(packetBytes, &packet)
	if err != nil {
		return Packet{}, fmt.Errorf("could not unmarshal packet bytes: %s", err)
	}
	return packet, nil
}

// Converts `Packet` struct into []byte
func EncodePacket(packet Packet) ([]byte, error) {
	packetBytes, err := json.Marshal(packet)
	if err != nil {
		return nil, fmt.Errorf("could not marshal packet bytes: %s", err)
	}
	return packetBytes, nil
}

// Measures the packet length
func MeasurePacket(packet Packet) (uint64, error) {
	packetBytes, err := EncodePacket(packet)
	if err != nil {
		return 0, fmt.Errorf("could not measure the packet: %s", err)
	}
	return uint64(len(packetBytes)), nil
}

// Checks if given packet is valid, returns a boolean and an explanation message
func IsValidPacket(packet Packet) (bool, string) {
	packetSize, err := MeasurePacket(packet)
	if err != nil {
		return false, "Measurement error"
	}
	if packetSize > uint64(MAXPACKETSIZE) {
		return false, "Exceeded MAXPACKETSIZE"
	}
	if len(packet.FileData) > MAXFILEDATASIZE {
		return false, "Exceeded MAXFILEDATASIZE"
	}

	if strings.TrimSpace(string(packet.Header)) == "" {
		return false, "Empty header"
	}
	return true, ""
}

// Sends a given packet to connection using a special sending format
// ALL packets MUST be sent by this method
func SendPacket(connection net.Conn, packet Packet) error {
	isvalid, msg := IsValidPacket(packet)
	if !isvalid {
		return fmt.Errorf("this packet is invalid !: %v; The error: %v", packet, msg)
	}

	packetSize, err := MeasurePacket(packet)
	if err != nil {
		return err
	}

	// write packetsize between delimeters (ie: |727|{"HEADER":"PING"...})
	connection.Write([]byte(fmt.Sprintf("%s%d%s", PACKETSIZEDELIMETER, packetSize, PACKETSIZEDELIMETER)))

	// write the packet itself
	packetBytes, err := EncodePacket(packet)
	if err != nil {
		return fmt.Errorf("could not send a packet: %s", err)
	}
	connection.Write(packetBytes)

	return nil
}

// Reads a packet from given connection.
// ASSUMING THAT THE PACKETS ARE SENT BY `SendPacket` function !!!!
func ReadFromConn(connection net.Conn) (Packet, error) {
	var err error
	var delimeterCounter int = 0
	var packetSizeStrBuffer string = ""
	var packetSize int = 0

	for {
		// reading byte-by-byte
		buffer := make([]byte, 1)
		connection.Read(buffer) // no fixed time limit, so no need to check for an error

		// found a delimeter
		if string(buffer) == PACKETSIZEDELIMETER {
			delimeterCounter++

			// the first delimeter has been found, skipping the rest of the loop
			if delimeterCounter == 1 {
				continue
			}
		}

		// found the first delimeter, skip was performed, now reading an actual packetsize
		if delimeterCounter == 1 {
			// adding a character of the packet size to the `string buffer`; ie: | <- read, reading now -> 1 23|PACKET_HERE
			packetSizeStrBuffer += string(buffer)

		} else if delimeterCounter == 2 {
			// found the last delimeter, thus already read the whole packetsize
			// converting from string to int
			packetSize, err = strconv.Atoi(packetSizeStrBuffer)
			if err != nil {
				return Packet{}, fmt.Errorf("could not convert packet size into integer: %s", err)
			}
			// packet size has been found, breaking from the loop
			break
		}
	}

	// have a packetsize, now reading the whole packet
	packetBuffer := make([]byte, packetSize)
	connection.Read(packetBuffer)

	packet, err := ReadPacketBytes(packetBuffer)
	if err != nil {
		return Packet{}, err
	}

	isvalid, msg := IsValidPacket(packet)
	if isvalid {
		return packet, nil
	}

	return Packet{}, fmt.Errorf("received an invalid packet. Reason: %s", msg)
}
