package sender

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/Unbewohnte/ftu/checksum"
	"github.com/Unbewohnte/ftu/encryption"
	"github.com/Unbewohnte/ftu/protocol"
)

// The main sender struct
type Sender struct {
	Port            int
	FileToTransfer  *file
	Listener        net.Listener
	Connection      net.Conn
	IncomingPackets chan protocol.Packet
	EncryptionKey   []byte
	TransferInfo    *transferInfo
	TransferAllowed bool // the receiver had agreed to receive a file
	ReceiverIsReady bool // receiver is waiting for a new packet
	Stopped         bool // controlls the mainloop
}

// Creates a new sender with default|necessary fields
func NewSender(port int, filepath string) *Sender {
	fileToTransfer, err := getFile(filepath)
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	incomingPacketsChan := make(chan protocol.Packet, 100)

	remoteIP, err := GetRemoteIP()
	if err != nil {
		// don`t panic if couldn`t get remote IP
		remoteIP = ""
	}

	localIP, err := GetLocalIP()
	if err != nil {
		panic(err)
	}

	// !!!
	key := encryption.Generate32AESkey()
	fmt.Printf("Generated an encryption key: %s\n", key)

	fmt.Printf("Created a new sender at %s:%d (remote)\n%s:%d (local)\n\n", remoteIP, port, localIP, port)
	return &Sender{
		Port:            port,
		FileToTransfer:  fileToTransfer,
		Listener:        listener,
		Connection:      nil,
		IncomingPackets: incomingPacketsChan,
		TransferInfo: &transferInfo{
			SentFileBytesPackets:    0,
			ApproximateNumOfPackets: uint64(float32(fileToTransfer.Filesize) / float32(protocol.MAXPACKETSIZE)),
		},
		EncryptionKey:   key,
		TransferAllowed: false,
		ReceiverIsReady: false,
		Stopped:         false,
	}
}

// When the interrupt signal is sent - exit cleanly
func (s *Sender) HandleInterrupt() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		<-signalChan
		s.Stop()
	}()
}

// Closes the connection, warns about it the receiver and exits the mainloop
func (s *Sender) Stop() {
	if s.Connection != nil {
		disconnectionPacket := protocol.Packet{
			Header: protocol.HeaderDisconnecting,
		}
		err := protocol.SendEncryptedPacket(s.Connection, disconnectionPacket, s.EncryptionKey)
		if err != nil {
			panic(fmt.Sprintf("could not send a disconnection packet: %s", err))
		}

		s.Connection.Close()
	}

	s.Stopped = true
}

// Accepts one connection
func (s *Sender) WaitForConnection() {
	connection, err := s.Listener.Accept()
	if err != nil {
		fmt.Printf("Could not accept a connection: %s", err)
		os.Exit(-1)
	}
	s.Connection = connection
	fmt.Println("New connection from ", s.Connection.RemoteAddr())
}

// Closes the listener. Used only when there is still no connection from `AcceptConnections`
func (s *Sender) StopListening() {
	s.Listener.Close()
}

// Sends generated earlier eas encryption key to receiver
func (s *Sender) SendEncryptionKey() error {

	keyPacket := protocol.Packet{
		Header: protocol.HeaderEncryptionKey,
		Body:   s.EncryptionKey,
	}
	err := protocol.SendPacket(s.Connection, keyPacket)
	if err != nil {
		return fmt.Errorf("could not send a packet: %s", err)
	}

	return nil
}

// Sends multiple packets with all information about the file to receiver
// (filename, filesize, checksum)
func (s *Sender) SendOffer() error {
	// filename
	filenamePacket := protocol.Packet{
		Header: protocol.HeaderFilename,
		Body:   []byte(s.FileToTransfer.Filename),
	}
	err := protocol.SendEncryptedPacket(s.Connection, filenamePacket, s.EncryptionKey)
	if err != nil {
		return fmt.Errorf("could not send an information about the file: %s", err)
	}

	// filesize
	filesizePacket := protocol.Packet{
		Header: protocol.HeaderFileSize,
		Body:   []byte(strconv.Itoa(int(s.FileToTransfer.Filesize))),
	}

	err = protocol.SendEncryptedPacket(s.Connection, filesizePacket, s.EncryptionKey)
	if err != nil {
		return fmt.Errorf("could not send an information about the file: %s", err)
	}

	// checksum
	checksumPacket := protocol.Packet{
		Header: protocol.HeaderChecksum,
		Body:   checksum.ChecksumToBytes(s.FileToTransfer.CheckSum),
	}
	err = protocol.SendEncryptedPacket(s.Connection, checksumPacket, s.EncryptionKey)
	if err != nil {
		return fmt.Errorf("could not send an information about the file: %s", err)
	}

	// indicate that we`ve sent everything we needed to send
	donePacket := protocol.Packet{
		Header: protocol.HeaderDone,
	}
	err = protocol.SendEncryptedPacket(s.Connection, donePacket, s.EncryptionKey)
	if err != nil {
		return fmt.Errorf("could not send an information about the file: %s", err)
	}
	return nil
}

