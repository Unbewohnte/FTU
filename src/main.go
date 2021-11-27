/*
ftu - file transferring utility.
Copyright (C) 2021  Kasyanov Nikolay Alexeevich (Unbewohnte (https://unbewohnte.xyz/))

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

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/Unbewohnte/ftu/node"
)

var (
	VERSION string = "v2.1.2"

	versionInformation string = fmt.Sprintf("ftu %s\n\nCopyright (C) 2021  Kasyanov Nikolay Alexeevich (Unbewohnte (https://unbewohnte.xyz/))\nThis program comes with ABSOLUTELY NO WARRANTY.\nThis is free software, and you are welcome to redistribute it under certain conditions; type \"ftu -l\" for details.\n", VERSION)

	//go:embed COPYING
	licenseInformation string

	// flags
	PORT          *uint   = flag.Uint("p", 7270, "Specifies a port to work with")
	RECUSRIVE     *bool   = flag.Bool("r", false, "Recursively send a directory")
	ADDRESS       *string = flag.String("a", "", "Specifies an address to connect to")
	DOWNLOADS_DIR *string = flag.String("d", ".", "Downloads folder")
	SEND          *string = flag.String("s", "", "Specify a file|directory to send")
	VERBOSE       *bool   = flag.Bool("?", false, "Turn on/off verbose output")
	PRINT_VERSION *bool   = flag.Bool("v", false, "Print version information")
	PRINT_LICENSE *bool   = flag.Bool("l", false, "Print license information")

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
		fmt.Printf("| -? [true|false] to turn on|off verbose output\n")
		fmt.Printf("| -l print license information\n")
		fmt.Printf("| -v print version information\n\n\n")

		fmt.Printf("[Examples]\n\n")

		fmt.Printf("| ftu -p 89898 -s /home/user/Downloads/someVideo.mp4\n")
		fmt.Printf("| creates a node on a non-default port 89898 that will send \"someVideo.mp4\" to the other node that connects to you\n\n")

		fmt.Printf("| ftu -p 7277 -a 192.168.1.104 -d .\n")
		fmt.Printf("| creates a node that will connect to 192.168.1.104:7277 and download served file|directory to the working directory\n\n")

		fmt.Printf("| ftu -p 7277 -a 87.117.55.229 -d .\n")
		fmt.Printf("| creates a node that will connect to 87.117.55.229:7277 and download served file|directory to the working directory\n\n")

		fmt.Printf("| ftu -p 7277 -a 192.168.1.104 -d /home/user/Downloads/\n")
		fmt.Printf("| creates a node that will connect to 192.168.1.104:7277 and download served file|directory to \"/home/user/Downloads/\"\n\n")

		fmt.Printf("| ftu -s /home/user/homework\n")
		fmt.Printf("| creates a node that will send every file in the directory\n\n")

		fmt.Printf("| ftu -r -s /home/user/homework/\n")
		fmt.Printf("| creates a node that will send every file in the directory !RECUSRIVELY!\n\n\n")
	}
	flag.Parse()

	if *PRINT_VERSION {
		fmt.Println(versionInformation)
		os.Exit(0)
	}

	if *PRINT_LICENSE {
		fmt.Println(licenseInformation)
		os.Exit(0)
	}

	// validate flags
	if *SEND == "" && *ADDRESS == "" {
		fmt.Printf("Neither sending nor receiving flag was specified. Run ftu -h for help\n")
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
		VerboseOutput: *VERBOSE,
		IsSending:     isSending,
		WorkingPort:   *PORT,
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
