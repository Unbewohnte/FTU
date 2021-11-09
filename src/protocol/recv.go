// Methods allowing to receive and preprocess packets from connection
package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

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

	fmt.Printf("[RECV] read from connection: %s; length: %d\n", packetBuffer.Bytes()[:30], packetBuffer.Len())

	return packetBuffer.Bytes(), nil
}

var ErrorNotConnected error = fmt.Errorf("not connected")

// Reads packets from connection in an endless loop, sends them to the channel
func ReceivePackets(connection net.Conn, packetPipe chan *Packet) error {
	for {
		if connection == nil {
			return ErrorNotConnected
		}

		packetBytes, err := ReadFromConn(connection)
		if err != nil {
			close(packetPipe)
			return err
		}

		incomingPacket, err := BytesToPacket(packetBytes)
		if err != nil {
			close(packetPipe)
			return err
		}

		packetPipe <- incomingPacket
	}
}
