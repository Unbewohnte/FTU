/*
ftu - file transferring utility.
Copyright (C) 2021,2022  Kasyanov Nikolay Alexeyevich (Unbewohnte (https://unbewohnte.xyz/))

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

// global constants of the protocol
package protocol

// MAXPACKETSIZE.
// How many bytes can contain one packet (header + body) at maximum
// (packets with size bigger than MAXPACKETSIZE are invalid and will not be sent)
const MAXPACKETSIZE uint = 131072 // 128 KiB

// HEADERDELIMETER.
// Character that delimits header of the packet from the body of the packet.
// ie: (packet header)~(packet body)
const HEADERDELIMETER string = "~"

// FILECODE.
const FILECODE string = "f"

// DIRCODE.
const DIRCODE string = "d"
