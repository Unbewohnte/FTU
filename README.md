# ftu (FileTransferringUtility)
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


## ● Installation

### ● From release (Pre-compiled)
- Proceed to [releases page](https://github.com/Unbewohnte/ftu/releases)
- Choose a version/architecture you have and download an archive
- Unpack an archive
- If on GNU/Linux - run `sudo make install`

### ● From source (Compile it yourself) (You need [Go](https://golang.org/dl/) and [git](https://git-scm.com/) to be installed on your machine)
- `git clone https://github.com/Unbewohnte/ftu.git`
- `cd` into the folder
- `make`
- If on GNU/Linux - run `sudo make install` 

Now you have ftu installed !

---

## ● Usage
`ftu -h` - to print a usage message

`ftu [FLAGS]`

### ● FLAGs
- -p [Uinteger_here] for port
- -r [true|false] for recursive sending of a directory
- -a [ip_address|domain_name] address to connect to (cannot be used with -s)
- -d [path_to_directory] where the files will be downloaded to (cannot be used with -s)
- -s [path_to_file|directory] to send it (cannot be used with -a)
- -l for license text

### ● Examples

`ftu -p 89898 -s /home/user/Downloads/someVideo.mp4`
creates a node on a non-default port 89898 that will send "someVideo.mp4" to the other node that connects to you

`ftu -p 7277 -a 192.168.1.104 -d .`
creates a node that will connect to 192.168.1.104:7277 and download served file|directory to the working directory

`ftu -p 7277 -a 192.168.1.104 -d /home/user/Downloads/`
creates a node that will connect to 192.168.1.104:7277 and download served file|directory to "/home/user/Downloads/"

`ftu -s /home/user/homework`
creates a node that will send every file in the directory

`ftu -r -s /home/user/homework/`
creates a node that will send every file in the directory !RECUSRIVELY!

---

## ● Testing

In 'src' directory:

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
- Send directory
- Wire back encryption