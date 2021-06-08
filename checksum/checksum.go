package checksum

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

type CheckSum [32]byte

// returns a checksum of given file. NOTE, that it creates checksum
// not of a full file (from all file bytes), but from separate byte blocks.
// This is done as an optimisation because the file can be very large in size.
// The general idea:
// BOF... CHUNK -> STEP -> CHUNK... EOF
// checksum := sha256.Sum256(ALLCHUNKS)
// GetPartialCheckSum is default method used to get a file checksum by sender and receiver
func GetPartialCheckSum(file *os.File) (CheckSum, error) {
	// "capturing" CHUNKSIZE bytes and then skipping STEP bytes before the next chunk until the last one
	const CHUNKS uint = 100
	const CHUNKSIZE uint = 100
	const STEP uint = 250

	fileStats, err := file.Stat()
	if err != nil {
		return [32]byte{}, fmt.Errorf("could not get the stats: %s", err)
	}

	fileSize := fileStats.Size()

	if fileSize < int64(CHUNKS*CHUNKSIZE+STEP*(CHUNKS-1)) {
		// file is too small, doing full checksum

		checksum, err := getFullCheckSum(file)
		if err != nil {
			return [32]byte{}, err
		}
		return checksum, nil
	}

	var capturedChunks string
	var read uint64 = 0
	for i := 0; uint(i) < CHUNKS; i++ {
		buffer := make([]byte, CHUNKSIZE)
		r, _ := file.ReadAt(buffer, int64(read))

		capturedChunks += string(buffer)

		read += uint64(r)
		read += uint64(STEP)
	}

	checksum := sha256.Sum256([]byte(capturedChunks))
	return checksum, nil
}

// returns a sha256 checksum of given file
func getFullCheckSum(file *os.File) (CheckSum, error) {
	filebytes, err := io.ReadAll(file)
	if err != nil {
		return [32]byte{}, fmt.Errorf("could not read the file: %s", err)
	}
	checksum := sha256.Sum256(filebytes)

	return checksum, nil
}

// simply compares 2 given checksums. If they are equal - returns true
func AreEqual(checksum1, checksum2 CheckSum) bool {
	var i int = 0
	for _, checksum1Byte := range checksum1 {
		checksum2Byte := checksum2[i]
		if checksum1Byte != checksum2Byte {
			return false
		}
		i++
	}
	return true
}
