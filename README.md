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

In order to establish a connection - there needs to be a 1) server (the owner of the file), waiting for connections, and a 2) client, who will try to connect to a server. If the requirements are met - a client will connect to a server and the packet exchange will begin.
 
The server and the client needs to communicate with packets according to certain rules, given by a [protocol](https://github.com/Unbewohnte/FTU/tree/main/protocol).

In my implementation there is only one basic packet template with fixed fields. The packets are divided into several groups by its headers, this way my basic packet`s template can be used in many ways, without need of creating a brand-new packet with a different kind of a template.

Thus, with a connection and a way of communication, the server will send a fileinfo packet to a client that describes a filename and its size. The client will have the choice of accepting or rejecting the packet. If rejected - the connection will be closed and the program will exit. If accepted - the file will be transferred via packets. 

---

## Usage
`./FTU [FLAGS_HERE]` or `FTU [FLAGS_HERE]`

### Flags

- `-server` (bool) - if true - creates a server (also need to provide a `-sharefile` flag in that case), if false - creates a client 
- `-port` (int) - specifies a port; if `-server` == true - listens on that port, else - connects to given port
- `addr` (string) - specifies an address to connect to (used when `-server=false`)
- `-sharefile` (string) - specifies path to a file you want to share (used in pair with `-server=true`), if given a valid path - a server will offer to share this file to a client
- `-downloadto` (string) - specifies path to a folder where the client wants to store downloaded file

### Examples

- `./FTU -server=true -sharefile="/home/some_path_here/FILETOSHARE.zip"` - creates a server that will share `FILETOSHARE.zip` on port `8080`
- `./FTU -server=true -sharefile="/home/some_path_here/FILETOSHARE.zip" - port=727` - same as before, but on port `727`
- `./FTU -server=false -downloadto="/home/some_path_here/Downloads/" -addr="localhost"` - creates a client that will try to connect to `localhost` on port `8080` and if successful - downloads a file to given path
- `./FTU -server=false -downloadto="/home/some_path_here/Downloads/" -addr=145.125.53.212 -port=8888` - same as before, but will try to connect to `145.125.53.212` on port `8888`


---

## Known issues|problems|lack of features|reasons why it`s bad
- **VERY** slow; FIXED - [ ]   
- **VERY** expensive on resources; FIXED - [ ]
- If `MAXFILEDATASIZE` is bigger than appr. 1024 - the packets on the other end will not be unmarshalled due to error ??; FIXED - [ ]
- Lack of proper error-handling; somewhat FIXED - [x]
- Lack of information about the process of transferring (ETA, lost packets, etc.); FIXED - [ ]
- No way to verify if the transferred file is not corrupted; FIXED - [ ]
- No encryption; FIXED - [ ] 

## Good points
- It... works ?

---

## IMPORTANT NOTE
This is NOT intended to be a serious application. I'm learning and this is a product of my curiosity. If you're a beginner too, please don't try to find something useful in my code, I am not an expert.

Also, this utility only works if both the server and the client have a port-forwarding enabled and configured. Fortunatelly, locally it works without any port-forwarding.

---

## Inspired by [croc](https://github.com/schollz/croc)

--- 

## License
MIT