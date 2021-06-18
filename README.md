# FTU (FileTransferringUtility)
## Send files through the Net ! 

---

## What is that ?
This application is like an FTP server, but overcomplicated and probably overengineered monstrosity. (basically a file server, but P2P).


---

## Why ?
Learning

---

## How does this work ?
In order to transfer one file on one computer to another - they need to establish a connection. 

In order to establish a connection - there needs to be a 1) sender (server) (the owner of the file), waiting for connections, and a 2) receiver (client), who will try to connect to a sender (server). If the requirements are met - a client will connect to a server and the packet exchange will begin.
 
The server and the client needs to communicate with packets according to certain rules, given by a [protocol](https://github.com/Unbewohnte/FTU/tree/main/protocol).

The packet has its header and body. They are divided into several groups of use by headers, this way we can specify what kind of data is stored inside packet`s body and react accordingly.

Thus, with a connection and a way of communication, the sender will send some packets with necessary information about the file to the receiver that describe a filename, its size and a checksum. The client (receiver) will have the choice of accepting or rejecting the packet. If rejected - the connection will be closed and the program will exit. If accepted - the file will be transferred via packets. 

---


## Known issues|problems|lack of features|reasons why it`s bad
- **VERY** slow; somewhat FIXED - [x], now **faster** than before   
- **VERY** expensive on resources; somewhat FIXED - [x], no more **json manipulations**, only **raw bytes**`s wizardry ! 
- If `MAXFILEDATASIZE` is bigger than appr. 1024 - the packets on the other end will not be unmarshalled due to error ??; FIXED - [x], unnecessary, wrong, deprecated, **destroyed !!!**
- Lack of proper error-handling; somewhat FIXED - [x]
- Lack of information about the process of transferring; FIXED - [x]
- No way to verify if the transferred file is not corrupted; FIXED via checksum- [x]
- No encryption; FIXED via AES encryption of packets` body - [x] 
- Messy and hard to follow code && file structure; FIXED? - [x] 
- No tests; FIXED - [x]; Not every packet has its tests, but they are present

## Good points
- It works.

---

## Installation

### From release (Pre-compiled)
- Proceed to [releases page](https://github.com/Unbewohnte/FTU/releases)
- Choose a version/architecture you have and download an archive
- Unpack an archive

### From source (Compile it yourself) (You need [Go](https://golang.org/dl/) and [git](https://git-scm.com/) to be installed on your machine)
- `git clone https://github.com/Unbewohnte/FTU.git`
- `cd` into the folder
- `go build` - to simply compile for your OS/ARCHITECTURE || `CGO_ENABLED=0 GOOS=os_here GOARCH=arch_here go build` - to cross-compile a static executable for the OS/ARCHITECTURE of your choice (`go tool dist list` - to view the available ones)

### After installation
- You probably want to put the executable in some folder and in order not to use it directly from there all the time - add it to the **$PATH** variable

---

## Usage
`./FTU [FLAGS_HERE]` or `FTU [FLAGS_HERE]`

### Flags
`./FTU --help` - to get all flags` description

- `-port` (int) - specifies a working port (if sending - listens on this port, else - tries to connect to this port);
- `-addr` (string) - specifies an address to connect to;
- `-sharefile` (string) - specifies path to a file you want to share, if given a valid path - sender will offer to download this file to receiver;
- `-downloadto` (string) - specifies path to a folder where the receiver wants to store downloaded file;

### Examples

- `./FTU -sharefile="/home/some_path_here/FILETOSHARE.zip"` - creates a server that will share `FILETOSHARE.zip` on port `8080`
- `./FTU -sharefile="/home/some_path_here/FILETOSHARE.zip" - port=727` - same as before, but on port `727`
- `./FTU -downloadto="/home/some_path_here/Downloads/" -addr="192.168.1.104"` - creates a client (receiver) that will try to connect to `192.168.1.104` (local device) on port `8080` and if successful - downloads a file to given path
- `./FTU -downloadto="/home/some_path_here/Downloads/" -addr=145.125.53.212 -port=8888` - same as before, but will try to connect to `145.125.53.212` on port `8888`

---

## Testing

In FTU directory:

- `go test ./...` - to test everything
- `go test -v ./...` - to test everything, with additional information
- `go test ./NAME_OF_THE_PACKAGE` - to test a certain package

---

## IMPORTANT NOTE
This is NOT intended to be a serious application. I'm learning and this is a product of my curiosity. If you're a beginner too, please don't try to find something useful in my code, I am not an expert.

Also, this utility only works if both the server and the client have a port-forwarding|virtual server enabled and configured. Fortunatelly, locally it works without any port-forwarding|virtual servers.

---

## Inspired by [croc](https://github.com/schollz/croc)

--- 

## License
MIT