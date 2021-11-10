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
