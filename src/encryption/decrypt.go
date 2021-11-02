package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// Decrypts encrypted aes data with given key.
// https://www.melvinvivas.com/how-to-encrypt-and-decrypt-data-using-aes/ - very grateful to the author, THANK YOU.
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
