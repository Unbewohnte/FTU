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
)

type Symlink struct {
	TargetPath string
	Path       string
}

// Checks whether path is referring to a symlink or not
func IsSymlink(path string) (bool, error) {
	stats, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	isSymlink := stats.Mode()&os.ModeSymlink != 0

	return isSymlink, nil
}

var ErrorNotSymlink error = fmt.Errorf("not a symlink")

// get necessary information about a symlink in a filesystem. If check is false -
// does not check if path REALLY refers to a symlink
func GetSymlink(path string, check bool) (*Symlink, error) {
	if check {
		isSymlink, err := IsSymlink(path)
		if err != nil {
			return nil, err
		}
		if !isSymlink {
			return nil, ErrorNotSymlink
		}
	}

	target, err := os.Readlink(path)
	if err != nil {
		return nil, err
	}

	symlink := Symlink{
		TargetPath: target,
		Path:       path,
	}

	return &symlink, nil
}
