/*
ftu - file transferring utility.
Copyright (C) 2021,2022  Kasyanov Nikolay Alexeevich (Unbewohnte (https://unbewohnte.xyz/))

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

// netInfowork specific settings
type netInfo struct {
	ConnAddr      string   // address to connect to. Does not include port
	Conn          net.Conn // the core TCP connection of the node. Self-explanatory
	Port          uint     // a port to connect to/listen on
	EncryptionKey []byte   // if != nil - incoming packets will be decrypted with it and outcoming packets will be encrypted
}

// Sending-side node information
type sending struct {
	ServingPath         string // path to the thing that will be sent
	IsDirectory         bool   // is ServingPath a directory
	Recursive           bool   // recursively send directory
	CanSendBytes        bool   // is the other node ready to receive another piece
	AllowedToTransfer   bool   // the way to notify the mainloop of a sending node to start sending pieces of files
	InTransfer          bool   // already transferring|receiving files
	FilesToSend         []*fsys.File
	SymlinksToSend      []*fsys.Symlink
	CurrentFileID       uint64 // an id of a file that is currently being transported
	SentBytes           uint64 // how many bytes sent already
	TotalTransferSize   uint64 // how many bytes will be sent in total
	CurrentSymlinkIndex uint64 // current index of a symlink that is
}

// Receiving-side node information
type receiving struct {
	AcceptedFiles     []*fsys.File // files that`ve been accepted to be received
	DownloadsPath     string       // where to download
	TotalDownloadSize uint64       // how many bytes will be received in total
	ReceivedBytes     uint64       // how many bytes downloaded so far
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
	stopped       bool                  // the way to exit the mainloop in case of an external error or a successful end of a transfer
	netInfo       *netInfo
	transferInfo  *transferInfo
}

// Creates a new either a sending or receiving node with specified options
func NewNode(options *NodeOptions) (*Node, error) {
	var isDir bool
	if options.IsSending {
		// sending node preparation
		sendingPathStats, err := os.Stat(options.SenderSide.ServingPath)
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
		options.ReceiverSide.DownloadsFolderPath, err = filepath.Abs(options.ReceiverSide.DownloadsFolderPath)
		if err != nil {
			return nil, err
		}

		err = os.MkdirAll(options.ReceiverSide.DownloadsFolderPath, os.ModePerm)
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
			ConnAddr:      options.ReceiverSide.ConnectionAddr,
			EncryptionKey: nil,
			Conn:          nil,
		},
		stopped: false,
		transferInfo: &transferInfo{
			Sending: &sending{
				ServingPath:       options.SenderSide.ServingPath,
				Recursive:         options.SenderSide.Recursive,
				IsDirectory:       isDir,
				TotalTransferSize: 0,
				SentBytes:         0,
			},
			Receiving: &receiving{
				AcceptedFiles:     nil,
				DownloadsPath:     options.ReceiverSide.DownloadsFolderPath,
				ReceivedBytes:     0,
				TotalDownloadSize: 0,
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

		node.stopped = true
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
		if !node.transferInfo.Sending.AllowedToTransfer {
			// do not print if the transfer has not been accepted yet
			break
		}
		fmt.Printf("\r| (%.2f/%.2f MB)",
			float32(node.transferInfo.Sending.SentBytes)/1024/1024,
			float32(node.transferInfo.Sending.TotalTransferSize)/1024/1024,
		)

	case false:
		fmt.Printf("\r| (%.2f/%.2f MB)",
			float32(node.transferInfo.Receiving.ReceivedBytes)/1024/1024,
			float32(node.transferInfo.Receiving.TotalDownloadSize)/1024/1024,
		)
	}
	return nil
}

func (node *Node) send() {
	// SENDER NODE

	localIP, err := addr.GetLocal()
	if err != nil {
		panic(err)
	}

	// retrieve information about the file|directory
	var FILETOSEND *fsys.File
	var DIRTOSEND *fsys.Directory
	switch node.transferInfo.Sending.IsDirectory {
	case true:
		DIRTOSEND, err = fsys.GetDir(node.transferInfo.Sending.ServingPath, node.transferInfo.Sending.Recursive)
		if err != nil {
			panic(err)
		}
	case false:
		FILETOSEND, err = fsys.GetFile(node.transferInfo.Sending.ServingPath)
		if err != nil {
			panic(err)
		}
	}

	if DIRTOSEND != nil {
		node.transferInfo.Sending.TotalTransferSize = DIRTOSEND.Size

		displaySize := float32(DIRTOSEND.Size) / 1024 / 1024
		sizeLevel := "MiB"
		if displaySize >= 1024 {
			// GiB
			displaySize = displaySize / 1024
			sizeLevel = "GiB"
		}

		fmt.Printf("\nSending \"%s\" (%.3f %s) locally on %s:%d and remotely (if configured)", DIRTOSEND.Name, displaySize, sizeLevel, localIP, node.netInfo.Port)
	} else {
		node.transferInfo.Sending.TotalTransferSize = FILETOSEND.Size

		displaySize := float32(FILETOSEND.Size) / 1024 / 1024
		sizeLevel := "MiB"
		if displaySize >= 1024 {
			// GiB
			displaySize = displaySize / 1024
			sizeLevel = "GiB"
		}
		fmt.Printf("\nSending \"%s\" (%.3f %s) locally on %s:%d and remotely (if configured)", FILETOSEND.Name, displaySize, sizeLevel, localIP, node.netInfo.Port)

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
	go protocol.SendTransferOffer(node.netInfo.Conn, FILETOSEND, DIRTOSEND, node.netInfo.EncryptionKey)

	// mainloop
	for {
		if node.stopped {
			fmt.Printf("\n")
			node.disconnect()
			break
		}

		if !node.verboseOutput {
			go node.printTransferInfo(time.Second)
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
			node.transferInfo.Sending.AllowedToTransfer = true

			// prepare files to send
			switch node.transferInfo.Sending.IsDirectory {
			case true:
				// send file packets for the files in the directory

				err = DIRTOSEND.SetRelativePaths(DIRTOSEND.Path, node.transferInfo.Sending.Recursive)
				if err != nil {
					panic(err)
				}
				filesToSend := DIRTOSEND.GetAllFiles(node.transferInfo.Sending.Recursive)
				symlinksToSend := DIRTOSEND.GetAllSymlinks(node.transferInfo.Sending.Recursive)

				node.transferInfo.Sending.SymlinksToSend = symlinksToSend

				for counter, file := range filesToSend {
					// assign ID and add it to the node sendlist
					file.ID = uint64(counter)
					node.transferInfo.Sending.FilesToSend = append(node.transferInfo.Sending.FilesToSend, file)
				}

				// set current file id to the first file
				node.transferInfo.Sending.CurrentFileID = 0

			case false:
				FILETOSEND.ID = 0
				node.transferInfo.Sending.FilesToSend = append(node.transferInfo.Sending.FilesToSend, FILETOSEND)

				// set current file index to the first and only file
				node.transferInfo.Sending.CurrentFileID = 0
			}
			fmt.Printf("\n")

		case protocol.HeaderReject:
			node.stopped = true
			fmt.Printf("\nTransfer rejected. Disconnecting...")

		case protocol.HeaderDisconnecting:
			node.stopped = true
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

					node.transferInfo.Sending.SentBytes += fileToSend.Size

					node.transferInfo.Sending.InTransfer = false

					if node.verboseOutput {
						fmt.Printf("\n[File] receiver already has \"%s\"", fileToSend.Name)
					}
				}
			}
		}

		// Transfer section

		// if all files have been sent -> send symlinks
		if len(node.transferInfo.Sending.FilesToSend) == 0 && node.transferInfo.Sending.CurrentSymlinkIndex < uint64(len(node.transferInfo.Sending.SymlinksToSend)) {
			protocol.SendSymlink(node.transferInfo.Sending.SymlinksToSend[node.transferInfo.Sending.CurrentSymlinkIndex], node.netInfo.Conn, encrKey)
			node.transferInfo.Sending.CurrentSymlinkIndex++
			continue
		}

		if len(node.transferInfo.Sending.FilesToSend) == 0 && node.transferInfo.Sending.CurrentSymlinkIndex == uint64(len(node.transferInfo.Sending.SymlinksToSend)) {
			// if there`s nothing else to send - create and send DONE packet
			protocol.SendPacket(node.netInfo.Conn, protocol.Packet{
				Header: protocol.HeaderDone,
			})

			node.stopped = true

			continue
		}

		if node.transferInfo.Sending.AllowedToTransfer && !node.transferInfo.Sending.InTransfer {
			// notify the node about the next file to be sent

			// determine an index of a file with current ID
			var currentFileIndex uint64 = 0
			for index, fileToSend := range node.transferInfo.Sending.FilesToSend {
				if fileToSend.ID == node.transferInfo.Sending.CurrentFileID {
					currentFileIndex = uint64(index)
					break
				}
			}

			fpacket, err := protocol.CreateFilePacket(node.transferInfo.Sending.FilesToSend[currentFileIndex])
			if err != nil {
				panic(err)
			}

			if node.netInfo.EncryptionKey != nil {
				err = fpacket.EncryptBody(node.netInfo.EncryptionKey)
				if err != nil {
					panic(err)
				}
			}

			err = protocol.SendPacket(node.netInfo.Conn, *fpacket)
			if err != nil {
				panic(err)
			}

			// initiate the transfer for this file on the next iteration
			node.transferInfo.Sending.InTransfer = true
			continue
		}

		// if allowed to transfer and the other node is ready to receive packets - send one piece
		// and wait for it to be ready again
		if node.transferInfo.Sending.AllowedToTransfer && node.transferInfo.Sending.CanSendBytes && node.transferInfo.Sending.InTransfer {
			// sending a piece of a single file

			// determine an index of a file with current ID
			var currentFileIndex uint64 = 0
			for index, fileToSend := range node.transferInfo.Sending.FilesToSend {
				if fileToSend.ID == node.transferInfo.Sending.CurrentFileID {
					currentFileIndex = uint64(index)
					break
				}
			}

			sentBytes, err := protocol.SendPiece(node.transferInfo.Sending.FilesToSend[currentFileIndex], node.netInfo.Conn, node.netInfo.EncryptionKey)
			node.transferInfo.Sending.SentBytes += sentBytes
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

				// remove this file from the queue
				node.transferInfo.Sending.FilesToSend = append(node.transferInfo.Sending.FilesToSend[:currentFileIndex], node.transferInfo.Sending.FilesToSend[currentFileIndex+1:]...)

				// set counter to the next file ID
				node.transferInfo.Sending.CurrentFileID++
				node.transferInfo.Sending.InTransfer = false

			case nil:
				node.transferInfo.Sending.CanSendBytes = false

			default:
				node.stopped = true

				fmt.Printf("\n[ERROR] An error occured while sending a piece of \"%s\": %s", node.transferInfo.Sending.FilesToSend[currentFileIndex].Name, err)
				panic(err)
			}
		}
	}
}

func (node *Node) receive() {
	// RECEIVER NODE

	// connect to the sending node
	err := node.connect()
	if err != nil {
		fmt.Printf("\n[ERROR] Could not connect to %s:%d", node.netInfo.ConnAddr, node.netInfo.Port)
		os.Exit(-1)
	}

	// listen for incoming packets
	go protocol.ReceivePackets(node.netInfo.Conn, node.packetPipe)

	// mainloop
	for {
		node.mutex.Lock()
		stopped := node.stopped
		node.mutex.Unlock()

		if stopped {
			fmt.Printf("\n")
			node.disconnect()
			break
		}

		if !node.verboseOutput && node.transferInfo.Receiving.ReceivedBytes != 0 {
			go node.printTransferInfo(time.Second)
		}

		// receive incoming packets and decrypt them if necessary
		incomingPacket, ok := <-node.packetPipe
		if !ok {
			fmt.Printf("\nConnection has been closed unexpectedly\n")
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
					node.transferInfo.Receiving.TotalDownloadSize = file.Size

					size := float32(file.Size) / 1024 / 1024
					sizeLevel := "MiB"
					if size >= 1024 {
						// GiB
						size = size / 1024
						sizeLevel = "GiB"
					}
					fmt.Printf("\n| Filename: %s\n| Size: %.3f %s\n| Checksum: %s\n", file.Name, size, sizeLevel, file.Checksum)

				} else if dir != nil {
					node.transferInfo.Receiving.TotalDownloadSize = dir.Size

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
							fmt.Printf("\n[ERROR] could not create a directory, downloading directly to the specified location")
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
					node.stopped = true
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

					node.transferInfo.Receiving.ReceivedBytes += file.Size

					if node.verboseOutput {
						fmt.Printf("\n[File] already have \"%s\"", file.Name)
					}

				} else {
					// not the same file. Remove it and await new bytes
					os.Remove(file.Path)

					node.mutex.Lock()
					node.transferInfo.Receiving.AcceptedFiles = append(node.transferInfo.Receiving.AcceptedFiles, file)
					node.mutex.Unlock()

					err = protocol.SendPacket(node.netInfo.Conn, protocol.Packet{
						Header: protocol.HeaderReady,
					})
					if err != nil {
						panic(err)
					}
				}

				existingFileHandler.Close()
			} else {
				// does not exist

				node.mutex.Lock()
				node.transferInfo.Receiving.AcceptedFiles = append(node.transferInfo.Receiving.AcceptedFiles, file)
				node.mutex.Unlock()

				err = protocol.SendPacket(node.netInfo.Conn, protocol.Packet{
					Header: protocol.HeaderReady,
				})
				if err != nil {
					panic(err)
				}
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
					if acceptedFile.Handler == nil {
						err = acceptedFile.Open()
						if err != nil {
							panic(err)
						}
					}

					fileBytes := fileBytesBuffer.Bytes()

					wrote, err := acceptedFile.Handler.WriteAt(fileBytes, int64(acceptedFile.SentBytes))
					if err != nil {
						panic(err)
					}
					acceptedFile.SentBytes += uint64(wrote)
					node.transferInfo.Receiving.ReceivedBytes += uint64(wrote)
				}
			}

			readyPacket := protocol.Packet{
				Header: protocol.HeaderReady,
			}
			protocol.SendPacket(node.netInfo.Conn, readyPacket)

		case protocol.HeaderEndfile:
			// one of the files has been received completely

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

					if acceptedFile.Handler == nil {
						err = acceptedFile.Open()
						if err != nil {
							panic(err)
						}
					}

					// remove this file from the pool
					node.transferInfo.Receiving.AcceptedFiles = append(node.transferInfo.Receiving.AcceptedFiles[:index], node.transferInfo.Receiving.AcceptedFiles[index+1:]...)

					// compare checksums
					realChecksum, err := checksum.GetPartialCheckSum(acceptedFile.Handler)
					if err != nil {
						panic(err)
					}

					if realChecksum != acceptedFile.Checksum {
						if node.verboseOutput {
							fmt.Printf("\n[ERROR] \"%s\" is corrupted", acceptedFile.Name)
						}

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

		case protocol.HeaderSymlink:
			// SYMLINK~(string size in binary)(location in the filesystem)(string size in binary)(location of a target)
			packetReader := bytes.NewReader(incomingPacket.Body)

			// extract the location of the symlink
			var locationSize uint64
			binary.Read(packetReader, binary.BigEndian, &locationSize)

			symlinkLocationBytes := make([]byte, locationSize)
			packetReader.Read(symlinkLocationBytes)

			// extract the target of a symlink
			var targetSize uint64
			binary.Read(packetReader, binary.BigEndian, &targetSize)

			symlinkTargetLocationBytes := make([]byte, targetSize)
			packetReader.Read(symlinkTargetLocationBytes)

			symlinkLocation := string(symlinkLocationBytes)
			symlinkTargetLocation := string(symlinkTargetLocationBytes)

			// create a symlink

			// should be already downloaded
			symlinkDir := filepath.Join(node.transferInfo.Receiving.DownloadsPath, filepath.Dir(symlinkLocation))
			os.MkdirAll(symlinkDir, os.ModePerm)

			os.Symlink(
				filepath.Join(node.transferInfo.Receiving.DownloadsPath, symlinkTargetLocation),
				filepath.Join(node.transferInfo.Receiving.DownloadsPath, symlinkLocation))

			protocol.SendPacket(node.netInfo.Conn, protocol.Packet{
				Header: protocol.HeaderReady,
			})

		case protocol.HeaderDone:
			node.mutex.Lock()
			node.stopped = true
			node.mutex.Unlock()

		case protocol.HeaderDisconnecting:
			node.mutex.Lock()
			node.stopped = true
			node.mutex.Unlock()

			fmt.Printf("\n%s disconnected", node.netInfo.Conn.RemoteAddr())
		}
	}
}

// Starts the node in either sending or receiving state and performs the transfer
func (node *Node) Start() {
	switch node.isSending {
	case true:
		node.send()
	case false:
		node.receive()
	}
}
