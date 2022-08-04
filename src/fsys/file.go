/*
ftu - file transferring utility.
Copyright (C) 2021,2022  Kasyanov Nikolay Alexeyevich (Unbewohnte)

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

package fsys

import (
	"fmt"
	"os"
	"path/filepath"

	"unbewohnte/ftu/checksum"
)

// A struct that represents the necessary file information for transportation through node
type File struct {
	ID                 uint64 // Set manually
	Name               string
	Path               string
	RelativeParentPath string // Relative path to the file, where the highest directory in the hierarchy is the upmost parent dir. Set manually
	Size               uint64
	Checksum           string
	Handler            *os.File // Set when .Open() is called
	SentBytes          uint64   // Set manually during transportation
}

var ErrorNotFile error = fmt.Errorf("not a file")

// Get general information about a file with the
// future ability to open it.
// NOTE that Handler field is nil BY DEFAULT until you
// manually call (file *File).Open() to open it !
func GetFile(path string) (*File, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	stats, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	// check if it is a directory
	if stats.IsDir() {
		return nil, ErrorNotFile
	}

	file := File{
		Name:    stats.Name(),
		Path:    absPath,
		Size:    uint64(stats.Size()),
		Handler: nil,
	}

	// get checksum
	err = file.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	checksum, err := checksum.GetPartialCheckSum(file.Handler)
	if err != nil {
		return nil, err
	}

	file.Checksum = checksum

	return &file, nil
}

// Opens file for read/write operations
func (file *File) Open() error {
	if file.Handler != nil {
		file.Close()
	}

	handler, err := os.OpenFile(file.Path, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	file.Handler = handler

	return nil
}

// file.Handler.Close wrapper
func (file *File) Close() error {
	if file.Handler != nil {
		err := file.Handler.Close()
		if err != nil {
			return err
		}

		file.Handler = nil
	}
	return nil
}
