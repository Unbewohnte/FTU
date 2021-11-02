package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// returns a checksum of given file. NOTE, that it creates checksum
// not of a full file (from all file bytes), but from separate byte blocks.
// This is done as an optimisation because the file can be very large in size.
// The general idea:
// BOF... CHUNK -> STEP -> CHUNK... EOF
// checksum := sha256.Sum256(ALLCHUNKS)
// GetPartialCheckSum is default method used to get a file checksum by sender and receiver
func GetPartialCheckSum(file *os.File) (string, error) {
	// "capturing" CHUNKSIZE bytes and then skipping STEP bytes before the next chunk until the last one
	const CHUNKS uint = 100
	const CHUNKSIZE uint = 100
	const STEP uint = 250

	fileStats, err := file.Stat()
	if err != nil {
		return "", err
	}

	fileSize := fileStats.Size()

	if fileSize < int64(CHUNKS*CHUNKSIZE+STEP*(CHUNKS-1)) {
		// file is too small to chop it in chunks, so just doing full checksum

		checksum, err := getFullCheckSum(file)
		if err != nil {
			return "", err
		}
		return checksum, nil
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
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

	checksumBytes := sha256.Sum256([]byte(capturedChunks))
	checksum := hex.EncodeToString(checksumBytes[:])

	return checksum, nil
}

// Returns a sha256 checksum of given file
func getFullCheckSum(file *os.File) (string, error) {
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	filebytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	checksumBytes := sha256.Sum256(filebytes)
	checksum := hex.EncodeToString(checksumBytes[:])

	return checksum, nil
}
