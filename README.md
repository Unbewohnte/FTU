# FTU (FileTransferringUtility)
## Send files through the Net ! 

---

## What is that ?
This application is like an FTP server, but overcomplicated, probably overengineered monstrosity. (basically, a file server, but P2P and only one file is shared).


---

## Why ?
Learning

---

## How does this work ?
In order to transfer one file on one computer to another - they need to establish a connection. 

In order to establish a connection - there needs to be a 1) server (the owner of the file), waiting for connections, and a 2) client, who will try to connect to a server. If the requirements are met - a client will connect to a server and the packet exchange will begin.
 
The server and the client needs to communicate with packets according to certain rules, given by a [protocol](https://github.com/unbewohnte/FTU/protocol/).

In my implementation there is only one basic packet template with fixed fields. The packets are divided into several groups by its headers, this way my basic packet`s template can be used in many ways, without need of creating a brand-new packet with a different kind of a template.

Thus, with a connection and a way of communication, the server will send a fileinfo packet to a client that describes a filename and its size. The client will have the choice of accepting or rejecting the packet. If rejected - the connection will be closed and the program will exit. If accepted - the file will be transfered via packets. 

---

## Known issues|problems|lack of features|reasons why it`s bad
1. **VERY** slow
2. **VERY** expensive on resources
3. If `MAXFILEDATASIZE` is bigger than appr. 1024 - the packets on the other end will not be unmarshalled due to error ??
4. Lack of proper error-handling
5. Lack of information about the process of transferring (ETA, lost packets, etc.) 
6. No way to verify if the transferred file is not corrupted
7. No encryption

## Good points
1. It... works ?

---

## IMPORTANT NOTE
This is NOT intended to be a serious application. I'm learning and this is a product of my curiosity. If you're a beginner too, please don't try to find something useful in my code, I am not an expert.

Also, this utility only works if both the server and the client have a port-forwarding enabled and configured. Fortunatelly, locally it works without any port-forwarding.

---

## Inspired by [croc](https://github.com/schollz/croc)

--- 

## License
MIT