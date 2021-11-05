package node

import (
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fmt"

	"github.com/Unbewohnte/ftu/addr"
	"github.com/Unbewohnte/ftu/checksum"
	"github.com/Unbewohnte/ftu/encryption"
	"github.com/Unbewohnte/ftu/fsys"
	"github.com/Unbewohnte/ftu/protocol"
)

// node-controlling states
type NodeInnerStates struct {
	Stopped           bool // the way to exit the mainloop in case of an external error or a successful end of a transfer
	AllowedToTransfer bool // the way to notify the mainloop of a sending node to start sending pieces of files
}

// Network specific settings
type Net struct {
	ConnAddr      string   // address to connect to. Does not include port
	Conn          net.Conn // the core TCP connection of the node. Self-explanatory
	Port          uint     // a port to connect to/listen on
	EncryptionKey []byte   // if != nil - incoming packets will be decrypted with it and outcoming packets will be encrypted
}

// Both sending-side and receiving-side information
type TransferInfo struct {
	Ready         bool         // is the other node ready to receive another piece
	ServingPath   string       // path to the thing that will be sent
	Recursive     bool         // recursively send directory
	AcceptedFiles []*fsys.File // files that`ve been accepted to be received
	DownloadsPath string       // where to download
}

// Sender and receiver in one type !
type Node struct {
	PacketPipe   chan *protocol.Packet // a way to receive incoming packets from another goroutine
	IsSending    bool                  // sending or a receiving node
	Net          *Net
	State        *NodeInnerStates
	TransferInfo *TransferInfo
}

// Creates a new either a sending or receiving node with specified options
func NewNode(options *NodeOptions) (*Node, error) {
	node := Node{
		PacketPipe: make(chan *protocol.Packet, 100),
		IsSending:  options.IsSending,
		Net: &Net{
			Port:          options.WorkingPort,
			ConnAddr:      options.ClientSide.ConnectionAddr,
			EncryptionKey: nil,
			Conn:          nil,
		},
		State: &NodeInnerStates{
			AllowedToTransfer: false,
			Stopped:           false,
		},
		TransferInfo: &TransferInfo{
			ServingPath:   options.ServerSide.ServingPath,
			Recursive:     options.ServerSide.Recursive,
			AcceptedFiles: nil,
			DownloadsPath: options.ClientSide.DownloadsFolderPath,
		},
	}
	return &node, nil
}

// Connect node to another listening one with a pre-defined address&&port
func (node *Node) connect() error {
	if node.Net.Port == 0 {
		node.Net.Port = 7270
	}

	fmt.Printf("Connecting to %s:%d...\n", node.Net.ConnAddr, node.Net.Port)

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", node.Net.ConnAddr, node.Net.Port), time.Second*5)
	if err != nil {
		return err
	}

	fmt.Printf("Connected\n")

	node.Net.Conn = conn

	return nil
}

// Notify the other node and close the connection
func (node *Node) disconnect() error {
	if node.Net.Conn != nil {
		// notify the other node and close the connection
		err := protocol.SendPacket(node.Net.Conn, protocol.Packet{
			Header: protocol.HeaderDisconnecting,
		})
		if err != nil {
			return err
		}

		err = node.Net.Conn.Close()
		if err != nil {
			return err
		}

		node.State.Stopped = true
	}

	return nil
}

// Wait for a connection on a pre-defined port
func (node *Node) waitForConnection() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", node.Net.Port))
	if err != nil {
		return err
	}

	// accept only one conneciton
	connection, err := listener.Accept()
	if err != nil {
		return err
	}

	fmt.Printf("New connection from %s\n", connection.RemoteAddr().String())

	node.Net.Conn = connection

	return nil
}

