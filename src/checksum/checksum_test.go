/*
ftu - file transferring utility.
Copyright (C) 2021,2022  Kasyanov Nikolay Alexeyevich (Unbewohnte (https://unbewohnte.xyz/))

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
	"os"
	"strings"
	"testing"
)

func Test_GetPartialCheckSum(t *testing.T) {
	tesfilePath := "../testfiles/testfile.txt"

	file, err := os.Open(tesfilePath)
	if err != nil {
		t.Fatalf("%s", err)
	}

	checksum, err := GetPartialCheckSum(file)
	if err != nil {
		t.Fatalf("GetPartialCheckSum error: %s", err)
	}

	if !strings.EqualFold("fa6d92493ac0c73c9fa85d10c92b41569017454c5b4387d315f3d2c4ad1d6766", checksum) {
		t.Fatalf("GetPartialCheckSum error: hashes of a testfile.txt do not match")
	}
}
