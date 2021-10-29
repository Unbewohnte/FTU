package node

import (
	"github.com/Unbewohnte/ftu/fs"
	"github.com/Unbewohnte/ftu/protocol"
)

type ServerSideNodeOptions struct {
	ServingDirectory *fs.Directory // Can be set to nil
	ServingFile      *fs.File      // Can be set to nil
}

type ClientSideNodeOptions struct {
	DownloadsFolder *fs.Directory // Must be set during the Node creation, even if it will be changed afterwards
}

// Options to configure the node
type NodeOptions struct {
	WorkingPort uint
	PacketPipe  chan protocol.Packet
	ServerSide  *ServerSideNodeOptions
	ClientSide  *ClientSideNodeOptions
}
