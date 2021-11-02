package node

type ServerSideNodeOptions struct {
	ServingPath string
	Recursive   bool
}

type ClientSideNodeOptions struct {
	ConnectionAddr      string
	DownloadsFolderPath string
}

// Options to configure the node
type NodeOptions struct {
	IsSending   bool
	WorkingPort uint
	ServerSide  *ServerSideNodeOptions
	ClientSide  *ClientSideNodeOptions
}
