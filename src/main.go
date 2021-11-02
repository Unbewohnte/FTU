package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/Unbewohnte/ftu/node"
)

// flags
var (
	PORT          *uint   = flag.Uint("p", 7270, "Specifies a port to work with")
	PRINT_LICENSE *bool   = flag.Bool("l", false, "Prints a license text")
	RECUSRIVE     *bool   = flag.Bool("r", false, "Recursively send a directory")
	ADDRESS       *string = flag.String("a", "", "Specifies an address to connect to")
	DOWNLOADS_DIR *string = flag.String("d", ".", "Downloads folder")
	SEND          *string = flag.String("s", "", "Specify a file|directory to send")

	//go:embed LICENSE
	licenseText string

	isSending bool
)

func init() {
	flag.Usage = func() {
		fmt.Printf("ftu -[FLAG]...\n\n")

		fmt.Printf("[FLAGs]\n\n")
		fmt.Printf("| -p [Uinteger_here] for port\n")
		fmt.Printf("| -r [true|false] for recursive sending of a directory\n")
		fmt.Printf("| -a [ip_address|domain_name] address to connect to (cannot be used with -s)\n")
		fmt.Printf("| -d [path_to_directory] where the files will be downloaded to (cannot be used with -s)\n")
		fmt.Printf("| -s [path_to_file|directory] to send it (cannot be used with -a)\n")
		fmt.Printf("| -l for license text\n\n\n")

		fmt.Printf("[Examples]\n\n")

		fmt.Printf("| ftu -p 89898 -s /home/user/Downloads/someVideo.mp4\n")
		fmt.Printf("| creates a node on a non-default port 89898 that will send \"someVideo.mp4\" to the other node that connects to you\n\n")

		fmt.Printf("| ftu -p 7277 -a 192.168.1.104 -d .\n")
		fmt.Printf("| creates a node that will connect to 192.168.1.104:7277 and download served file|directory to the working directory\n\n")

		fmt.Printf("| ftu -p 7277 -a 192.168.1.104 -d /home/user/Downloads/\n")
		fmt.Printf("| creates a node that will connect to 192.168.1.104:7277 and download served file|directory to \"/home/user/Downloads/\"\n\n")

		fmt.Printf("| ftu -s /home/user/homework\n")
		fmt.Printf("| creates a node that will send every file in the directory\n\n")

		fmt.Printf("| ftu -r -s /home/user/homework/\n")
		fmt.Printf("| creates a node that will send every file in the directory !RECUSRIVELY!\n\n\n")

	}
	flag.Parse()

	if *PRINT_LICENSE {
		fmt.Println(licenseText)
		os.Exit(0)
	}

	// validate flags
	if *SEND == "" && *ADDRESS == "" {
		fmt.Printf("Neither sending nor receiving flag was specified. Run ftu -h for help")
		os.Exit(-1)
	}

	if *SEND != "" && *ADDRESS != "" {
		fmt.Printf("Can`t send and receive at the same time. Specify only -s or -a\n")
		os.Exit(-1)
	}

	// sending or receiving
	if *SEND != "" {
		// sending
		isSending = true
	} else if *ADDRESS != "" {
		// receiving
		isSending = false
	}
}

func main() {
	nodeOptions := node.NodeOptions{
		IsSending:   isSending,
		WorkingPort: *PORT,
		ServerSide: &node.ServerSideNodeOptions{
			ServingPath: *SEND,
			Recursive:   *RECUSRIVE,
		},
		ClientSide: &node.ClientSideNodeOptions{
			ConnectionAddr:      *ADDRESS,
			DownloadsFolderPath: *DOWNLOADS_DIR,
		},
	}

	node, err := node.NewNode(&nodeOptions)
	if err != nil {
		fmt.Printf("Error constructing a new node: %s\n", err)
		os.Exit(-1)
	}

	node.Start()
}
