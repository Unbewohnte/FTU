package receiver

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/Unbewohnte/FTU/checksum"
	"github.com/Unbewohnte/FTU/protocol"
)

// Representation of a receiver
type Receiver struct {
	DownloadsFolder string
	Connection      net.Conn
	IncomingPackets chan protocol.Packet
	FileToDownload  *File
	Stopped         bool
	ReadyToReceive  bool
	PacketCounter   uint64
}

// Creates a new client with default fields
func NewReceiver(downloadsFolder string) *Receiver {
	os.MkdirAll(downloadsFolder, os.ModePerm)

	downloadsFolderInfo, err := os.Stat(downloadsFolder)
	if err != nil {
		panic(err)
	}
	if !downloadsFolderInfo.IsDir() {
		panic("Downloads folder is not a directory")
	}

	incomingPacketsChan := make(chan protocol.Packet, 5)

	var PacketCounter uint64 = 0
	fmt.Println("Created a new client")
	return &Receiver{
		DownloadsFolder: downloadsFolder,
		Connection:      nil,
		IncomingPackets: incomingPacketsChan,
		Stopped:         false,
		ReadyToReceive:  false,
		PacketCounter:   PacketCounter,
	}
}

// Closes the connection
func (r *Receiver) Disconnect() {
	r.Connection.Close()
}

// Closes the connection, warns the sender and exits the mainloop
func (r *Receiver) Stop() {
	disconnectionPacket := protocol.Packet{
		Header: protocol.HeaderDisconnecting,
	}
	protocol.SendPacket(r.Connection, disconnectionPacket)
	r.Stopped = true
	r.Disconnect()
}

// Connects to a given address over tcp. Sets a connection to a corresponding field in receiver
func (r *Receiver) Connect(addr string) error {
	fmt.Printf("Trying to connect to %s...\n", addr)
	connection, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("could not connect to %s: %s", addr, err)
	}
	r.Connection = connection
	fmt.Println("Connected to ", r.Connection.RemoteAddr())

	return nil
}

// Handles the fileinfo packet. The choice of acceptance is given to the user
func (r *Receiver) HandleFileOffer(fileinfoPacket protocol.Packet) error {

	// inform the user about the file
	// fmt.Printf("Incoming fileinfo packet:\nFilename: %s\nFilesize: %.3fMB\nCheckSum: %s\nAccept ? [Y/N]: ",
	// fileinfoPacket.Filename, float32(fileinfoPacket.Filesize)/1024/1024, fileinfoPacket.FileCheckSum)

	fmt.Printf(`
 Incoming fileinfo packet:
 | Filename: %s
 | Filesize: %.3fMB
 | Checksum: %x
 | 
 | Download ? [Y/N]: `,
		fileinfoPacket.Filename, float32(fileinfoPacket.Filesize)/1024/1024, fileinfoPacket.FileCheckSum,
	)

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
		err := protocol.SendPacket(r.Connection, rejectionPacket)
		if err != nil {
			return fmt.Errorf("could not send a rejection packet: %s", err)
		}

		r.ReadyToReceive = false

		return nil
	}
	// accept the file

	r.FileToDownload = &File{
		Filename: fileinfoPacket.Filename,
		Filesize: fileinfoPacket.Filesize,
		CheckSum: fileinfoPacket.FileCheckSum,
	}

	acceptancePacket := protocol.Packet{
		Header: protocol.HeaderAccept,
	}
	err := protocol.SendPacket(r.Connection, acceptancePacket)
	if err != nil {
		return fmt.Errorf("could not send an acceptance packet: %s", err)
	}

	// can and ready to receive file packets
	r.ReadyToReceive = true

	return nil
}

// Handles the download by writing incoming bytes into the file
func (r *Receiver) WritePieceOfFile(filePacket protocol.Packet) error {
	r.ReadyToReceive = false

	// open|create a file with the same name as the filepacket`s file name
	file, err := os.OpenFile(filepath.Join(r.DownloadsFolder, filePacket.Filename), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	// just write the filedata
	file.Write(filePacket.FileData)
	file.Close()
	r.PacketCounter++

	r.ReadyToReceive = true

	return nil
}

// Listens in an endless loop; reads incoming packages and puts them into channel
func (r *Receiver) ReceivePackets() {
	for {
		incomingPacket, err := protocol.ReadFromConn(r.Connection)
		if err != nil {
			// in current implementation there is no way to receive a working file even if only one packet is missing
			fmt.Printf("Error reading a packet: %s\nExiting...", err)
			os.Exit(-1)
		}
		r.IncomingPackets <- incomingPacket
	}
}

// The "head" of the receiver. Similarly as in server`s logic "glues" everything together.
// Current structure allows the receiver to receive any type of packet
// in any order and react correspondingly
func (r *Receiver) MainLoop() {
	go r.ReceivePackets()

	for {
		if r.Stopped {
			// exit the mainloop
			break
		}

		// send a packet telling sender to send another piece of file
		if r.ReadyToReceive {
			readyPacket := protocol.Packet{
				Header: protocol.HeaderReady,
			}
			protocol.SendPacket(r.Connection, readyPacket)
			r.ReadyToReceive = false
		}

		// no incoming packets ? Skipping the packet handling part
		if len(r.IncomingPackets) == 0 {
			continue
		}

		// take the packet and handle depending on the header
		incomingPacket := <-r.IncomingPackets

		// handling each packet header differently
		switch incomingPacket.Header {

		case protocol.HeaderFileInfo:
			go r.HandleFileOffer(incomingPacket)

		case protocol.HeaderFileData:
			go r.WritePieceOfFile(incomingPacket)

		case protocol.HeaderDisconnecting:
			// the sender has completed its mission,
			// checking hashes and exiting

			fmt.Println("Got ", r.PacketCounter, " packets in total")
			fmt.Println("Checking checksums...")

			file, err := os.Open(filepath.Join(r.DownloadsFolder, r.FileToDownload.Filename))
			if err != nil {
				fmt.Printf("error while opening downloaded file for checking: %s\n", err)
				os.Exit(-1)
			}
			realCheckSum, err := checksum.GetPartialCheckSum(file)
			if err != nil {
				fmt.Printf("error perfoming partial checksum: %s\n", err)
				os.Exit(-1)
			}

			fmt.Printf("\n%x ----- %x\n", r.FileToDownload.CheckSum, realCheckSum)
			if !checksum.AreEqual(realCheckSum, r.FileToDownload.CheckSum) {
				fmt.Println("Downloaded file is corrupted !")
			}
			r.Stop()
		}
	}
}
