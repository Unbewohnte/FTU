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
	"github.com/Unbewohnte/ftu/fsys"
	"github.com/Unbewohnte/ftu/protocol"
)

type NodeInnerStates struct {
	Stopped           bool
	Connected         bool
	AllowedToTransfer bool
}

type Net struct {
	ConnAddr      string
	Conn          net.Conn
	Port          uint
	EncryptionKey []byte
}

type TransferInfo struct {
	Ready         bool   // is the other node ready to receive another piece
	ServingPath   string // path to the thing that will be sent
	Recursive     bool
	AcceptedFiles []*fsys.File
	DownloadsPath string
}

// Sender and receiver in one type !
type Node struct {
	PacketPipe   chan *protocol.Packet
	Mutex        *sync.Mutex
	IsSending    bool
	Net          *Net
	State        *NodeInnerStates
	TransferInfo *TransferInfo
}

// Creates a new node
func NewNode(options *NodeOptions) (*Node, error) {
	mutex := new(sync.Mutex)

	node := Node{
		PacketPipe: make(chan *protocol.Packet, 100),
		Mutex:      mutex,
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
			Connected:         false,
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

func (node *Node) connect(addr string, port uint) error {
	if port == 0 {
		port = node.Net.Port
	}

	fmt.Printf("Connecting to %s:%d...\n", addr, port)

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", addr, port), time.Second*5)
	if err != nil {
		return err
	}

	fmt.Printf("Connected\n")

	node.Net.Conn = conn
	node.State.Connected = true

	return nil
}

func (node *Node) disconnect() error {
	if node.State.Connected && node.Net.Conn != nil {
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
		node.State.Connected = false
	}

	return nil
}

// Waits for connection on a pre-defined port
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
	node.State.Connected = true

	return nil
}

// Starts the node in either sending or receiving state and performs the transfer
func (node *Node) Start() {
	switch node.IsSending {
	case true:
		// SENDER

		localIP, err := addr.GetLocalIP()
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

		// listen for incoming packets
		go receivePackets(node.Net.Conn, node.PacketPipe)

		// send fileoffer
		go sendFilePacket(node.Net.Conn, file)

		// mainloop
		for {
			node.Mutex.Lock()
			stopped := node.State.Stopped
			node.Mutex.Unlock()

			if stopped {
				node.Mutex.Lock()
				node.disconnect()
				node.Mutex.Unlock()
				break
			}

			incomingPacket := <-node.PacketPipe

			switch incomingPacket.Header {
			case protocol.HeaderReady:
				node.Mutex.Lock()
				node.TransferInfo.Ready = true
				node.Mutex.Unlock()

			case protocol.HeaderAccept:
				node.Mutex.Lock()
				node.State.AllowedToTransfer = true
				node.Mutex.Unlock()
				go fmt.Printf("Transfer allowed. Sending...\n")

			case protocol.HeaderDisconnecting:
				node.Mutex.Lock()
				node.State.Stopped = true
				node.Mutex.Unlock()
				go fmt.Printf("%s disconnected\n", node.Net.Conn.RemoteAddr())

			case protocol.HeaderReject:
				node.Mutex.Lock()
				node.State.Stopped = true
				node.Mutex.Unlock()
				go fmt.Printf("Transfer rejected. Disconnecting...")
			}

			if node.State.AllowedToTransfer {
				err = sendPiece(file, node.Net.Conn)
				if err != nil {
					if err == ErrorSentAll {
						// the file has been sent fully
						fileIDBuff := new(bytes.Buffer)
						err = binary.Write(fileIDBuff, binary.BigEndian, file.ID)
						if err != nil {
							node.Mutex.Lock()
							node.State.Stopped = true
							node.Mutex.Unlock()
						}

						protocol.SendPacket(node.Net.Conn, protocol.Packet{
							Header: protocol.HeaderEndfile,
							Body:   fileIDBuff.Bytes(),
						})

						node.Mutex.Lock()
						node.State.Stopped = true
						node.Mutex.Unlock()
					} else {
						node.Mutex.Lock()
						node.State.Stopped = true
						node.Mutex.Unlock()

						fmt.Printf("An error occured when sending a piece of \"%s\": %s\n", file.Name, err)
						panic(err)
					}
				}
			}
		}

	case false:
		// RECEIVER

		// connect to the sending node
		err := node.connect(node.Net.ConnAddr, node.Net.Port)
		if err != nil {
			panic(err)
		}

		// listen for incoming packets
		go receivePackets(node.Net.Conn, node.PacketPipe)

		// mainloop
		for {
			node.Mutex.Lock()
			stopped := node.State.Stopped
			node.Mutex.Unlock()

			if stopped {
				node.Mutex.Lock()
				node.disconnect()
				node.Mutex.Unlock()
				break
			}

			incomingPacket, ok := <-node.PacketPipe
			if !ok {
				break
			}

			switch incomingPacket.Header {
			case protocol.HeaderFile:
				go func() {
					file, err := decodeFilePacket(incomingPacket)
					if err != nil {
						panic(err)
					}

					fmt.Printf("| ID: %d\n| Filename: %s\n| Size: %.2f MB\n| Checksum: %s\n", file.ID, file.Name, float32(file.Size)/1024/1024, file.Checksum)
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

						node.Mutex.Lock()
						node.TransferInfo.AcceptedFiles = append(node.TransferInfo.AcceptedFiles, file)
						node.Mutex.Unlock()

						// notify the node that we`re ready to transportation
						err = protocol.SendPacket(node.Net.Conn, protocol.Packet{
							Header: protocol.HeaderReady,
						})
						if err != nil {
							panic(err)
						}

						// send aceptance packet
						protocol.SendPacket(node.Net.Conn, protocol.Packet{
							Header: protocol.HeaderAccept,
							Body:   responsePacketFileIDBuffer.Bytes(),
						})

					} else {
						// no
						err = protocol.SendPacket(node.Net.Conn, protocol.Packet{
							Header: protocol.HeaderReject,
							Body:   responsePacketFileIDBuffer.Bytes(),
						})
						if err != nil {
							panic(err)
						}

						node.Mutex.Lock()
						node.State.Stopped = true
						node.Mutex.Unlock()
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

				node.Mutex.Lock()
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
				node.Mutex.Unlock()

				err = protocol.SendPacket(node.Net.Conn, protocol.Packet{
					Header: protocol.HeaderReady,
				})
				if err != nil {
					panic(err)
				}

			case protocol.HeaderEndfile:
				fileIDReader := bytes.NewReader(incomingPacket.Body)
				var fileID uint64
				err := binary.Read(fileIDReader, binary.BigEndian, &fileID)
				if err != nil {
					panic(err)
				}

				node.Mutex.Lock()
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

				node.State.Stopped = true
				node.Mutex.Unlock()

			case protocol.HeaderDisconnecting:
				node.Mutex.Lock()
				node.State.Stopped = true
				node.Mutex.Unlock()

				go fmt.Printf("%s disconnected\n", node.Net.Conn.RemoteAddr())
			}
		}

	}
}
