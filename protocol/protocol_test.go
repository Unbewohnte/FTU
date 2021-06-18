package protocol

import (
	"bytes"
	"net"
	"testing"
)

// Practically tests the whole protocol
func TestTransfer(t *testing.T) {
	packet := Packet{
		Header: HeaderFilename,
		Body:   []byte("fIlEnAmE.txt"),
	}

	packetBuffer := new(bytes.Buffer)
	packetBuffer.Write([]byte(packet.Header))
	packetBuffer.Write([]byte(HEADERDELIMETER))
	packetBuffer.Write(packet.Body)

	// a valid representation of received packet`s bytes
	packetBytes := packetBuffer.Bytes()

	// imitating a connection
	l, err := net.Listen("tcp", ":9999")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	c, err := net.Dial("tcp", "localhost:9999")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	cc, err := l.Accept()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	defer c.Close()
	defer cc.Close()

	// sending packet
	err = SendPacket(cc, packet)
	if err != nil {
		t.Errorf("SendPacket failed: %s", err)
	}

	//
	receivedPacket, err := ReadFromConn(c)
	if err != nil {
		t.Errorf("ReadFromConn failed: %s", err)
	}

	for index, b := range receivedPacket {
		if b != packetBytes[index] {
			t.Errorf("Failed: wanted: %v, got: %v", packetBytes[index], b)
		}
	}
}

func TestBytesToPacket(t *testing.T) {
	packet := Packet{
		Header: HeaderFilename,
		Body:   []byte("fIlEnAmE.txt"),
	}

	packetBuffer := new(bytes.Buffer)
	packetBuffer.Write([]byte(packet.Header))
	packetBuffer.Write([]byte(HEADERDELIMETER))
	packetBuffer.Write(packet.Body)

	// a valid representation of received packet`s bytes
	packetBytes := packetBuffer.Bytes()

	convertedPacket := BytesToPacket(packetBytes)

	if convertedPacket.Header != packet.Header || string(convertedPacket.Body) != string(packet.Body) {
		t.Errorf("BytesToPacket failed")
	}
}
