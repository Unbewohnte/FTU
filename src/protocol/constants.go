// This file contains global constants of the protocol
package protocol

// MAXPACKETSIZE.
// How many bytes can contain one packet (header + body) at maximum
// (packets with size bigger than MAXPACKETSIZE are invalid and will not be sent)
const MAXPACKETSIZE uint = 131072 // 128 KiB

// HEADERDELIMETER.
// Character that delimits header of the packet from the body of the packet.
// ie: FILEINFO~img.png
const HEADERDELIMETER string = "~"