// Sends one packet that contains a piece of file to the receiver
func (s *Sender) SendPiece() error {
	// if no data to send - exit
	if s.FileToTransfer.LeftBytes == 0 {
		fmt.Printf("Done. Sent %d file packets\nTook %v\n", s.TransferInfo.SentFileBytesPackets, time.Since(s.TransferInfo.StartTime))
		s.Stop()
	}

	// empty body
	fileBytesPacket := protocol.Packet{
		Header: protocol.HeaderFileBytes,
	}

	// how many bytes we can send at maximum (including some little space for padding)
	maxFileBytes := protocol.MAXPACKETSIZE - (uint(protocol.MeasurePacketSize(fileBytesPacket)) + 90)

	fileBytes := make([]byte, maxFileBytes)
	// if there is less data to send than the limit - create a buffer of needed size
	if s.FileToTransfer.LeftBytes < uint64(maxFileBytes) {
		fileBytes = make([]byte, uint64(maxFileBytes)-(uint64(maxFileBytes)-s.FileToTransfer.LeftBytes))
	}

	// reading bytes from the point where we left
	read, err := s.FileToTransfer.Handler.ReadAt(fileBytes, int64(s.FileToTransfer.SentBytes))
	if err != nil {
		return fmt.Errorf("could not read from a file: %s", err)
	}

	// filling BODY with bytes
	fileBytesPacket.Body = fileBytes

	err = protocol.SendEncryptedPacket(s.Connection, fileBytesPacket, s.EncryptionKey)
	if err != nil {
		return fmt.Errorf("could not send a file packet : %s", err)
	}

	// doing a "logging" for the next piece
	s.FileToTransfer.LeftBytes -= uint64(read)
	s.FileToTransfer.SentBytes += uint64(read)
	s.TransferInfo.SentFileBytesPackets++

	return nil
}

// Prints a brief information about the state of the transfer
func (s *Sender) PrintTransferInfo(pauseDuration time.Duration) {
	next := time.Now().UTC()
	s.TransferInfo.StartTime = next
	for {
		if !s.TransferAllowed {
			time.Sleep(time.Second)
			continue
		}

		now := time.Now().UTC()

		if !now.After(next) {
			continue
		}
		next = now.Add(pauseDuration)

		fmt.Printf(`
 | Sent packets/Approximate number of packets
 | (%d|%d) (%.2f%%/100%%)
`, s.TransferInfo.SentFileBytesPackets,
			s.TransferInfo.ApproximateNumOfPackets,
			float32(s.TransferInfo.SentFileBytesPackets)/float32(s.TransferInfo.ApproximateNumOfPackets)*100)

		time.Sleep(pauseDuration)
	}
}

// Listens in an endless loop; reads incoming packets, decrypts their BODY and puts into channel
func (s *Sender) ReceivePackets() {
	for {
		incomingPacketBytes, err := protocol.ReadFromConn(s.Connection)
		if err != nil {
			fmt.Printf("Error reading a packet: %s\nExiting...", err)
			s.Stop()
			os.Exit(-1)
		}

		incomingPacket := protocol.BytesToPacket(incomingPacketBytes)

		decryptedBody, err := encryption.Decrypt(s.EncryptionKey, incomingPacket.Body)
		if err != nil {
			fmt.Printf("Error decrypting an incoming packet: %s\nExiting...", err)
			s.Stop()
			os.Exit(-1)
		}

		incomingPacket.Body = decryptedBody

		s.IncomingPackets <- incomingPacket
	}
}

// The "head" of the sender. "Glues" all things together.
// Current structure allows the sender to receive any type of packet
// in any order and react correspondingly
func (s *Sender) MainLoop() {
	// receive and print in separate goroutines
	go s.ReceivePackets()
	go s.PrintTransferInfo(time.Second * 3)

	// instantly sending an encryption key, following the protocol`s rule
	err := s.SendEncryptionKey()
	if err != nil {
		fmt.Printf("Could not send an encryption key: %s\nExiting...", err)
		s.Stop()
	}

	// send an information about the shared file to the receiver
	err = s.SendOffer()
	if err != nil {
		fmt.Printf("Could not send an info about the file: %s\nExiting...", err)
		s.Stop()
	}

	for {
		if s.Stopped {
			break
		}

		if s.TransferAllowed && s.ReceiverIsReady {
			err := s.SendPiece()
			if err != nil {
				fmt.Printf("could not send a piece of file: %s", err)
				s.Stop()
			}
			s.ReceiverIsReady = false
		}

		// no incoming packets ? Skipping the packet handling part
		if len(s.IncomingPackets) == 0 {
			continue
		}

		incomingPacket := <-s.IncomingPackets

		// handling each packet header differently
		switch incomingPacket.Header {

		case protocol.HeaderAccept:
			// allowed to send file packets
			fmt.Println("The transfer has been accepted !")
			s.TransferAllowed = true

		case protocol.HeaderReject:
			fmt.Println("The transfer has been rejected")
			s.Stop()

		case protocol.HeaderReady:
			s.ReceiverIsReady = true

		case protocol.HeaderDisconnecting:
			// receiver is dropping the file transfer ?
			fmt.Println("Receiver has disconnected")
			s.Stop()
		}
	}
}
