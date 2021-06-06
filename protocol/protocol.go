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
func ReadPacketBytes(packetBytes []byte) Packet {
	// makes sure that the packet is ALWAYS less or equal to the maximum packet size
	// this allows not to use any client or server checks

	//fmt.Println("READING packet: ", string(packetBytes))

	var packet Packet
	err := json.Unmarshal(packetBytes, &packet)
	if err != nil {
		fmt.Printf("Could not unmarshal the packet: %s\n", err)
		return Packet{}
	}
	return packet
}

// Converts `Packet` struct into []byte
func EncodePacket(packet Packet) []byte {
	packetBytes, err := json.Marshal(packet)
	if err != nil {
		return []byte("")
	}
	return packetBytes
}

// Measures the packet length
func MeasurePacket(packet Packet) uint64 {
	packetBytes := EncodePacket(packet)
	return uint64(len(packetBytes))
}

// Checks if given packet is valid, returns a boolean and an explanation message
func IsValidPacket(packet Packet) (bool, string) {
	if MeasurePacket(packet) > uint64(MAXPACKETSIZE) {
		return false, "Exceeded MAXPACKETSIZE"
	}
	if len(packet.FileData) > MAXFILEDATASIZE {
		return false, "Exceeded MAXFILEDATASIZE"
	}

	if strings.TrimSpace(string(packet.Header)) == "" {
		return false, "Blank header"
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

	packetSize := MeasurePacket(packet)

	// write packetsize between delimeters (ie: |727|{"HEADER":"PING"...})
	connection.Write([]byte(fmt.Sprintf("%s%d%s", PACKETSIZEDELIMETER, packetSize, PACKETSIZEDELIMETER)))

	// write an actual packet
	connection.Write(EncodePacket(packet))

	//fmt.Println("Sending packet: ", string(EncodePacket(packet)), "  Length: ", packetSize)
	return nil
}

// Reads a packet from a connection by retrieving the packet length. Only once
// ASSUMING THAT THE PACKETS ARE SENT BY `SendPacket` method !!!!
func ReadFromConn(connection net.Conn) Packet {
	var gotPacketSize bool = false
	var delimeterCounter int = 0

	var packetSizeStr string = ""
	var packetSize int = 0
	for {
		// still need to get a packetsize
		if !gotPacketSize {
			// reading byte-by-byte
			buffer := make([]byte, 1)
			connection.Read(buffer)

			// found a delimeter
			if string(buffer) == PACKETSIZEDELIMETER {
				delimeterCounter++

				// the first delimeter is found, skipping the rest of the code
				if delimeterCounter == 1 {
					continue
				}
			}

			// found the first delimeter, skip was performed, now reading an actual packetsize
			if delimeterCounter == 1 {
				packetSizeStr += string(buffer)
			} else if delimeterCounter == 2 {
				// found the last delimeter, thus already read the whole packetsize
				packetSize, _ = strconv.Atoi(packetSizeStr)
				gotPacketSize = true
			}
			// skipping the rest of the code because we don`t know the packet size yet
			continue
		}
		// have a packetsize, now reading the whole packet

		//fmt.Println("Got a packetsize!: ", packetSize)
		packetBuffer := make([]byte, packetSize)
		connection.Read(packetBuffer)

		packet := ReadPacketBytes(packetBuffer)

		isvalid, _ := IsValidPacket(packet)
		if isvalid {
			return packet
		}

		break
	}

	return Packet{}
}
