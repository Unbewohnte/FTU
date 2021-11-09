package node

import (
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fmt"

	"github.com/Unbewohnte/ftu/addr"
	"github.com/Unbewohnte/ftu/checksum"
	"github.com/Unbewohnte/ftu/encryption"
	"github.com/Unbewohnte/ftu/fsys"
	"github.com/Unbewohnte/ftu/protocol"
)

// node-controlling states
type nodeInnerstates struct {
	Stopped           bool // the way to exit the mainloop in case of an external error or a successful end of a transfer
	AllowedToTransfer bool // the way to notify the mainloop of a sending node to start sending pieces of files
}

// netInfowork specific settings
type netInfoInfo struct {
	ConnAddr      string   // address to connect to. Does not include port
	Conn          net.Conn // the core TCP connection of the node. Self-explanatory
	Port          uint     // a port to connect to/listen on
	EncryptionKey []byte   // if != nil - incoming packets will be decrypted with it and outcoming packets will be encrypted
}

// Sending-side node information
type sending struct {
	ServingPath   string // path to the thing that will be sent
	IsDirectory   bool   // is ServingPath a directory
	Recursive     bool   // recursively send directory
	CanSendBytes  bool   // is the other node ready to receive another piece
	FilesToSend   []*fsys.File
	CurrentFileID uint64 // an id of a file that is currently being transported
}

// Receiving-side node information
type receiving struct {
	AcceptedFiles []*fsys.File // files that`ve been accepted to be received
	DownloadsPath string       // where to download
}

// Both sending-side and receiving-side information
type transferInfo struct {
	Receiving *receiving
	Sending   *sending
}

// Sender and receiver in one type !
type Node struct {
	mutex        *sync.Mutex
	packetPipe   chan *protocol.Packet // a way to receive incoming packets from another goroutine
	isSending    bool                  // sending or a receiving node
	netInfo      *netInfoInfo
	state        *nodeInnerstates
	transferInfo *transferInfo
}

