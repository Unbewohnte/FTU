package client

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/Unbewohnte/FTU/protocol"
)

// Representation of a tcp client
type Client struct {
	DownloadsFolder string
	Connection      net.Conn
	IncomingPackets chan protocol.Packet
	Stopped         bool
	ReadyToReceive  bool
	PacketCounter   uint64
}

// Creates a new client with default fields
func NewClient(downloadsFolder string) *Client {
	os.MkdirAll(downloadsFolder, os.ModePerm)

	info, err := os.Stat(downloadsFolder)
	if err != nil {
		panic(err)
	}
	if !info.IsDir() {
		panic("Downloads folder is not a directory")
	}

	incomingPacketsChan := make(chan protocol.Packet, 5)

	var PacketCounter uint64 = 0
	fmt.Println("Created a new client")
	return &Client{
		DownloadsFolder: downloadsFolder,
		Connection:      nil,
		IncomingPackets: incomingPacketsChan,
		Stopped:         false,
		ReadyToReceive:  false,
		PacketCounter:   PacketCounter,
	}
}

// Closes the connection
func (c *Client) Disconnect() {
	c.Connection.Close()
}

// Closes the connection, warns the server and exits the mainloop
func (c *Client) Stop() {
	disconnectionPacket := protocol.Packet{
		Header: protocol.HeaderDisconnecting,
	}
	protocol.SendPacket(c.Connection, disconnectionPacket)
	c.Stopped = true
	c.Disconnect()
}

// Connects to a given address over tcp. Sets a connection to a client
func (c *Client) Connect(addr string) error {
	fmt.Printf("Trying to connect to %s...\n", addr)
	connection, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("could not connect to %s: %s", addr, err)
	}
	c.Connection = connection
	fmt.Println("Connected to ", c.Connection.RemoteAddr())

	return nil
}

// Handles the fileinfo packet. The choice of acceptance is given to the user
func (c *Client) HandleFileOffer(fileinfoPacket protocol.Packet) error {

	// inform the user about the file
	fmt.Printf("Incoming fileinfo packet:\nFilename: %s\nFilesize: %.3fMB\nAccept ? [Y/N]: ",
		fileinfoPacket.Filename, float32(fileinfoPacket.Filesize)/1024/1024)

	// get and process the input
	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)

	// reject the file
	if input != "y" {
		rejectionPacket := protocol.Packet{
			Header: protocol.HeaderReject,
		}
		err := protocol.SendPacket(c.Connection, rejectionPacket)
		if err != nil {
			return fmt.Errorf("could not send a rejection packet: %s", err)
		}

		c.ReadyToReceive = false

		return nil
	}

	// accept the file
	acceptancePacket := protocol.Packet{
		Header: protocol.HeaderAccept,
	}
	err := protocol.SendPacket(c.Connection, acceptancePacket)
	if err != nil {
		return fmt.Errorf("could not send an acceptance packet: %s", err)
	}

	// can and ready to receive file packets
	c.ReadyToReceive = true

	return nil
}

// Handles the download by writing incoming bytes into the file
func (c *Client) WritePieceOfFile(filePacket protocol.Packet) error {
	c.ReadyToReceive = false

	// open|create a file with the same name as the filepacket`s file name
	file, err := os.OpenFile(filepath.Join(c.DownloadsFolder, filePacket.Filename), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	// just write the filedata
	file.Write(filePacket.FileData)
	file.Close()
	c.PacketCounter++

	c.ReadyToReceive = true

	return nil
}

// Listens in an endless loop; reads incoming packages and puts them into channel
func (c *Client) ReceivePackets() {
	for {
		incomingPacket := protocol.ReadFromConn(c.Connection)
		isvalid, _ := protocol.IsValidPacket(incomingPacket)
		if !isvalid {
			continue
		}
		c.IncomingPackets <- incomingPacket
	}
}

// The "head" of the client. Similarly as in server`s logic "glues" everything together.
// Current structure allows the client to receive any type of packet
// in any order and react correspondingly
func (c *Client) MainLoop() {
	go c.ReceivePackets()

	for {
		if c.Stopped {
			// exit the mainloop
			break
		}
		// 1) send -> 2) handle received if necessary

		// send a packet telling server to send another piece of file
		if c.ReadyToReceive {
			readyPacket := protocol.Packet{
				Header: protocol.HeaderReady,
			}
			protocol.SendPacket(c.Connection, readyPacket)
			c.ReadyToReceive = false
		}

		// no incoming packets ? Skipping the packet handling part
		if len(c.IncomingPackets) == 0 {
			continue
		}

		// take the packet and handle depending on the header
		incomingPacket := <-c.IncomingPackets

		// handling each packet header differently
		switch incomingPacket.Header {

		case protocol.HeaderFileInfo:
			go c.HandleFileOffer(incomingPacket)

		case protocol.HeaderFileData:
			go c.WritePieceOfFile(incomingPacket)

		case protocol.HeaderDisconnecting:
			// the server is ded, no need to stay alive as well
			fmt.Println("Done. Got ", c.PacketCounter, " packets in total")
			c.Stopped = true
			c.Stop()
		}
	}
}
