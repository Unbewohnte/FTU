// This file describes various headers of the protocol and how to use them
package protocol

type Header string

// Headers

//// In the following below examples "|" is PACKETSIZEDELIMETER and "~" is HEADERDELIMETER

// FILENAME.
// This header is sent only by sender. The packet with this header
// must contain a name of the transported file in BODY.
// ie: |18|FILENAME~image.png
const HeaderFilename Header = "FILENAME"

// FILESIZE.
// This header is sent only by sender. The packet with this header
// must contain a size of the transported file in its BODY.
// ie: |15|FILESIZE~512442
const HeaderFileSize Header = "FILESIZE"

// CHECKSUM.
// Just like before, this header must be sent in a packet only by sender,
// BODY must contain a checksum of the transported file.
// ie: |74|CHECKSUM~1673f585148148d0c105af0d55646d6cbbf37e33a7366d3b72d8c5caca13434a
const HeaderChecksum Header = "CHECKSUM"

// DOYOACCEPT.
// Sent by sender after all the information about the transfered file has been sent.
// Receiving a packet with this header means that there will be no more additional information about the
// file and the sender is waiting for response (acceptance or rejection of the file).
// ie: |13|DOYOUACCEPT?~
const HeaderAcceptance Header = "DOYOUACCEPT?"

// FILEBYTES.
// Sent only by sender. The packet`s body must contain
// a portion of transported file`s bytes.
// ie: |70|FILEBYTES~fj2pgfjek;hjg02yg082qyuhg83hvuahjvlhsaoughuihgp9earhguhergh\n
const HeaderFileBytes Header = "FILEBYTES"

// FILEREJECT.
// Sent only by receiver if the user has decided to not download the file.
// The BODY may or may not be empty (preferably empty, of course), in any way, it will not be
// used in any way.
// ie: |11|FILEREJECT~
const HeaderReject Header = "FILEREJECT"

// FILEACCEPT.
// The opposite of the previous FILEREJECT. Send by receiver when
// the user has agreed to download the file.
// ie: |11|FILEACCEPT~
const HeaderAccept Header = "FILEACCEPT"

// DONE.
// Sent by sender. Warns the receiver that the file transfer has been done and
// there is no more information to give.
// ie: |5|DONE~
// Usually after the packet with this header has been sent, the receiver will send
// another packet back with header BYE!, telling that it`s going to disconnect
const HeaderDone Header = "DONE"

// READY.
// Sent by receiver when it hass read and processed the last
// FILEBYTES packet. The sender does not allowed to "spam" FILEBYTES
// packets without the permission of receiver.
// ie: |7|READY!~
const HeaderReady Header = "READY"

// BYE!.
// Packet with this header can be sent both by receiver and sender.
// It`s used when the sender or the receiver are going to disconnect
// and will not be able to communicate.
// (Usually it`s when the error has happened, OR, in a good situation, after the DONE header
// has been sent by sender, warning receiver that there is no data to send)
// The BODY is better to be empty.
// ie: |5|BYE!~
const HeaderDisconnecting Header = "BYE!"