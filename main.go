package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Unbewohnte/FTU/client"
	"github.com/Unbewohnte/FTU/server"
)

// flags
var PORT *int = flag.Int("port", 8080, "Specifies a port for a server")
var SERVERADDR *string = flag.String("addr", "", "Specifies an IP for connection")
var ISSERVER *bool = flag.Bool("server", false, "Server")
var DOWNLOADSFOLDER *string = flag.String("downloadto", "", "Specifies where the client will store downloaded file")
var SHAREDFILE *string = flag.String("sharefile", "", "Specifies what file server will serve")

// helpMessage
var HELPMSG string = `
"-port", default: 8080, Specifies a port for a server
"-addr", default: "", Specifies an IP for connection
"-server", default: false, Share file or connect and receive one ?
"-downloadto", default: "", Specifies where the client will store downloaded file
"-sharefile", default: "", Specifies what file server will share`

// Input-validation
func checkFlags() {
	if *ISSERVER {
		if strings.TrimSpace(*SHAREDFILE) == "" {
			fmt.Println("No file specified !\n", HELPMSG)
			os.Exit(1)
		}
		if *PORT <= 0 {
			fmt.Println("Invalid port !\n", HELPMSG)
			os.Exit(1)
		}
	} else if !*ISSERVER {
		if strings.TrimSpace(*SERVERADDR) == "" {
			fmt.Println("Invalid IP address !\n", HELPMSG)
			os.Exit(1)
		}
		if strings.TrimSpace(*DOWNLOADSFOLDER) == "" {
			*DOWNLOADSFOLDER = "./downloads/"
			fmt.Println("Empty downloads folder. Changed to ./downloads/")
		}
	}
}

// parse flags, validate given values
func init() {
	flag.Parse()
	checkFlags()
}

func main() {
	if *ISSERVER {
		// 1) create server -> 2) wait for a client ->|
		// 3) send handshake packet -> 4) if accepted - upload file
		server := server.NewServer(*PORT, *SHAREDFILE)
		server.WaitForConnection()
		server.MainLoop()

	} else {
		// 1) create client -> 2) try to connect to a server -> 3) wait for a handshake ->|
		// 4) accept or refuse -> 5) download|don`t_download file
		client := client.NewClient(*DOWNLOADSFOLDER)
		client.Connect(fmt.Sprintf("%s:%d", *SERVERADDR, *PORT))
		client.MainLoop()
	}
}
