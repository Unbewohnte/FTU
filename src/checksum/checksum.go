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

package checksum

import (
	"bytes"
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
	const CHUNKS uint = 50
	const CHUNKSIZE uint = 50
	const STEP uint = 250

	fileStats, err := file.Stat()
	if err != nil {
		return "", err
	}

	fileSize := fileStats.Size()

	if fileSize < int64(CHUNKS*CHUNKSIZE+STEP*(CHUNKS-1)) {
		// file is too small to chop it in chunks, so just get the full checksum

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

	// var capturedChunks string
	var capturedChunks bytes.Buffer
	var read uint64 = 0
	for i := 0; uint(i) < CHUNKS; i++ {
		buffer := make([]byte, CHUNKSIZE)
		r, _ := file.ReadAt(buffer, int64(read))

		capturedChunks.Write(buffer)

		read += uint64(r)
		read += uint64(STEP)
	}

	checksumBytes := sha256.Sum256(capturedChunks.Bytes())
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
