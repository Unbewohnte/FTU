/*
ftu - file transferring utility.
Copyright (C) 2021,2022  Kasyanov Nikolay Alexeevich (Unbewohnte (https://unbewohnte.xyz/))

This file is a part of ftu

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// This file describes various headers of the protocol and how to use them
package protocol

type Header string

// Headers

//// In the following examples "~" is the HEADERDELIMETER
//// and (size) is 8 bytes long big-endian binary encoded uint64

// ENCRKEY.
// The FIRST header to be sent if you`re going to encrypt the transfer. Sent immediately after the connection has been established
// by sender. Body contains a size of a key and the key itself.
// ie: ENCRKEY~(size)(encryption key)
const HeaderEncryptionKey Header = "ENCRKEY"

// REJECT.
// Sent only by receiver if the receiver has decided to not download the contents.
// ie: REJECT~
const HeaderReject Header = "REJECT"

// ACCEPT.
// The opposite of the previous REJECT. Sent by receiver when
// it has agreed to download the file|directory.
// ie: ACCEPT~
const HeaderAccept Header = "ACCEPT"

// DONE.
// Sent by sender. Warns the receiver that the transfer has been done and
// there is no more information to give.
// ie: DONE~
// Usually after the packet with this header has been sent, the receiver will send
// another packet back with header BYE!, telling that it`s going to disconnect
const HeaderDone Header = "DONE"

// READY.
// Sent by receiver when it has read and processed the last
// FILEBYTES or FILE packet. The sender is not allowed to "spam" FILEBYTES or FILE
// packets without the permission (packet with this header) from receiver.
// ie: READY!~
const HeaderReady Header = "READY"

// BYE!.
// Packet with this header can be sent both by receiver and sender.
// It`s used when the sender or the receiver are going to disconnect
// and will not be able to communicate.
// (Usually it`s when the error has happened OR after the DONE header
// has been sent by sender, warning receiver that there is no data to send)
// The BODY is better to be empty.
// ie: BYE!~
const HeaderDisconnecting Header = "BYE!"

// TRANSFEROFFER.
// Sent by sender AFTER ENCRKEY packet if present and BEFORE any other transfer-specific
// packet ONLY ONCE. Asks the receiving node whether it accepts or rejects the transfer of
// offered single file or a directory.
// The body must contain a file or directory code that tells whether
// a file or a directory will be sent in case of acceptance. The rest must be identical either to the FILE or DIRECTORY packet.
// e for directory: TRANSFER~(dircode)(dirname size in binary)(dirname)(dirsize)
// e for a single file: TRANSFER~(filecode)(id in binary)(filename length in binary)(filename)(filesize)(checksum length in binary)(checksum)
// dircode and filecode are pre-declared in the constants of the protocol (d) and (f).
// The actual transfer must start only after the other node has accepted the dir/file with ACCEPT packet.
const HeaderTransferOffer Header = "TRANSFEROFFER"

// FILE.
// Sent by sender, indicating that the file is going to be sent.
// The body structure must follow such structure:
// FILE~(id in binary)(filename length in binary)(filename)(filesize)(checksum length in binary)(checksum)(relative path to the upper directory size in binary if present)(relative path)
// relative path is not needed when the file is already in the root of the initial directory, but must be included when
// the whole directory is being sent recursively
const HeaderFile Header = "FILE"

// FILEBYTES.
// Sent only by sender. The packet`s body must contain
// a file`s Identifier and a portion of its bytes.
// ie: FILEBYTES~(file ID in binary)(file`s binary data)
const HeaderFileBytes Header = "FILEBYTES"

// ENDFILE
// Sent by sender when the file`s contents fully has been sent.
// The body must contain a file ID.
// ie: ENDFILE~(file ID in binary)
const HeaderEndfile Header = "ENDFILE"

// DIRECTORY
// Sent by sender. Used in TRANSFEROFFER packet to tell the difference
// between a file and a directory.
// ie: DIRECTORY~(dirname size in binary)(dirname)(dirsize)
const HeaderDirectory Header = "DIRECTORY"

// ALREADYHAVE
// Sent by receiver in case there is the same file that already exists.
// Sender upon receiving such packet with specified file ID must not send it.
// Body must contain a file ID.
// ie: ALREADYHAVE~(file ID in binary)
const HeaderAlreadyHave Header = "ALREADYHAVE"
