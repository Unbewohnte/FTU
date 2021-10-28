# FTU (FileTransferringUtility)
## Send files through the Net ! 

---

## ● What is that ?
A P2P file sharing program, but overcomplicated and probably an overengineered one.


## ● Why ?
Learning


## ● How does this work ?
In order to transfer one file on one computer to another - they need to establish a connection. 

In order to establish a connection - there needs to be a 1) sender (server) (the owner of the file), waiting for connections, and a 2) receiver (client), who will try to connect to a sender (server). If the requirements are met - client will connect to server and the packet exchange will begin.
 
The server and the client needs to communicate with packets according to certain rules, given by a [protocol](https://github.com/Unbewohnte/ftu/tree/main/protocol).

The packet has its header and body. They are divided into several groups of use by headers, this way we can specify what kind of data is stored inside packet`s body and react accordingly.

Thus, with a connection and a way of communication, the sender will send some packets with necessary information about the file to the receiver that describe a filename, its size and a checksum. The client (receiver) will have the choice of accepting or rejecting the packet. If rejected - the connection will be closed and the program will exit. If accepted - the file will be transferred via packets. 

---


## ● Known issues|problems|lack of features|reasons why it`s bad
- ~~**VERY** slow~~
- ~~**VERY**~~ expensive on resources
- ~~Lack of proper error-handling~~
- ~~Lack of information about the process of transferring~~
- ~~No way to verify if the transferred file is not corrupted~~
- ~~No encryption~~
- ~~No tests~~
- ~~No interrupt signal handling~~

## ● Good points
- It works.

---

## ● Installation

### ● From release (Pre-compiled)
- Proceed to [releases page](https://github.com/Unbewohnte/ftu/releases)
- Choose a version/architecture you have and download an archive
- Unpack an archive

### ● From source (Compile it yourself) (You need [Go](https://golang.org/dl/) and [git](https://git-scm.com/) to be installed on your machine)
- `git clone https://github.com/Unbewohnte/ftu.git`
- `cd` into the folder
- `go build` - to simply compile for your OS/ARCHITECTURE || `CGO_ENABLED=0 go build` - to compile a static executable

### ● Final steps (optional)
- `cd` into folder if you`re not there already
- `chmod +x install.sh` - make installation script executable
- `sudo ./install.sh`

Now you have ftu installed !

---

## ● Usage
`ftu [FLAGS_HERE]`

### ● Flags
`ftu --help` - to get all flags` description

- `-port` (int) - specifies a working port (if sending - listens on this port, else - tries to connect to this port);
- `-addr` (string) - specifies an address to connect to;
- `-sharefile` (string) - specifies path to a file you want to share, if given a valid path - sender will offer to download this file to receiver;
- `-downloadto` (string) - specifies path to a folder where the receiver wants to store downloaded file;

### ● Examples

- `ftu -sharefile="/home/some_path_here/FILETOSHARE.zip"` - creates a server that will share `FILETOSHARE.zip` on port `8080`
- `ftu -sharefile="/home/some_path_here/FILETOSHARE.zip" - port=727` - same as before, but on port `727`
- `ftu -downloadto="/home/some_path_here/Downloads/" -addr="192.168.1.104"` - creates a client (receiver) that will try to connect to `192.168.1.104` (local device) on port `8080` and if successful - downloads a file to given path
- `ftu -downloadto="/home/some_path_here/Downloads/" -addr=145.125.53.212 -port=8888` - same as before, but will try to connect to `145.125.53.212` on port `8888`

---

## ● Testing

In 'ftu' directory:

- `go test ./...` - to test everything
- `go test -v ./...` - to test everything, with additional information
- `go test ./NAME_OF_THE_PACKAGE` - to test a certain package

---

## ● IMPORTANT NOTE
This is NOT intended to be a serious application. I'm learning and this is a product of my curiosity. If you're a beginner too, please don't try to find something useful in my code, I am not an expert.

Also, this utility only works if the server side has a port-forwarding|virtual server enabled and configured. Fortunatelly, locally it works without any port-forwarding|virtual servers.

---

## ● Inspired by [croc](https://github.com/schollz/croc)

--- 

## ● License
MIT

## ● TODO
- multiple filepaths as args, not as a flag
- send all files in a directory
- send all files in a directory recursively
- ip address as an arg, not as a flag