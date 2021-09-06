package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Unbewohnte/ftu/receiver"
	"github.com/Unbewohnte/ftu/sender"
)

// flags
var PORT *int = flag.Int("port", 8080, "Specifies a port for a sender|port to connect to")
var SENDERADDR *string = flag.String("addr", "", "Specifies an address to connect to")
var DOWNLOADSFOLDER *string = flag.String("downloadto", ".", "Specifies where the receiver will store downloaded file")
var SHAREDFILE *string = flag.String("sharefile", "", "Specifies what file sender will send")

var SENDING bool

// Input-validation
func validateFlags() {
	// port validation
	if *PORT < 0 {
		fmt.Println("Invalid port !")
		os.Exit(-1)
	}

	// sending or receiving
	if strings.TrimSpace(*SHAREDFILE) != "" {
		SENDING = true
	} else if strings.TrimSpace(*SENDERADDR) != "" {
		SENDING = false
	}

	// check for default values in vital flags in case they were not provided
	if strings.TrimSpace(*SENDERADDR) == "" && strings.TrimSpace(*SHAREDFILE) == "" {
		fmt.Println("--help to see available flags")
		os.Exit(-1)
	} else if !SENDING && strings.TrimSpace(*SENDERADDR) == "" {
		fmt.Println("No specified sender`s address")
		os.Exit(-1)
	} else if SENDING && strings.TrimSpace(*SHAREDFILE) == "" {
		fmt.Println("No specified file")
		os.Exit(-1)
	}
}

// parse flags, validate given values
func init() {
	flag.Parse()
	validateFlags()
}

func main() {
	if SENDING {
		// 1) create sender -> 2) wait for a connection ->|
		// 3) send info about the file -> 4) if accepted - upload file
		sender := sender.NewSender(*PORT, *SHAREDFILE)
		sender.WaitForConnection()
		sender.HandleInterrupt()
		sender.MainLoop()

	} else {
		// 1) create receiver -> 2) try to connect to a sender -> 3) wait for an info on the file ->|
		// 4) accept or refuse -> 5) download|don`t_download file
		receiver := receiver.NewReceiver(*DOWNLOADSFOLDER)
		receiver.Connect(fmt.Sprintf("%s:%d", *SENDERADDR, *PORT))
		receiver.HandleInterrupt()
		receiver.MainLoop()
	}
}