// Starts the node in either sending or receiving state and performs the transfer
func (node *Node) Start() {
	switch node.IsSending {
	case true:
		// SENDER

		// retrieve necessary information, wait for connection
		localIP, err := addr.GetLocal()
		if err != nil {
			panic(err)
		}

		file, err := fsys.GetFile(node.TransferInfo.ServingPath)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Sending \"%s\" (%.2f MB) locally on %s:%d\n", file.Name, float32(file.Size)/1024/1024, localIP, node.Net.Port)

		// wain for another node to connect
		err = node.waitForConnection()
		if err != nil {
			panic(err)
		}

		// generate and send encryption key
		encrKey := encryption.Generate32AESkey()
		node.Net.EncryptionKey = encrKey
		fmt.Printf("Generated encryption key: %s\n", encrKey)

		err = sendEncryptionKey(node.Net.Conn, encrKey)
		if err != nil {
			panic(err)
		}

		// listen for incoming packets
		go receivePackets(node.Net.Conn, node.PacketPipe)

		// send info on file/directory
		go sendFilePacket(node.Net.Conn, file, node.Net.EncryptionKey)

		// mainloop
		for {
			if node.State.Stopped {
				node.disconnect()
				break
			}

			// receive incoming packets and decrypt them if necessary
			incomingPacket, ok := <-node.PacketPipe
			if !ok {
				node.State.Stopped = true
			}

			if node.Net.EncryptionKey != nil {
				err = incomingPacket.DecryptBody(node.Net.EncryptionKey)
				if err != nil {
					panic(err)
				}
			}

			// react based on a header of a received packet
			switch incomingPacket.Header {
			case protocol.HeaderReady:
				// the other node is ready to receive file data
				node.TransferInfo.Ready = true

			case protocol.HeaderAccept:
				node.State.AllowedToTransfer = true

				fmt.Printf("Transfer allowed. Sending...\n")

			case protocol.HeaderReject:
				node.State.Stopped = true

				fmt.Printf("Transfer rejected. Disconnecting...")

			case protocol.HeaderDisconnecting:
				node.State.Stopped = true

				fmt.Printf("%s disconnected\n", node.Net.Conn.RemoteAddr())
			}

			// if allowed to transfer and the other node is ready to receive packets - send one piece
			// and wait for it to be ready again
			if node.State.AllowedToTransfer && node.TransferInfo.Ready {
				err = sendPiece(file, node.Net.Conn, node.Net.EncryptionKey)
				if err != nil {
					if err == ErrorSentAll {
						// the file has been sent fully
						fileIDBuff := new(bytes.Buffer)
						err = binary.Write(fileIDBuff, binary.BigEndian, file.ID)
						if err != nil {
							panic(err)
						}

						endFilePacket := protocol.Packet{
							Header: protocol.HeaderEndfile,
							Body:   fileIDBuff.Bytes(),
						}

						if node.Net.EncryptionKey != nil {
							err = endFilePacket.EncryptBody(node.Net.EncryptionKey)
							if err != nil {
								panic(err)
							}
						}

						protocol.SendPacket(node.Net.Conn, endFilePacket)

						// because there`s still no handling for directories - send
						// done packet
						protocol.SendPacket(node.Net.Conn, protocol.Packet{
							Header: protocol.HeaderDone,
						})

						node.State.Stopped = true
					} else {
						node.State.Stopped = true

						fmt.Printf("An error occured when sending a piece of \"%s\": %s\n", file.Name, err)
						panic(err)
					}
				}

				node.TransferInfo.Ready = false
			}
		}

	case false:
		// RECEIVER

		// connect to the sending node
		err := node.connect()
		if err != nil {
			fmt.Printf("Could not connect to %s:%d\n", node.Net.ConnAddr, node.Net.Port)
			os.Exit(-1)
		}

		// listen for incoming packets
		go receivePackets(node.Net.Conn, node.PacketPipe)

		// mainloop
		for {
			if node.State.Stopped {
				node.disconnect()
				break
			}

			// receive incoming packets and decrypt them if necessary
			incomingPacket, ok := <-node.PacketPipe
			if !ok {
				break
			}
			if node.Net.EncryptionKey != nil {
				err = incomingPacket.DecryptBody(node.Net.EncryptionKey)
				if err != nil {
					panic(err)
				}
			}

			// react based on a header of a received packet
			switch incomingPacket.Header {

			case protocol.HeaderFile:
				// process an information about a singe file. Accept or reject the transfer
				go func() {
					file, err := decodeFilePacket(incomingPacket)
					if err != nil {
						panic(err)
					}

					fmt.Printf("| Filename: %s\n| Size: %.2f MB\n| Checksum: %s\n", file.Name, float32(file.Size)/1024/1024, file.Checksum)
					var answer string
					fmt.Printf("| Download ? [Y/n]: ")
					fmt.Scanln(&answer)
					fmt.Printf("\n\n")

					responsePacketFileIDBuffer := new(bytes.Buffer)
					binary.Write(responsePacketFileIDBuffer, binary.BigEndian, file.ID)

					if strings.EqualFold(answer, "y") || answer == "" {
						// yes

						err = os.MkdirAll(node.TransferInfo.DownloadsPath, os.ModePerm)
						if err != nil {
							panic(err)
						}

						fullFilePath := filepath.Join(node.TransferInfo.DownloadsPath, file.Name)

						// check if the file already exists; if yes - remove it and replace with a new one
						_, err := os.Stat(fullFilePath)
						if err == nil {
							// exists
							// remove it
							os.Remove(fullFilePath)
						}

						file.Path = fullFilePath
						file.Open()

						node.TransferInfo.AcceptedFiles = append(node.TransferInfo.AcceptedFiles, file)

						// send aceptance packet
						acceptancePacket := protocol.Packet{
							Header: protocol.HeaderAccept,
							Body:   responsePacketFileIDBuffer.Bytes(),
						}
						if node.Net.EncryptionKey != nil {
							err = acceptancePacket.EncryptBody(node.Net.EncryptionKey)
							if err != nil {
								panic(err)
							}
						}

						err = protocol.SendPacket(node.Net.Conn, acceptancePacket)
						if err != nil {
							panic(err)
						}

						// notify the node that we`re ready to transportation. No need
						// for encryption because the body is nil
						err = protocol.SendPacket(node.Net.Conn, protocol.Packet{
							Header: protocol.HeaderReady,
						})
						if err != nil {
							panic(err)
						}

					} else {
						// no
						rejectionPacket := protocol.Packet{
							Header: protocol.HeaderReject,
							Body:   responsePacketFileIDBuffer.Bytes(),
						}

						if node.Net.EncryptionKey != nil {
							err = rejectionPacket.EncryptBody(node.Net.EncryptionKey)
							if err != nil {
								panic(err)
							}
						}

						err = protocol.SendPacket(node.Net.Conn, rejectionPacket)
						if err != nil {
							panic(err)
						}

						node.State.Stopped = true
					}
				}()

			case protocol.HeaderFileBytes:
				// check if this file has been accepted to receive
				fileIDReader := bytes.NewReader(incomingPacket.Body)
				var fileID uint64
				err := binary.Read(fileIDReader, binary.BigEndian, &fileID)
				if err != nil {
					panic(err)
				}

				for _, acceptedFile := range node.TransferInfo.AcceptedFiles {
					if acceptedFile.ID == fileID {
						// accepted

						// append provided bytes to the file

						fileBytes := incomingPacket.Body[8:]
						_, err = acceptedFile.Handler.Write(fileBytes)
						if err != nil {
							panic(err)
						}
					}
				}

				// notify the other one that this node is ready
				err = protocol.SendPacket(node.Net.Conn, protocol.Packet{
					Header: protocol.HeaderReady,
				})
				if err != nil {
					panic(err)
				}

			case protocol.HeaderEndfile:
				// one of the files has been received completely,
				// compare checksums and check if it is the last
				// file in the transfer
				// (TODO)
				fileIDReader := bytes.NewReader(incomingPacket.Body)
				var fileID uint64
				err := binary.Read(fileIDReader, binary.BigEndian, &fileID)
				if err != nil {
					panic(err)
				}

				for index, acceptedFile := range node.TransferInfo.AcceptedFiles {
					if acceptedFile.ID == fileID {
						// accepted

						// close the handler afterwards
						defer acceptedFile.Handler.Close()

						// remove this file from the pool
						node.TransferInfo.AcceptedFiles = append(node.TransferInfo.AcceptedFiles[:index], node.TransferInfo.AcceptedFiles[index+1:]...)

						// compare checksums
						realChecksum, err := checksum.GetPartialCheckSum(acceptedFile.Handler)
						if err != nil {
							panic(err)
						}

						fmt.Printf("| Checking hashes for file \"%s\"\n", acceptedFile.Name)
						if realChecksum != acceptedFile.Checksum {
							fmt.Printf("| %s --- %s file is corrupted\n", realChecksum, acceptedFile.Checksum)
							break
						} else {
							fmt.Printf("| %s --- %s\n", realChecksum, acceptedFile.Checksum)
							break
						}
					}
				}

				// node.State.Stopped = true

			case protocol.HeaderEncryptionKey:
				// retrieve the key
				packetReader := bytes.NewReader(incomingPacket.Body)

				var keySize uint64
				binary.Read(packetReader, binary.BigEndian, &keySize)

				encrKey := make([]byte, keySize)
				packetReader.Read(encrKey)

				node.Net.EncryptionKey = encrKey

			case protocol.HeaderDone:
				node.State.Stopped = true

			case protocol.HeaderDisconnecting:
				node.State.Stopped = true

				fmt.Printf("%s disconnected\n", node.Net.Conn.RemoteAddr())
			}
		}

	}
}
