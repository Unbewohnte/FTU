package checksum

import (
	"testing"
)

func TestBytesToChecksum(t *testing.T) {
	invalidChecksumBytes := []byte("LESSTHAN32")
	_, err := BytesToChecksum(invalidChecksumBytes)
	if err == nil {
		t.Error("BytesToChecksum failed: expected an error")
	}

	invalidChecksumBytes = []byte("BIGGERTHAN32_IFJOWIJFOIHJGLVKNS'O[DFJQWG[OJHNE[OJGNJOREG")
	_, err = BytesToChecksum(invalidChecksumBytes)
	if err == nil {
		t.Error("BytesToChecksum failed: expected an error")
	}

	validChecksumBytes := []byte{5, 194, 47, 217, 251, 195, 69, 230, 216, 121, 253, 38,
		116, 68, 152, 68, 103, 226, 16, 58, 235, 47, 6, 55, 27, 20, 83, 152, 89, 38, 59, 29}
	_, err = BytesToChecksum(validChecksumBytes)
	if err != nil {
		t.Errorf("BytesToChecksum failed: not expected an error, got : %s; length of given bytes: %d", err, len(validChecksumBytes))
	}
}

func TestChecksumToBytes(t *testing.T) {
	validChecksumBytes := []byte{5, 194, 47, 217, 251, 195, 69, 230, 216, 121, 253, 38,
		116, 68, 152, 68, 103, 226, 16, 58, 235, 47, 6, 55, 27, 20, 83, 152, 89, 38, 59, 29}

	var validChecksum CheckSum = CheckSum{5, 194, 47, 217, 251, 195, 69, 230, 216, 121, 253, 38,
		116, 68, 152, 68, 103, 226, 16, 58, 235, 47, 6, 55, 27, 20, 83, 152, 89, 38, 59, 29}

	result := ChecksumToBytes(validChecksum)

	for index, b := range result {
		if b != validChecksumBytes[index] {
			t.Errorf("ChecksumToBytes failed, invalid result")
		}
	}
}
