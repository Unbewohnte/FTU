package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// Encrypts given data using aes encryption.
// https://www.melvinvivas.com/how-to-encrypt-and-decrypt-data-using-aes/ - very grateful to the author, THANK YOU.
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
