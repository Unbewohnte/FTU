package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// Encrypts given data using aes encryption.
// From https://www.melvinvivas.com/how-to-encrypt-and-decrypt-data-using-aes/
func Encrypt(key, dataToEncrypt []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("could not create new AES cipher: %s", err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("could not create new GCM: %s", err)
	}
	nonce := make([]byte, aesGCM.NonceSize())
	encryptedData := aesGCM.Seal(nonce, nonce, dataToEncrypt, nil)

	return encryptedData, nil
}
