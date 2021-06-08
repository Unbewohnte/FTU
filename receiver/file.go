package receiver

import "github.com/Unbewohnte/FTU/checksum"

// Receiver`s file struct. Used internally by receiver
type File struct {
	Filename string
	Filesize uint64
	CheckSum checksum.CheckSum
}
