package server

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/Unbewohnte/FTU/protocol"
)

// gets a local ip. Borrowed from StackOverflow, thank you, whoever I brought it from
func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

// gets a remote ip. Borrowed from StackOverflow, thank you, whoever I brought it from
func GetRemoteIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", fmt.Errorf("could not make a request to get your remote IP: %s", err)
	}
	defer resp.Body.Close()
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read a response: %s", err)
	}
	return string(ip), nil
}

// The main server struct
type Server struct {
	Port            int
	FileToTransfer  *File
	Listener        net.Listener
	Connection      net.Conn
	IncomingPackets chan protocol.Packet
	CanTransfer     bool
	Stopped         bool
}

// Creates a new server with default fields
func NewServer(port int, filepath string) *Server {
	fileToTransfer, err := getFile(filepath)
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	incomingPacketsChan := make(chan protocol.Packet, 5)

	remoteIP, err := GetRemoteIP()
	if err != nil {
		panic(err)
	}
	localIP, err := GetLocalIP()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created a new server at %s:%d (remote)\n%s:%d (local)\n", remoteIP, port, localIP, port)
	return &Server{
		Port:            port,
		FileToTransfer:  fileToTransfer,
		Listener:        listener,
		Connection:      nil,
		IncomingPackets: incomingPacketsChan,
		Stopped:         false,
	}
}

// Closes the connection, warns about it the client and exits the mainloop
func (s *Server) Stop() {
	disconnectionPacket := protocol.Packet{
		Header: protocol.HeaderDisconnecting,
	}
	err := protocol.SendPacket(s.Connection, disconnectionPacket)
	if err != nil {
		panic(fmt.Sprintf("could not send a disconnection packet: %s", err))
	}

	s.Stopped = true
	s.Disconnect()
}

// Closes current connection
func (s *Server) Disconnect() {
	s.Connection.Close()
}

// Accepts one connection
func (s *Server) WaitForConnection() {
	connection, err := s.Listener.Accept()
	if err != nil {
		fmt.Printf("Could not accept a connection: %s", err)
		os.Exit(-1)
	}
	s.Connection = connection
	fmt.Println("New connection from ", s.Connection.RemoteAddr())
}

// Closes the listener. Used only when there is still no connection from `AcceptConnections`
func (s *Server) StopListening() {
	s.Listener.Close()
}

// Sends a packet with all information about a file to current connection
func (s *Server) SendOffer() error {
	err := protocol.SendPacket(s.Connection, protocol.Packet{
		Header:   protocol.HeaderFileInfo,
		Filename: s.FileToTransfer.Filename,
		Filesize: s.FileToTransfer.Filesize,
	})
	if err != nil {
		return fmt.Errorf("could not send an information about the file: %s", err)
	}

	return nil
}

// Sends one file packet to the client
func (s *Server) SendPiece() error {
	// if no data to send - exit
	if s.FileToTransfer.LeftBytes == 0 {
		fmt.Printf("Done. Sent %d file packets\n", s.FileToTransfer.SentPackets)
		s.Stop()
	}

	fileBytes := make([]byte, protocol.MAXFILEDATASIZE)
	// if there is less data to send than the limit - create a buffer of needed size
	if s.FileToTransfer.LeftBytes < uint64(protocol.MAXFILEDATASIZE) {
		fileBytes = make([]byte, protocol.MAXFILEDATASIZE-(protocol.MAXFILEDATASIZE-int(s.FileToTransfer.LeftBytes)))
	}

	// reading bytes from the point where we left
	read, err := s.FileToTransfer.Handler.ReadAt(fileBytes, int64(s.FileToTransfer.SentBytes))
	if err != nil {
		return fmt.Errorf("could not read from a file: %s", err)
	}

	// constructing a file packet and sending it
	fileDataPacket := protocol.Packet{
		Header:   protocol.HeaderFileData,
		Filename: s.FileToTransfer.Filename,
		Filesize: s.FileToTransfer.Filesize,
		FileData: fileBytes,
	}

	err = protocol.SendPacket(s.Connection, fileDataPacket)
	if err != nil {
		return fmt.Errorf("could not send a file packet : %s", err)
	}

	// doing a "logging" for the next time
	s.FileToTransfer.LeftBytes -= uint64(read)
	s.FileToTransfer.SentBytes += uint64(read)
	s.FileToTransfer.SentPackets++

	return nil
}

// Listens in an endless loop; reads incoming packages and puts them into channel
func (s *Server) ReceivePackets() {
	for {
		incomingPacket, err := protocol.ReadFromConn(s.Connection)
		if err != nil {
			// in current implementation there is no way to receive a working file even if only one packet is missing
			fmt.Printf("Error reading a packet: %s\nExiting...", err)
			os.Exit(-1)
		}
		s.IncomingPackets <- incomingPacket
	}
}

// The "head" of the server. "Glues" all things together.
// Current structure allows the server to receive any type of packet
// in any order and react correspondingly
func (s *Server) MainLoop() {

	go s.ReceivePackets()

	// send an information about the shared file to the client
	s.SendOffer()

	for {
		if s.Stopped {
			// exit the mainloop
			break
		}
		// send, then process received packets

		// no incoming packets ? Skipping the packet handling part
		if len(s.IncomingPackets) == 0 {
			continue
		}

		incomingPacket := <-s.IncomingPackets

		// handling each packet header differently
		switch incomingPacket.Header {

		case protocol.HeaderAccept:
			fmt.Printf("Client has accepted the transfer !\n")
			// allowed to send file packets
			s.CanTransfer = true

		case protocol.HeaderReject:
			fmt.Println("Client has rejected the transfer")
			s.Stop()

		// client is ready to receive the next file packet
		case protocol.HeaderReady:
			if !s.CanTransfer {
				break
			}
			err := s.SendPiece()
			if err != nil {
				fmt.Printf("Could not send a piece of file: %s", err)
				os.Exit(-1)
			}

		case protocol.HeaderDisconnecting:
			s.Stop()
		}
	}
}
