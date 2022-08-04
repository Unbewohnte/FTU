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
	"os"
	"path/filepath"
	"testing"
)

func Test_IsSymlink(t *testing.T) {
	dirpath := "../testfiles/"

	symlinkPath := filepath.Join(dirpath, "testsymlink.txt")
	os.Symlink(filepath.Join(dirpath, "testfile.txt"), symlinkPath)

	isSymlink, err := IsSymlink(symlinkPath)
	if err != nil {
		t.Fatalf("%s\n", err)
	}
	if !isSymlink {
		t.Fatalf("%s expected to be a symlink\n", symlinkPath)
	}
}
