package node

import (
	"net"

	"github.com/Unbewohnte/ftu/protocol"
)

type NodeInnerStates struct {
	Connected  bool
	InTransfer bool
	IsWaiting  bool
	Stopped    bool
}

type Security struct {
	EncryptionKey []byte
}

// Server and a client in one type !
type Node struct {
	conn       net.Conn
	packetPipe chan []protocol.Packet
	State      *NodeInnerStates
	Security   *Security
}

// Creates a new either a server-side or client-side node
func NewNode(options *NodeOptions) (*Node, error) {

	node := Node{}
	return &node, nil
}

func (node *Node) Connect(addr string, port uint) error {
	return nil
}

func (node *Node) Disconnect() error {
	if node.State.Connected {
		err := node.conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (node *Node) Send(packet protocol.Packet) error {
	return nil
}

func (node *Node) Listen() error {
	return nil
}

