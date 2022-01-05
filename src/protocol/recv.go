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

	// fmt.Printf("[RECV] read from connection: %s; length: %d\n", packetBuffer.Bytes()[:30], packetBuffer.Len())

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
