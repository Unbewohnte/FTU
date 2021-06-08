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
var PORT *int = flag.Int("port", 8080, "Specifies a port for a sender")
var SENDERADDR *string = flag.String("addr", "", "Specifies an IP for connection")
var SENDING *bool = flag.Bool("sending", false, "Send or receive")
var DOWNLOADSFOLDER *string = flag.String("downloadto", "", "Specifies where the receiver will store downloaded file")
var SHAREDFILE *string = flag.String("sharefile", "", "Specifies what file sender will send")

// helpMessage
var HELPMSG string = `
"-port", default: 8080, Specifies a port for a sender
"-addr", default: "", Specifies an IP for connection
"-sending", default: false, Send or receive
"-downloadto", default: "", Specifies where the receiver will store downloaded file
"-sharefile", default: "", Specifies what file sender will send`

// Input-validation
func checkFlags() {
	if *SENDING {
		if strings.TrimSpace(*SHAREDFILE) == "" {
			fmt.Println("No file specified !\n", HELPMSG)
			os.Exit(1)
		}
		if *PORT <= 0 {
			fmt.Println("Invalid port !\n", HELPMSG)
			os.Exit(1)
		}
	} else if !*SENDING {
		if strings.TrimSpace(*SENDERADDR) == "" {
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
	if *SENDING {
		// 1) create sender -> 2) wait for a connection ->|
		// 3) send fileinfo packet -> 4) if accepted - upload file
		sender := sender.NewSender(*PORT, *SHAREDFILE)
		sender.WaitForConnection()
		sender.MainLoop()

	} else {
		// 1) create receiver -> 2) try to connect to a sender -> 3) wait for a fileinfo packet ->|
		// 4) accept or refuse -> 5) download|don`t_download file
		receiver := receiver.NewReceiver(*DOWNLOADSFOLDER)
		receiver.Connect(fmt.Sprintf("%s:%d", *SENDERADDR, *PORT))
		receiver.MainLoop()
	}
}
