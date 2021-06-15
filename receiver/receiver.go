package receiver

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Unbewohnte/FTU/checksum"
	"github.com/Unbewohnte/FTU/encryption"
	"github.com/Unbewohnte/FTU/protocol"
)

// Representation of a receiver
type Receiver struct {
	DownloadsFolder        string
	Connection             net.Conn
	IncomingPackets        chan protocol.Packet
	FileToDownload         *File
	EncryptionKey          []byte
	ReadyToReceive         bool
	Stopped                bool
	FileBytesPacketCounter uint64
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

	incomingPacketsChan := make(chan protocol.Packet, 5000)

	var PacketCounter uint64 = 0
	fmt.Println("Created a new receiver")
	return &Receiver{
		DownloadsFolder: downloadsFolder,
		Connection:      nil,
		IncomingPackets: incomingPacketsChan,
		Stopped:         false,
		ReadyToReceive:  false,
		FileToDownload: &File{
			Filename: "",
			Filesize: 0,
		},
		FileBytesPacketCounter: PacketCounter,
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
	protocol.SendEncryptedPacket(r.Connection, disconnectionPacket, r.EncryptionKey)
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

// Prints known information about the file that is about to be transported.
// Handles the input from the user after the sender sent "DOYOUACCEPT?" packet.
// The choice of acceptance is given to the user
func (r *Receiver) HandleFileOffer() error {

	// inform the user about the file

	fmt.Printf(`
 Incoming fileinfo packet:
 | Filename: %s
 | Filesize: %.3fMB
 | Checksum: %x
 | 
 | Download ? [Y/N]: `,
		r.FileToDownload.Filename, float32(r.FileToDownload.Filesize)/1024/1024, r.FileToDownload.CheckSum,
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
		err := protocol.SendEncryptedPacket(r.Connection, rejectionPacket, r.EncryptionKey)
		if err != nil {
			return fmt.Errorf("could not send a rejection packet: %s", err)
		}

		return nil
	}
	// accept the file

	// check if the file with the same name is present
	doesExist, err := r.CheckIfFileAlreadyExists()
	if err != nil {
		return fmt.Errorf("could not check if the file with the same name alredy exists: %s", err)
	}

	if doesExist {
		fmt.Printf(`
 | Looks like that there is a file with the same name in your downloads directory, do you want to overwrite it ? [Y/N]: `)

		fmt.Scanln(&input)
		input = strings.TrimSpace(input)
		input = strings.ToLower(input)

		if input == "y" {
			err = os.Remove(filepath.Join(r.DownloadsFolder, r.FileToDownload.Filename))
			if err != nil {
				return fmt.Errorf("could not remove the file: %s", err)
			}
		} else {
			// user did not agree to overwrite, adding checksum to the name
			r.FileToDownload.Filename = fmt.Sprint(time.Now().Unix()) + r.FileToDownload.Filename
		}
	}

	acceptancePacket := protocol.Packet{
		Header: protocol.HeaderAccept,
	}
	err = protocol.SendEncryptedPacket(r.Connection, acceptancePacket, r.EncryptionKey)
	if err != nil {
		return fmt.Errorf("could not send an acceptance packet: %s", err)
	}

	return nil
}

// Handles the download by writing incoming bytes into the file
func (r *Receiver) WritePieceOfFile(filePacket protocol.Packet) error {
	if filePacket.Header != protocol.HeaderFileBytes {
		return fmt.Errorf("packet with given header should not contain filebytes !: %v", filePacket)
	}

	// open|create a file with the same name as the filepacket`s file name
	file, err := os.OpenFile(filepath.Join(r.DownloadsFolder, r.FileToDownload.Filename), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	// just write the bytes
	file.Write(filePacket.Body)
	file.Close()
	r.FileBytesPacketCounter++

	return nil
}

// Listens in an endless loop; reads incoming packets, decrypts their BODY and puts into channel
func (r *Receiver) ReceivePackets() {
	for {
		incomingPacketBytes, err := protocol.ReadFromConn(r.Connection)
		if err != nil {
			fmt.Printf("Error reading a packet: %s\nExiting...", err)
			r.Stop()
			os.Exit(-1)
		}

		incomingPacket := protocol.BytesToPacket(incomingPacketBytes)

		// if this is the FIRST packet - it has HeaderEncryptionKey, so no need to decrypt
		if incomingPacket.Header == protocol.HeaderEncryptionKey {
			r.IncomingPackets <- incomingPacket
			continue
		}

		decryptedBody, err := encryption.Decrypt(r.EncryptionKey, incomingPacket.Body)
		if err != nil {
			fmt.Printf("Error decrypring incoming packet`s BODY: %s\nExiting...", err)
			r.Stop()
			os.Exit(-1)
		}

		incomingPacket.Body = decryptedBody

		r.IncomingPackets <- incomingPacket
	}
}

// The "head" of the receiver. Similarly as in server`s logic "glues" everything together.
// Current structure allows the receiver to receive any type of packet
// in any order and react correspondingly
func (r *Receiver) MainLoop() {
	go r.ReceivePackets()

	// r.Stop()

	for {
		if r.Stopped {
			break
		}

		if r.ReadyToReceive {
			readyPacket := protocol.Packet{
				Header: protocol.HeaderReady,
			}
			err := protocol.SendEncryptedPacket(r.Connection, readyPacket, r.EncryptionKey)
			if err != nil {
				fmt.Printf("Could not send the packet: %s\nExiting...", err)
				r.Stop()
			}

			r.ReadyToReceive = false
		}

		// no incoming packets ? Skipping the packet handling part
		if len(r.IncomingPackets) == 0 {
			continue
		}

		incomingPacket := <-r.IncomingPackets

		// handling each packet header differently
		switch incomingPacket.Header {

		case protocol.HeaderEncryptionKey:
			r.EncryptionKey = incomingPacket.Body
			fmt.Println("Got the encryption key: ", string(incomingPacket.Body))

		case protocol.HeaderFilename:
			r.FileToDownload.Filename = string(incomingPacket.Body)

		case protocol.HeaderFileSize:
			filesize, err := strconv.Atoi(string(incomingPacket.Body))
			if err != nil {
				fmt.Printf("could not convert a filesize: %s\n", err)
				r.Stop()
			}
			r.FileToDownload.Filesize = uint64(filesize)

		case protocol.HeaderChecksum:
			checksum, err := checksum.BytesToChecksum(incomingPacket.Body)
			if err != nil {
				fmt.Printf("could not get file`s checksum: %s\n", err)
				r.Stop()
			}
			r.FileToDownload.CheckSum = checksum

		case protocol.HeaderDone:
			if r.FileToDownload.Filename != "" && r.FileToDownload.Filesize != 0 && r.FileToDownload.CheckSum != [32]byte{} {
				err := r.HandleFileOffer()
				if err != nil {
					fmt.Printf("Could not handle a file download confirmation: %s\nExiting...", err)
					r.Stop()
				}
				r.ReadyToReceive = true
			} else {
				fmt.Println("Not enough data about the file was sent. Exiting...")
				r.Stop()
			}

		case protocol.HeaderFileBytes:
			err := r.WritePieceOfFile(incomingPacket)
			if err != nil {
				fmt.Printf("Could not write a piece of file: %s\nExiting...", err)
				r.Stop()
			}
			r.ReadyToReceive = true

		case protocol.HeaderDisconnecting:
			// the sender has completed its mission,
			// checking hashes and exiting

			fmt.Println("Got ", r.FileBytesPacketCounter, " file packets in total")
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
