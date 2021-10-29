package node

import "github.com/Unbewohnte/ftu/protocol"

// Implementation for sender and receiver nodes. [I`ll Probably remove it later. I don`t see the use-cases rn]
type Noder interface {
	Connect(addr string, port uint) error
	Disconnect() error
	Listen(packetPipe chan protocol.Packet)
	Send(packet protocol.Packet) error
}
