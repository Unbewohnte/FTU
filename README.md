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

## Usage
`./FTU [FLAGS_HERE]` or `FTU [FLAGS_HERE]`

### Flags

- `-sending` (bool) - if true - creates a server (sender) (also need to provide a `-sharefile` flag in that case), if false - creates a client (receiver) 
- `-port` (int) - specifies a port; if `-sending` == true - listens on that port, else - connects to given port
- `addr` (string) - specifies an address to connect to (used when `-sending=false`)
- `-sharefile` (string) - specifies path to a file you want to share (used in pair with `-sending=true`), if given a valid path - a server will offer to share this file to a client
- `-downloadto` (string) - specifies path to a folder where the client wants to store downloaded file

### Examples

- `./FTU -sending=true -sharefile="/home/some_path_here/FILETOSHARE.zip"` - creates a server that will share `FILETOSHARE.zip` on port `8080`
- `./FTU -sending=true -sharefile="/home/some_path_here/FILETOSHARE.zip" - port=727` - same as before, but on port `727`
- `./FTU -sending=false -downloadto="/home/some_path_here/Downloads/" -addr="192.168.1.104"` - creates a client (receiver) that will try to connect to `192.168.1.104` (local device) on port `8080` and if successful - downloads a file to given path
- `./FTU -sending=false -downloadto="/home/some_path_here/Downloads/" -addr=145.125.53.212 -port=8888` - same as before, but will try to connect to `145.125.53.212` on port `8888`


---

## Known issues|problems|lack of features|reasons why it`s bad
- **VERY** slow; somewhat FIXED - [x], now **faster** than before   
- **VERY** expensive on resources; somewhat FIXED - [x], no more **json manipulations**, only **raw bytes**`s wizardry ! 
- If `MAXFILEDATASIZE` is bigger than appr. 1024 - the packets on the other end will not be unmarshalled due to error ??; FIXED - [x], unnecessary, wrong, deprecated, **destroyed !!!**
- Lack of proper error-handling; somewhat FIXED - [x]
- Lack of information about the process of transferring (ETA, lost packets, etc.); FIXED - [ ]
- No way to verify if the transferred file is not corrupted; FIXED via checksum- [x]
- No encryption; FIXED - [ ] 
- Messy and hard to follow code && file structure; partially FIXED (protocol is looking fairly good rn) - [ X ]
- No way to stop the download/upload and resume it later or even during the next connection; FIXED - [ ] 
- No tests; FIXED - [ ]

## Good points
- It works.

---

## IMPORTANT NOTE
This is NOT intended to be a serious application. I'm learning and this is a product of my curiosity. If you're a beginner too, please don't try to find something useful in my code, I am not an expert.

Also, this utility only works if both the server and the client have a port-forwarding enabled and configured. Fortunatelly, locally it works without any port-forwarding.

---

## Inspired by [croc](https://github.com/schollz/croc)

--- 

## License
MIT