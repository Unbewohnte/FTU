package node

// Implementation for sender and receiver node
type Noder interface {
	Connect(addr string, port uint) error
	Disconnect() error
	Listen(dataPipe chan []byte)
	Send(data []byte) error
}
