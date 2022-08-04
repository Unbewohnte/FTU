/*
ftu - file transferring utility.
Copyright (C) 2021,2022  Kasyanov Nikolay Alexeyevich (Unbewohnte)

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

package protocol

import (
	"bytes"
	"net"
	"testing"
)

func Test_WriteRead(t *testing.T) {
	packet := Packet{
		Header: "randomheader",
		Body:   []byte("fIlEnAmE.txt"),
	}

	// a valid representation of received packet`s bytes
	packetBytes, err := packet.ToBytes()
	if err != nil {
		t.Fatalf("%s", err)
	}

	// imitating a connection
	l, err := net.Listen("tcp", ":9999")
	if err != nil {
		t.Fatalf("%s", err)
	}
	c, err := net.Dial("tcp", "localhost:9999")
	if err != nil {
		t.Fatalf("%s", err)
	}
	cc, err := l.Accept()
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer c.Close()
	defer cc.Close()

	// sending packet
	err = SendPacket(cc, packet)
	if err != nil {
		t.Fatalf("SendPacket failed: %s", err)
	}

	// reading it from c
	receivedPacket, err := ReadFromConn(c)
	if err != nil {
		t.Fatalf("ReadFromConn failed: %s", err)
	}

	// drop packetsize for valid packet bytes because they are also dropped in ReadFromConn
	packetBytes = packetBytes[8:]

	for index, b := range receivedPacket {
		if b != packetBytes[index] {
			t.Fatalf("Error: packet bytes do not match: expected %v, got: %v; valid packet: %v; received packet: %v", string(packetBytes[index]), string(b), packetBytes, receivedPacket)
		}
	}
}

func Test_BytesToPacket(t *testing.T) {
	packet := Packet{
		Header: HeaderFileBytes,
		Body:   []byte("fIlEnAmE.txt"),
	}

	packetBuffer := new(bytes.Buffer)
	packetBuffer.Write([]byte(packet.Header))
	packetBuffer.Write([]byte(HEADERDELIMETER))
	packetBuffer.Write(packet.Body)

	// a valid representation of received packet`s bytes
	packetBytes := packetBuffer.Bytes()

	convertedPacket, err := BytesToPacket(packetBytes)
	if err != nil {
		t.Fatalf("BytesToPacket error: %s", err)
	}

	if convertedPacket.Header != packet.Header || string(convertedPacket.Body) != string(packet.Body) {
		t.Fatalf("BytesToPacket error: header or body of converted packet does not match with the original")
	}
}
