package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Unbewohnte/FTU/receiver"
	"github.com/Unbewohnte/FTU/sender"
)

// flags
var PORT *int = flag.Int("port", 8080, "Specifies a port for a sender|port to connect to")
var SENDERADDR *string = flag.String("addr", "", "Specifies an address to connect to")
var DOWNLOADSFOLDER *string = flag.String("downloadto", "", "Specifies where the receiver will store downloaded file")
var SHAREDFILE *string = flag.String("sharefile", "", "Specifies what file sender will send")

var SENDING bool

// Input-validation
func processFlags() {
	if *PORT < 0 {
		fmt.Println("Invalid port !")
		os.Exit(-1)
	}

	// going to send file -> sending
	if strings.TrimSpace(*SHAREDFILE) != "" {
		SENDING = true
	}
	// specifying address to connect to -> receiving
	if strings.TrimSpace(*SENDERADDR) != "" {
		if SENDING {
			fmt.Println("Cannot specify an address when sharing !")
			os.Exit(-1)
		}
		SENDING = false
	}
	// specifying path to download to -> receiving
	if strings.TrimSpace(*DOWNLOADSFOLDER) != "" {
		if SENDING {
			fmt.Println("Cannot specify a downloads directory when sharing !")
			os.Exit(-1)
		}
		SENDING = false
	}

}

// parse flags, validate given values
func init() {
	flag.Parse()
	processFlags()
}

func main() {
	if SENDING {
		// 1) create sender -> 2) wait for a connection ->|
		// 3) send info about the file -> 4) if accepted - upload file
		sender := sender.NewSender(*PORT, *SHAREDFILE)
		sender.WaitForConnection()
		sender.MainLoop()

	} else {
		// 1) create receiver -> 2) try to connect to a sender -> 3) wait for an info on the file ->|
		// 4) accept or refuse -> 5) download|don`t_download file
		receiver := receiver.NewReceiver(*DOWNLOADSFOLDER)
		receiver.Connect(fmt.Sprintf("%s:%d", *SENDERADDR, *PORT))
		receiver.MainLoop()
	}
}
