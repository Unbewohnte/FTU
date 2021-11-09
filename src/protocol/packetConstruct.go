// Methods to construct various packets defined in a protocol
package protocol

import (
	"bytes"
	"encoding/binary"

	"github.com/Unbewohnte/ftu/checksum"
	"github.com/Unbewohnte/ftu/fsys"
)

// constructs a ready to send FILE packet
func CreateFilePacket(file *fsys.File) (*Packet, error) {
	err := file.Open()
	if err != nil {
		return nil, err
	}
	defer file.Handler.Close()

	// FILE~(idInBinary)(filenameLengthInBinary)(filename)(filesize)(checksumLengthInBinary)(checksum)

	filePacket := Packet{
		Header: HeaderFile,
	}
	fPacketBodyBuff := new(bytes.Buffer)

	// file id
	binary.Write(fPacketBodyBuff, binary.BigEndian, &file.ID)

	// filename
	filenameLen := uint64(len([]byte(file.Name)))
	binary.Write(fPacketBodyBuff, binary.BigEndian, &filenameLen)
	fPacketBodyBuff.Write([]byte(file.Name))

	// size
	binary.Write(fPacketBodyBuff, binary.BigEndian, &file.Size)

	// checksum
	fileChecksum, err := checksum.GetPartialCheckSum(file.Handler)
	if err != nil {
		return nil, err
	}

	checksumLen := uint64(len([]byte(fileChecksum)))
	binary.Write(fPacketBodyBuff, binary.BigEndian, &checksumLen)
	fPacketBodyBuff.Write([]byte(fileChecksum))

	filePacket.Body = fPacketBodyBuff.Bytes()

	// we do not check for packet size because there is no way that it`ll exceed current
	// maximum of 128 KiB
	return &filePacket, nil
}

// constructs a ready to send DIRECTORY packet
func CreateDirectoryPacket(dir *fsys.Directory) (*Packet, error) {
	dirPacket := Packet{
		Header: HeaderDirectory,
	}

	// DIRECTORY~(dirname size in binary)(dirname)(dirsize)(checksumLengthInBinary)(checksum)

	dirPacketBuffer := new(bytes.Buffer)

	// dirname
	dirnameLength := uint64(len(dir.Name))
	err := binary.Write(dirPacketBuffer, binary.BigEndian, &dirnameLength)
	if err != nil {
		return nil, err
	}
	dirPacketBuffer.Write([]byte(dir.Name))

	// dirsize
	err = binary.Write(dirPacketBuffer, binary.BigEndian, dir.Size)
	if err != nil {
		return nil, err
	}

	dirPacket.Body = dirPacketBuffer.Bytes()

	// we do not check for packet size because there is no way that it`ll exceed current
	// maximum of 128 KiB
	return &dirPacket, nil
}