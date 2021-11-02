package node

import (
	"fmt"
	"os"
	"testing"
)

// Not complete
func Test_Sendfile(t *testing.T) {
	rnodeOptions := NodeOptions{
		IsSending:   false,
		WorkingPort: 8888,
		ServerSide: &ServerSideNodeOptions{
			ServingPath: "",
			Recursive:   false,
		},
		ClientSide: &ClientSideNodeOptions{
			ConnectionAddr:      "localhost",
			DownloadsFolderPath: "../testfiles/testDownload/",
		},
	}
	receivingNode, err := NewNode(&rnodeOptions)
	if err != nil {
		fmt.Printf("Error constructing a new node: %s\n", err)
		os.Exit(-1)
	}

	snodeOptions := NodeOptions{
		IsSending:   true,
		WorkingPort: 8888,
		ServerSide: &ServerSideNodeOptions{
			ServingPath: "../testfiles/testfile.txt",
			Recursive:   false,
		},
		ClientSide: &ClientSideNodeOptions{
			ConnectionAddr:      "",
			DownloadsFolderPath: "",
		},
	}

	sendingNode, err := NewNode(&snodeOptions)
	if err != nil {
		fmt.Printf("Error constructing a new node: %s\n", err)
		os.Exit(-1)
	}

	go receivingNode.Start()

	sendingNode.Start()
}
