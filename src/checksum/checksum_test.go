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
