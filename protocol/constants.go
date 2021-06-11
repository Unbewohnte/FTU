// This file contains global constants of the protocol
package protocol

// MAXPACKETSIZE.
// How many bytes can contain one packet (header + body) at maximum
// (packets with size bigger than MAXPACKETSIZE are invalid and will not be sent)
const MAXPACKETSIZE uint = 1024 // the same problem as in the previous versions: if the packet is big enough - the conn.Read()
// will result in some sort of error where it does not read the intended amount of bytes (less, in fact),
// which is strange, but I guess that
// I just do something wrong in my code

// PACKETSIZEDELIMETER.
// Character that delimits one and the other sides of the next incoming packet.
// ie: |packet_size_here|packet_here, where "|" is PACKETSIZEDELIMETER
const PACKETSIZEDELIMETER string = "|"

// HEADERDELIMETER.
// Character that delimits header of the packet from the body of the packet.
// ie: FILEINFO~img.png
const HEADERDELIMETER string = "~"
