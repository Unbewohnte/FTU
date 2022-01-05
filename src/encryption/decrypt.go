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

package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// Thanks to https://www.melvinvivas.com/how-to-encrypt-and-decrypt-data-using-aes/

// Decrypts encrypted aes data with given key.
func Decrypt(key, dataToDecrypt []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("could not create new AES cipher: %s", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("could not create new GCM: %s", err)
	}

	nonce, encryptedBytes := dataToDecrypt[:aesGCM.NonceSize()], dataToDecrypt[aesGCM.NonceSize():]

	decryptedData, err := aesGCM.Open(nil, nonce, encryptedBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt given data: %s", err)
	}

	return decryptedData, nil
}
