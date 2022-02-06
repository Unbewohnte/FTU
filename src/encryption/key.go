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

package encryption

import (
	"math/rand"
	"time"
)

// using aes256, so 32 bytes-long key
const KEYLEN uint = 32
const CHARS string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Generates 32 pseudo-random bytes to use as a key
func Generate32AESkey() []byte {
	var generatedKey []byte

	rand.Seed(time.Now().UTC().UnixNano())
	// choosing "random" 32 bytes from CHARS
	for {
		if len(generatedKey) == int(KEYLEN) {
			break
		}
		randomIndex := rand.Intn(len(CHARS))
		generatedKey = append(generatedKey, CHARS[randomIndex])
	}

	return generatedKey
}
