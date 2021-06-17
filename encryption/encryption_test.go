package encryption

import "testing"

func TestGenerate32AESkey(t *testing.T) {
	generatedKey := Generate32AESkey()

	if len(generatedKey) != int(KEYLEN) {
		t.Errorf("Generate32AESkey failed: generated key`s length does not equal KEYLEN const (32)")
	}
}
