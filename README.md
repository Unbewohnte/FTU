# ftu (FileTransferringUtility)
## Send files through the Net ! 

---

## ● What is that ?
A P2P (decentralized) file sharing program, overcomplicated and an overengineered one.


## ● Why ?
Learning


## ● How does this work ?
In order to transfer one file on one computer to another - they need to establish a connection. 

In order to establish a connection - there needs to be a 1) sender (server) (the owner of the file), waiting for connections, and a 2) receiver (client), who will try to connect to a sender (server). If the requirements are met - client will connect to server and the packet exchange will begin.
 
The server and the client needs to communicate with packets according to certain rules, given by a [protocol](http://unbewohnte.xyz:3000/Unbewohnte/ftu/src/branch/main/src/protocol).

The packet has its header and body. They are divided into several groups of use by headers, this way we can specify what kind of data is stored inside packet`s body and react accordingly.

Thus, with a connection and a way of communication, the sender will send some packets with necessary information about the file to the receiver that describe a filename, its size and a checksum. The client (receiver) will have the choice of accepting or rejecting the packet. If rejected - the connection will be closed and the program will exit. If accepted - the file will be transferred via packets. 

---


## ● Installation

### ● From release (Pre-compiled)
- Proceed to [releases page](http://unbewohnte.xyz:3000/Unbewohnte/ftu/releases)
- Choose a version/architecture you have and download an archive
- Unpack an archive
- If on GNU/Linux - run `chmod +x install.sh && sudo ./install.sh`

### ● From source (Compile it yourself) (You need [Go](https://golang.org/dl/) and [git](https://git-scm.com/) to be installed on your machine)
- `git clone http://unbewohnte.xyz:3000/Unbewohnte/ftu`
- `cd` into the folder
- If on GNU/Linux - run `make && sudo make install` or `make && chmod +x install.sh && sudo ./install`
- else - cd into src/ folder and simply run `go build`; after that you`re free to put the binary wherever you desire 

Now you have ftu installed !

---

## ● Usage
`ftu -h` - to print a usage message

`ftu [FLAGs]`

### ● FLAGs
- -p [uint] for port
- -r [true|false] for recursive sending of a directory
- -a [ip_address|domain_name] address to connect to (cannot be used with -s)
- -d [path_to_directory] where the files will be downloaded to (cannot be used with -s)
- -s [path_to_file|directory] to send it (cannot be used with -a)
- -? [true|false] to turn on|off verbose output
- -v print version text
- -l print license 

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

`make test` or in "src" directory `go test ./...`

--- 

## ● License
GPLv3 license