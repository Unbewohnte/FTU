/*
ftu - file transferring utility.
Copyright (C) 2021  Kasyanov Nikolay Alexeevich (Unbewohnte (https://unbewohnte.xyz/))

This file is a part of ftu

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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
type netInfo struct {
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
	verboseOutput bool
	mutex         *sync.Mutex
	packetPipe    chan *protocol.Packet // a way to receive incoming packets from another goroutine
	isSending     bool                  // sending or a receiving node
	netInfo       *netInfo
	state         *nodeInnerstates
	transferInfo  *transferInfo
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
		var err error
		options.ClientSide.DownloadsFolderPath, err = filepath.Abs(options.ClientSide.DownloadsFolderPath)
		if err != nil {
			return nil, err
		}

		err = os.MkdirAll(options.ClientSide.DownloadsFolderPath, os.ModePerm)
		if err != nil {
			return nil, err
		}

	}

	node := Node{
		verboseOutput: options.VerboseOutput,
		mutex:         &sync.Mutex{},
		packetPipe:    make(chan *protocol.Packet, 100),
		isSending:     options.IsSending,
		netInfo: &netInfo{
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

	fmt.Printf("\nConnecting to %s:%d...", node.netInfo.ConnAddr, node.netInfo.Port)

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", node.netInfo.ConnAddr, node.netInfo.Port), time.Second*5)
	if err != nil {
		return err
	}

	fmt.Printf("\nConnected")

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

	fmt.Printf("\nNew connection from %s", connection.RemoteAddr().String())

	node.netInfo.Conn = connection

	return nil
}

// Prints information about the transfer after defined delay
func (node *Node) printTransferInfo(delay time.Duration) error {
	time.Sleep(delay)

	switch node.isSending {
	case true:
		if !node.state.AllowedToTransfer {
			break
		}
		fmt.Printf("\r| files(s) left to send: %4d", len(node.transferInfo.Sending.FilesToSend))

	case false:
		if len(node.transferInfo.Receiving.AcceptedFiles) <= 0 {
			break
		}
		fmt.Printf("\r| file(s) left to receive: %4d", len(node.transferInfo.Receiving.AcceptedFiles))
	}
	return nil
}

// Starts the node in either sending or receiving state and performs the transfer
func (node *Node) Start() {
	switch node.isSending {
	case true:
		// SENDER NODE

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
			size := float32(dirToSend.Size) / 1024 / 1024
			sizeLevel := "MiB"
			if size >= 1024 {
				// GiB
				size = size / 1024
				sizeLevel = "GiB"
			}

			fmt.Printf("\nSending \"%s\" (%.3f %s) locally on %s:%d", dirToSend.Name, size, sizeLevel, localIP, node.netInfo.Port)
		} else {
			size := float32(fileToSend.Size) / 1024 / 1024
			sizeLevel := "MiB"
			if size >= 1024 {
				// GiB
				size = size / 1024
				sizeLevel = "GiB"
			}
			fmt.Printf("\nSending \"%s\" (%.3f %s) locally on %s:%d and remotely", fileToSend.Name, size, sizeLevel, localIP, node.netInfo.Port)

		}

		// wain for another node to connect
		err = node.waitForConnection()
		if err != nil {
			panic(err)
		}

		// generate and send encryption key
		encrKey := encryption.Generate32AESkey()
		node.netInfo.EncryptionKey = encrKey
		fmt.Printf("\nGenerated encryption key: %s\n", encrKey)

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
				fmt.Printf("\n")
				node.disconnect()
				break
			}

			// receive incoming packets and decrypt them if necessary
			incomingPacket, ok := <-node.packetPipe
			if !ok {
				fmt.Printf("\nThe connection has been closed unexpectedly\n")
				os.Exit(-1.)
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

				fmt.Printf("\nTransfer allowed. Sending...")

				// notify it about all the files that are going to be sent
				switch node.transferInfo.Sending.IsDirectory {
				case true:
					// send file packets for the files in the directory

					err = dirToSend.SetRelativePaths(dirToSend.Path, node.transferInfo.Sending.Recursive)
					if err != nil {
						panic(err)
					}
					filesToSend := dirToSend.GetAllFiles(node.transferInfo.Sending.Recursive)

					// notify the other node about all the files that are going to be sent
					for counter, file := range filesToSend {
						// assign ID and add it to the node sendlist

						file.ID = uint64(counter)
						node.transferInfo.Sending.FilesToSend = append(node.transferInfo.Sending.FilesToSend, file)

						// set current file id to the first file
						node.transferInfo.Sending.CurrentFileID = 0

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

						if node.verboseOutput {
							fmt.Printf("\n[File] Sent filepacket for \"%s\"", file.Name)
						}

						time.Sleep(time.Microsecond * 3)
					}

					filesInfoDonePacket := protocol.Packet{
						Header: protocol.HeaderFilesInfoDone,
					}
					protocol.SendPacket(node.netInfo.Conn, filesInfoDonePacket)

					if node.verboseOutput {
						fmt.Printf("\n[File] Done sending filepackets")
					}

				case false:
					// send a filepacket of a single file
					fileToSend.ID = 0
					node.transferInfo.Sending.FilesToSend = append(node.transferInfo.Sending.FilesToSend, fileToSend)

					// set current file index to the first and only file
					node.transferInfo.Sending.CurrentFileID = 0

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

					filesInfoDonePacket := protocol.Packet{
						Header: protocol.HeaderFilesInfoDone,
					}
					protocol.SendPacket(node.netInfo.Conn, filesInfoDonePacket)

					if node.verboseOutput {
						fmt.Printf("\n[File] Sent filepacket for \"%s\"", fileToSend.Name)
					}

					if node.verboseOutput {
						fmt.Printf("\n[File] Done sending filepackets")
					}

				}

			case protocol.HeaderReject:
				node.state.Stopped = true
				fmt.Printf("\nTransfer rejected. Disconnecting...")

			case protocol.HeaderDisconnecting:
				node.state.Stopped = true
				fmt.Printf("\n%s disconnected", node.netInfo.Conn.RemoteAddr())

			case protocol.HeaderAlreadyHave:
				// the other node already has a file with such ID.
				// do not send it

				fileIDReader := bytes.NewReader(incomingPacket.Body)
				var fileID uint64
				binary.Read(fileIDReader, binary.BigEndian, &fileID)

				for index, fileToSend := range node.transferInfo.Sending.FilesToSend {
					if fileToSend.ID == fileID {
						node.transferInfo.Sending.FilesToSend = append(node.transferInfo.Sending.FilesToSend[:index], node.transferInfo.Sending.FilesToSend[index+1:]...)

						node.transferInfo.Sending.CurrentFileID++

						if node.verboseOutput {
							fmt.Printf("\n[File] receiver already has \"%s\"", fileToSend.Name)
						}
					}
				}

			}

			if !node.verboseOutput {
				go node.printTransferInfo(time.Second)
			}

			// Transfer section

			if len(node.transferInfo.Sending.FilesToSend) == 0 {
				// if there`s nothing else to send - create and send DONE packet
				protocol.SendPacket(node.netInfo.Conn, protocol.Packet{
					Header: protocol.HeaderDone,
				})

				fmt.Printf("\nTransfer ended successfully")
				node.state.Stopped = true

				continue
			}

			// if allowed to transfer and the other node is ready to receive packets - send one piece
			// and wait for it to be ready again
			if node.state.AllowedToTransfer && node.transferInfo.Sending.CanSendBytes {
				// sending a piece of a single file

				// determine an index of a file with current ID
				var currentFileIndex uint64 = 0
				for index, fileToSend := range node.transferInfo.Sending.FilesToSend {
					if fileToSend.ID == node.transferInfo.Sending.CurrentFileID {
						currentFileIndex = uint64(index)
						break
					}
				}

				err = protocol.SendPiece(node.transferInfo.Sending.FilesToSend[currentFileIndex], node.netInfo.Conn, node.netInfo.EncryptionKey)
				switch err {
				case protocol.ErrorSentAll:
					// the file has been sent fully

					if node.verboseOutput {
						fmt.Printf("\n[File] fully sent \"%s\" -- %d bytes", node.transferInfo.Sending.FilesToSend[currentFileIndex].Name, node.transferInfo.Sending.FilesToSend[currentFileIndex].Size)
					}

					fileIDBuff := new(bytes.Buffer)
					err = binary.Write(fileIDBuff, binary.BigEndian, node.transferInfo.Sending.FilesToSend[currentFileIndex].ID)
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

					node.transferInfo.Sending.FilesToSend = append(node.transferInfo.Sending.FilesToSend[:currentFileIndex], node.transferInfo.Sending.FilesToSend[currentFileIndex+1:]...)

					// start sending the next file
					node.transferInfo.Sending.CurrentFileID++

				case nil:
					node.transferInfo.Sending.CanSendBytes = false

				default:
					node.state.Stopped = true

					fmt.Printf("\nAn error occured while sending a piece of \"%s\": %s", node.transferInfo.Sending.FilesToSend[currentFileIndex].Name, err)
					panic(err)
				}
			}
		}

	case false:
		// RECEIVER NODE

		// connect to the sending node
		err := node.connect()
		if err != nil {
			fmt.Printf("\nCould not connect to %s:%d", node.netInfo.ConnAddr, node.netInfo.Port)
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
				fmt.Printf("\n")
				node.disconnect()
				break
			}

			// receive incoming packets and decrypt them if necessary
			incomingPacket, ok := <-node.packetPipe
			if !ok {
				fmt.Printf("\nThe connection has been closed unexpectedly\n")
				os.Exit(-1)
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
						size := float32(file.Size) / 1024 / 1024
						sizeLevel := "MiB"
						if size >= 1024 {
							// GiB
							size = size / 1024
							sizeLevel = "GiB"
						}
						fmt.Printf("\n| Filename: %s\n| Size: %.3f %s\n| Checksum: %s\n", file.Name, size, sizeLevel, file.Checksum)
					} else if dir != nil {
						size := float32(dir.Size) / 1024 / 1024
						sizeLevel := "MiB"
						if size >= 1024 {
							// GiB
							size = size / 1024
							sizeLevel = "GiB"
						}
						fmt.Printf("\n| Directory name: %s\n| Size: %.3f %s\n", dir.Name, size, sizeLevel)
					}

					var answer string
					fmt.Printf("| Download ? [Y/n]: ")
					fmt.Scanln(&answer)
					fmt.Printf("\n\n")

					if strings.EqualFold(answer, "y") || answer == "" {
						// yes

						// in case it`s a directory - create it now
						if dir != nil {
							err = os.MkdirAll(filepath.Join(node.transferInfo.Receiving.DownloadsPath, dir.Name), os.ModePerm)
							if err != nil {
								// well, just download all files in the default downloads folder then
								fmt.Printf("\n[ERROR] could not create a directory")
							} else {
								// also download everything in a newly created directory
								node.transferInfo.Receiving.DownloadsPath = filepath.Join(node.transferInfo.Receiving.DownloadsPath, dir.Name)
							}

						}

						// send aceptance packet
						acceptancePacket := protocol.Packet{
							Header: protocol.HeaderAccept,
						}

						err = protocol.SendPacket(node.netInfo.Conn, acceptancePacket)
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

				if node.verboseOutput {
					fmt.Printf("\n[File] Received info on \"%s\" - %d bytes", file.Name, file.Size)
				}

				if strings.TrimSpace(file.RelativeParentPath) == "" {
					// does not have a parent dir
					file.Path = filepath.Join(node.transferInfo.Receiving.DownloadsPath, file.Name)
				} else {
					file.Path = filepath.Join(node.transferInfo.Receiving.DownloadsPath, file.RelativeParentPath)
				}

				// create all underlying directories right ahead
				err = os.MkdirAll(filepath.Dir(file.Path), os.ModePerm)
				if err != nil {
					panic(err)
				}

				// check if the file already exists
				_, err = os.Stat(file.Path)
				if err == nil {
					// exists
					// check if it is the exact file
					existingFileHandler, err := os.Open(file.Path)
					if err != nil {
						panic(err)
					}

					existingFileChecksum, err := checksum.GetPartialCheckSum(existingFileHandler)
					if err != nil {
						panic(err)
					}

					if existingFileChecksum == file.Checksum {
						// it`s the exact same file. No need to receive it again
						// notify the other node

						alreadyHavePacketBodyBuffer := new(bytes.Buffer)
						binary.Write(alreadyHavePacketBodyBuffer, binary.BigEndian, file.ID)

						alreadyHavePacket := protocol.Packet{
							Header: protocol.HeaderAlreadyHave,
							Body:   alreadyHavePacketBodyBuffer.Bytes(),
						}

						if node.netInfo.EncryptionKey != nil {
							encryptedBody, err := encryption.Encrypt(node.netInfo.EncryptionKey, alreadyHavePacket.Body)
							if err != nil {
								panic(err)
							}
							alreadyHavePacket.Body = encryptedBody
						}

						protocol.SendPacket(node.netInfo.Conn, alreadyHavePacket)

						if node.verboseOutput {
							fmt.Printf("\n[File] already have \"%s\"", file.Name)
						}

					} else {
						// not the same file. Remove it and await new bytes
						os.Remove(file.Path)
					}

					existingFileHandler.Close()
				} else {
					// does not exist

					node.mutex.Lock()
					node.transferInfo.Receiving.AcceptedFiles = append(node.transferInfo.Receiving.AcceptedFiles, file)
					node.mutex.Unlock()
				}

			case protocol.HeaderFileBytes:
				// check if this file has been accepted to receive

				fileBytesBuffer := bytes.NewBuffer(incomingPacket.Body)

				var fileID uint64
				err := binary.Read(fileBytesBuffer, binary.BigEndian, &fileID)
				if err != nil {
					panic(err)
				}

				for _, acceptedFile := range node.transferInfo.Receiving.AcceptedFiles {
					if acceptedFile.ID == fileID {
						// accepted

						// append provided bytes to the file

						err = acceptedFile.Open()
						if err != nil {
							panic(err)
						}

						fileBytes := fileBytesBuffer.Bytes()

						wrote, err := acceptedFile.Handler.WriteAt(fileBytes, int64(acceptedFile.SentBytes))
						if err != nil {
							panic(err)
						}
						acceptedFile.SentBytes += uint64(wrote)

						err = acceptedFile.Close()
						if err != nil {
							panic(err)
						}
					}
				}

				readyPacket := protocol.Packet{
					Header: protocol.HeaderReady,
				}
				protocol.SendPacket(node.netInfo.Conn, readyPacket)

			case protocol.HeaderFilesInfoDone:
				// have all information about the files

				// notify the other node that this one is ready
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

						if node.verboseOutput {
							fmt.Printf("\n[File] fully received \"%s\" -- %d bytes", acceptedFile.Name, acceptedFile.Size)
						}

						err = acceptedFile.Open()
						if err != nil {
							panic(err)
						}

						// remove this file from the pool
						node.transferInfo.Receiving.AcceptedFiles = append(node.transferInfo.Receiving.AcceptedFiles[:index], node.transferInfo.Receiving.AcceptedFiles[index+1:]...)

						// compare checksums
						realChecksum, err := checksum.GetPartialCheckSum(acceptedFile.Handler)
						if err != nil {
							panic(err)
						}

						if realChecksum != acceptedFile.Checksum {
							fmt.Printf("\n| \"%s\" is corrupted", acceptedFile.Name)
							acceptedFile.Close()
							break
						} else {
							acceptedFile.Close()
							break
						}
					}
				}

				err = protocol.SendPacket(node.netInfo.Conn, protocol.Packet{
					Header: protocol.HeaderReady,
				})
				if err != nil {
					panic(err)
				}

			case protocol.HeaderEncryptionKey:
				// retrieve the key
				packetReader := bytes.NewReader(incomingPacket.Body)

				var keySize uint64
				binary.Read(packetReader, binary.BigEndian, &keySize)

				encrKey := make([]byte, keySize)
				packetReader.Read(encrKey)

				node.netInfo.EncryptionKey = encrKey

				fmt.Printf("\nGot an encryption key: %s", encrKey)

			case protocol.HeaderDone:
				node.mutex.Lock()
				node.state.Stopped = true
				node.mutex.Unlock()

			case protocol.HeaderDisconnecting:
				node.mutex.Lock()
				node.state.Stopped = true
				node.mutex.Unlock()

				fmt.Printf("\n%s disconnected", node.netInfo.Conn.RemoteAddr())
			}

			if !node.verboseOutput {
				go node.printTransferInfo(time.Second)
			}
		}
	}
}