// Creates a new either a sending or receiving node with specified options
func NewNode(options *NodeOptions) (*Node, error) {
	var isDir bool
	if options.IsSending {
		// sending node preparation
		sendingPathStats, err := os.Stat(options.ServerSide.ServingPath)
		if err != nil {
			return nil, err
		}

		switch sendingPathStats.IsDir() {
		case true:
			isDir = true

		case false:
			isDir = false
		}
	} else {
		// receiving node preparation
		err := os.MkdirAll(options.ClientSide.DownloadsFolderPath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	node := Node{
		mutex:      &sync.Mutex{},
		packetPipe: make(chan *protocol.Packet, 100),
		isSending:  options.IsSending,
		netInfo: &netInfoInfo{
			Port:          options.WorkingPort,
			ConnAddr:      options.ClientSide.ConnectionAddr,
			EncryptionKey: nil,
			Conn:          nil,
		},
		state: &nodeInnerstates{
			AllowedToTransfer: false,
			Stopped:           false,
		},
		transferInfo: &transferInfo{
			Sending: &sending{
				ServingPath: options.ServerSide.ServingPath,
				Recursive:   options.ServerSide.Recursive,
				IsDirectory: isDir,
			},
			Receiving: &receiving{
				AcceptedFiles: nil,
				DownloadsPath: options.ClientSide.DownloadsFolderPath,
			},
		},
	}
	return &node, nil
}

// Connect node to another listening one with a pre-defined address&&port
func (node *Node) connect() error {
	if node.netInfo.Port == 0 {
		node.netInfo.Port = 7270
	}

	fmt.Printf("Connecting to %s:%d...\n", node.netInfo.ConnAddr, node.netInfo.Port)

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", node.netInfo.ConnAddr, node.netInfo.Port), time.Second*5)
	if err != nil {
		return err
	}

	fmt.Printf("Connected\n")

	node.netInfo.Conn = conn

	return nil
}

// Notify the other node and close the connection
func (node *Node) disconnect() error {
	if node.netInfo.Conn != nil {
		// notify the other node and close the connection
		err := protocol.SendPacket(node.netInfo.Conn, protocol.Packet{
			Header: protocol.HeaderDisconnecting,
		})
		if err != nil {
			return err
		}

		err = node.netInfo.Conn.Close()
		if err != nil {
			return err
		}

		node.state.Stopped = true
	}

	return nil
}

// Wait for a connection on a pre-defined port
func (node *Node) waitForConnection() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", node.netInfo.Port))
	if err != nil {
		return err
	}

	// accept only one conneciton
	connection, err := listener.Accept()
	if err != nil {
		return err
	}

	fmt.Printf("New connection from %s\n", connection.RemoteAddr().String())

	node.netInfo.Conn = connection

	return nil
}

// Starts the node in either sending or receiving state and performs the transfer
func (node *Node) Start() {
	switch node.isSending {
	case true:
		// SENDER

		localIP, err := addr.GetLocal()
		if err != nil {
			panic(err)
		}

		// retrieve information about the file|directory
		var fileToSend *fsys.File
		var dirToSend *fsys.Directory
		switch node.transferInfo.Sending.IsDirectory {
		case true:
			dirToSend, err = fsys.GetDir(node.transferInfo.Sending.ServingPath, node.transferInfo.Sending.Recursive)
			if err != nil {
				panic(err)
			}
		case false:
			fileToSend, err = fsys.GetFile(node.transferInfo.Sending.ServingPath)
			if err != nil {
				panic(err)
			}
		}

		if dirToSend != nil {
			fmt.Printf("Sending \"%s\" (%.2f MB) locally on %s:%d\n", dirToSend.Name, float32(dirToSend.Size)/1024/1024, localIP, node.netInfo.Port)
		} else {
			fmt.Printf("Sending \"%s\" (%.2f MB) locally on %s:%d\n", fileToSend.Name, float32(fileToSend.Size)/1024/1024, localIP, node.netInfo.Port)

		}

		// wain for another node to connect
		err = node.waitForConnection()
		if err != nil {
			panic(err)
		}

		// generate and send encryption key
		encrKey := encryption.Generate32AESkey()
		node.netInfo.EncryptionKey = encrKey
		fmt.Printf("Generated encryption key: %s\n", encrKey)

		err = protocol.SendEncryptionKey(node.netInfo.Conn, encrKey)
		if err != nil {
			panic(err)
		}

		// listen for incoming packets
		go protocol.ReceivePackets(node.netInfo.Conn, node.packetPipe)

		// send info about file/directory
		go protocol.SendTransferOffer(node.netInfo.Conn, fileToSend, dirToSend, node.netInfo.EncryptionKey)

		// mainloop
		for {
			if node.state.Stopped {
				node.disconnect()
				break
			}

			// receive incoming packets and decrypt them if necessary
			incomingPacket, ok := <-node.packetPipe
			if !ok {
				node.state.Stopped = true
			}

			// if encryption key is set - decrypt packet on the spot
			if node.netInfo.EncryptionKey != nil {
				err = incomingPacket.DecryptBody(node.netInfo.EncryptionKey)
				if err != nil {
					panic(err)
				}
			}

			// react based on a header of a received packet
			switch incomingPacket.Header {

			case protocol.HeaderReady:
				// the other node is ready to receive file data
				node.transferInfo.Sending.CanSendBytes = true

			case protocol.HeaderAccept:
				// the receiving node has accepted the transfer
				node.state.AllowedToTransfer = true

				fmt.Printf("Transfer allowed. Sending...\n")

				// notify it about all the files that are going to be sent
				switch node.transferInfo.Sending.IsDirectory {
				case true:
					// send file packets for the files in the directory

					var filesToSend []*fsys.File
					if node.transferInfo.Sending.Recursive {
						filesToSend = dirToSend.GetAllFiles(true)
					} else {
						filesToSend = dirToSend.GetAllFiles(false)
					}

					// notify the other node about all the files that are going to be sent
					for counter, file := range filesToSend {
						// assign ID and add it to the node sendlist

						file.ID = uint64(counter)
						node.transferInfo.Sending.FilesToSend = append(node.transferInfo.Sending.FilesToSend, file)

						filePacket, err := protocol.CreateFilePacket(file)
						if err != nil {
							panic(err)
						}

						// encrypt if necessary
						if node.netInfo.EncryptionKey != nil {
							encryptedBody, err := encryption.Encrypt(node.netInfo.EncryptionKey, filePacket.Body)
							if err != nil {
								panic(err)
							}
							filePacket.Body = encryptedBody
						}

						err = protocol.SendPacket(node.netInfo.Conn, *filePacket)
						if err != nil {
							panic(err)
						}
					}

					// set current file id to the first file
					node.transferInfo.Sending.CurrentFileID = 0

				case false:
					// send a filepacket of a single file
					fileToSend.ID = 0
					node.transferInfo.Sending.FilesToSend = append(node.transferInfo.Sending.FilesToSend, fileToSend)

					filePacket, err := protocol.CreateFilePacket(node.transferInfo.Sending.FilesToSend[0])
					if err != nil {
						panic(err)
					}

					// encrypt if necessary
					if node.netInfo.EncryptionKey != nil {
						encryptedBody, err := encryption.Encrypt(node.netInfo.EncryptionKey, filePacket.Body)
						if err != nil {
							panic(err)
						}
						filePacket.Body = encryptedBody
					}

					err = protocol.SendPacket(node.netInfo.Conn, *filePacket)
					if err != nil {
						panic(err)
					}

					// set current file id to the first and only file
					node.transferInfo.Sending.CurrentFileID = 0
				}

			case protocol.HeaderReject:
				node.state.Stopped = true

				fmt.Printf("Transfer rejected. Disconnecting...")

			case protocol.HeaderDisconnecting:
				node.state.Stopped = true
				fmt.Printf("%s disconnected\n", node.netInfo.Conn.RemoteAddr())
			}

			// if allowed to transfer and the other node is ready to receive packets - send one piece
			// and wait for it to be ready again
			if node.state.AllowedToTransfer && node.transferInfo.Sending.CanSendBytes {
				// handle a single file or a directory
				switch node.transferInfo.Sending.IsDirectory {
				case true:
					// sending a file in a directory (possibly more than 1)
					if !node.transferInfo.Sending.Recursive {

					} else {
						// send bytes of all files one by one

					}

				case false:
					// sending a piece of a single file
					currentFileID := node.transferInfo.Sending.CurrentFileID
					err = protocol.SendPiece(node.transferInfo.Sending.FilesToSend[currentFileID], node.netInfo.Conn, node.netInfo.EncryptionKey)
					switch err {
					case protocol.ErrorSentAll:
						// the file has been sent fully
						fileIDBuff := new(bytes.Buffer)
						err = binary.Write(fileIDBuff, binary.BigEndian, node.transferInfo.Sending.FilesToSend[currentFileID].ID)
						if err != nil {
							panic(err)
						}

						endFilePacket := protocol.Packet{
							Header: protocol.HeaderEndfile,
							Body:   fileIDBuff.Bytes(),
						}

						if node.netInfo.EncryptionKey != nil {
							err = endFilePacket.EncryptBody(node.netInfo.EncryptionKey)
							if err != nil {
								panic(err)
							}
						}

						protocol.SendPacket(node.netInfo.Conn, endFilePacket)

						// as only one file has been requested to send - there`s nothing else to do
						// sending DONE packet
						protocol.SendPacket(node.netInfo.Conn, protocol.Packet{
							Header: protocol.HeaderDone,
						})

						fmt.Printf("Transfer ended successfully\n")

						node.state.Stopped = true

					case nil:
						break

					default:
						node.state.Stopped = true

						currentFileID := node.transferInfo.Sending.CurrentFileID
						fmt.Printf("An error occured while sending a piece of \"%s\": %s\n", node.transferInfo.Sending.FilesToSend[currentFileID].Name, err)
						panic(err)
					}

					node.transferInfo.Sending.CanSendBytes = false
				}

			}
		}

	case false:
		// RECEIVER

		// connect to the sending node
		err := node.connect()
		if err != nil {
			fmt.Printf("Could not connect to %s:%d\n", node.netInfo.ConnAddr, node.netInfo.Port)
			os.Exit(-1)
		}

		// listen for incoming packets
		go protocol.ReceivePackets(node.netInfo.Conn, node.packetPipe)

		// mainloop
		for {
			node.mutex.Lock()
			stopped := node.state.Stopped
			node.mutex.Unlock()

			if stopped {
				node.disconnect()
				break
			}

			// receive incoming packets and decrypt them if necessary
			incomingPacket, ok := <-node.packetPipe
			if !ok {
				break
			}

			// if encryption key is set - decrypt packet on the spot
			if node.netInfo.EncryptionKey != nil {
				err = incomingPacket.DecryptBody(node.netInfo.EncryptionKey)
				if err != nil {
					panic(err)
				}
			}

			// react based on a header of a received packet
			switch incomingPacket.Header {

			case protocol.HeaderTransferOffer:
				// accept of reject offer
				go func() {
					file, dir, err := protocol.DecodeTransferPacket(incomingPacket)
					if err != nil {
						panic(err)
					}

					if file != nil {
						fmt.Printf("\n| Filename: %s\n| Size: %.2f MB\n| Checksum: %s\n", file.Name, float32(file.Size)/1024/1024, file.Checksum)
					} else if dir != nil {
						fmt.Printf("\n| Directory name: %s\n| Size: %.2f MB\n", dir.Name, float32(dir.Size)/1024/1024)
					}

					var answer string
					fmt.Printf("| Download ? [Y/n]: ")
					fmt.Scanln(&answer)
					fmt.Printf("\n\n")

					if strings.EqualFold(answer, "y") || answer == "" {
						// yes

						// send aceptance packet

						acceptancePacket := protocol.Packet{
							Header: protocol.HeaderAccept,
						}

						err = protocol.SendPacket(node.netInfo.Conn, acceptancePacket)
						if err != nil {
							panic(err)
						}

						// notify the node that we`re ready to transportation. No need
						// for encryption because the body is nil
						err = protocol.SendPacket(node.netInfo.Conn, protocol.Packet{
							Header: protocol.HeaderReady,
						})
						if err != nil {
							panic(err)
						}

					} else {
						// no

						rejectionPacket := protocol.Packet{
							Header: protocol.HeaderReject,
						}

						err = protocol.SendPacket(node.netInfo.Conn, rejectionPacket)
						if err != nil {
							panic(err)
						}

						node.mutex.Lock()
						node.state.Stopped = true
						node.mutex.Unlock()
					}
				}()

			case protocol.HeaderFile:
				// add file to the accepted files;

				file, err := protocol.DecodeFilePacket(incomingPacket)
				if err != nil {
					panic(err)
				}
				fullFilePath := filepath.Join(node.transferInfo.Receiving.DownloadsPath, file.Name)

				// check if the file already exists; if yes - remove it and replace with a new one
				_, err = os.Stat(fullFilePath)
				if err == nil {
					// exists
					// remove it
					os.Remove(fullFilePath)
				}

				file.Path = fullFilePath
				file.Open()

				node.mutex.Lock()
				node.transferInfo.Receiving.AcceptedFiles = append(node.transferInfo.Receiving.AcceptedFiles, file)
				node.mutex.Unlock()

			case protocol.HeaderDirectory:
				// directory
				dir, err := protocol.DecodeDirectoryPacket(incomingPacket)
				if err != nil {
					panic(err)
				}

				// add a new directory to downloads path
				node.transferInfo.Receiving.DownloadsPath = filepath.Join(node.transferInfo.Receiving.DownloadsPath, dir.Name)
				err = os.MkdirAll(node.transferInfo.Receiving.DownloadsPath, os.ModePerm)
				if err != nil {
					panic(err)
				}

			case protocol.HeaderFileBytes:
				// check if this file has been accepted to receive
				fileIDReader := bytes.NewReader(incomingPacket.Body)
				var fileID uint64
				err := binary.Read(fileIDReader, binary.BigEndian, &fileID)
				if err != nil {
					panic(err)
				}

				for _, acceptedFile := range node.transferInfo.Receiving.AcceptedFiles {
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
				err = protocol.SendPacket(node.netInfo.Conn, protocol.Packet{
					Header: protocol.HeaderReady,
				})
				if err != nil {
					panic(err)
				}

			case protocol.HeaderEndfile:
				// one of the files has been received completely,
				// compare checksums and check if it is the last
				// file in the transfer

				fileIDReader := bytes.NewReader(incomingPacket.Body)
				var fileID uint64
				err := binary.Read(fileIDReader, binary.BigEndian, &fileID)
				if err != nil {
					panic(err)
				}

				for index, acceptedFile := range node.transferInfo.Receiving.AcceptedFiles {
					if acceptedFile.ID == fileID {
						// accepted

						// close the handler afterwards
						defer acceptedFile.Handler.Close()

						// remove this file from the pool
						node.transferInfo.Receiving.AcceptedFiles = append(node.transferInfo.Receiving.AcceptedFiles[:index], node.transferInfo.Receiving.AcceptedFiles[index+1:]...)

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

				// node.state.Stopped = true

			case protocol.HeaderEncryptionKey:
				// retrieve the key
				packetReader := bytes.NewReader(incomingPacket.Body)

				var keySize uint64
				binary.Read(packetReader, binary.BigEndian, &keySize)

				encrKey := make([]byte, keySize)
				packetReader.Read(encrKey)

				node.netInfo.EncryptionKey = encrKey

				fmt.Printf("Got an encryption key: %s\n", encrKey)

			case protocol.HeaderDone:
				node.mutex.Lock()
				node.state.Stopped = true
				node.mutex.Unlock()

			case protocol.HeaderDisconnecting:
				node.mutex.Lock()
				node.state.Stopped = true
				node.mutex.Unlock()

				fmt.Printf("%s disconnected\n", node.netInfo.Conn.RemoteAddr())
			}
		}

	}
}
