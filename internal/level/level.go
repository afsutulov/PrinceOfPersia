// SPDX-License-Identifier: GPL-3.0-or-later
// Layout follows SDLPoP types.h, level_type (DOS PoP 1.x).
package level

import "fmt"

const Size = 2305
const Rooms = 24
const TilesPerRoom = 30

type Links struct{ Left, Right, Up, Down byte }
type Guard struct{ Tile, Direction, X, SequenceLow, Skill, SequenceHigh, Color byte }
type Level struct {
	Foreground     [720]byte
	Modifiers      [720]byte
	DoorLinks1     [256]byte
	DoorLinks2     [256]byte
	RoomLinks      [Rooms]Links
	UsedRooms      byte
	StartRoom      byte
	StartTile      byte
	StartDirection int8
	Guards         [Rooms]Guard
	Raw            [Size]byte
}

// Level 15 in the supplied archive is 2304 bytes. Only the last unused byte
// is absent; zero-padding reproduces a load into an initially zeroed struct.
func Decode(b []byte) (*Level, error) {
	if len(b) != Size && len(b) != Size-1 {
		return nil, fmt.Errorf("level size %d; expected 2304 or 2305", len(b))
	}
	l := new(Level)
	copy(l.Raw[:], b)
	copy(l.Foreground[:], l.Raw[:720])
	copy(l.Modifiers[:], l.Raw[720:1440])
	copy(l.DoorLinks1[:], l.Raw[1440:1696])
	copy(l.DoorLinks2[:], l.Raw[1696:1952])
	for i := 0; i < Rooms; i++ {
		p := 1952 + i*4
		l.RoomLinks[i] = Links{l.Raw[p], l.Raw[p+1], l.Raw[p+2], l.Raw[p+3]}
	}
	l.UsedRooms = l.Raw[2048]
	l.StartRoom = l.Raw[2112]
	l.StartTile = l.Raw[2113]
	l.StartDirection = int8(l.Raw[2114])
	if l.StartRoom < 1 || l.StartRoom > Rooms || l.StartTile >= TilesPerRoom {
		return nil, fmt.Errorf("invalid start room/tile: %d/%d", l.StartRoom, l.StartTile)
	}
	for i := 0; i < Rooms; i++ {
		l.Guards[i] = Guard{l.Raw[2119+i], l.Raw[2143+i], l.Raw[2167+i], l.Raw[2191+i], l.Raw[2215+i], l.Raw[2239+i], l.Raw[2263+i]}
	}
	return l, nil
}

func (l *Level) Tile(room, column, row int) (kind, modifier byte, err error) {
	if room < 1 || room > Rooms || column < 0 || column >= 10 || row < 0 || row >= 3 {
		return 0, 0, fmt.Errorf("invalid tile address")
	}
	i := (room-1)*TilesPerRoom + row*10 + column
	return l.Foreground[i] & 31, l.Modifiers[i], nil
}

// Trigger returns a door-link target and whether another link follows it.
func (l *Level) Trigger(index int) (room, tile int, next bool, err error) {
	if index < 0 || index >= 256 {
		return 0, 0, false, fmt.Errorf("invalid door link index")
	}
	a, b := l.DoorLinks1[index], l.DoorLinks2[index]
	return int((a&0x60)>>5) + int((b&0xe0)>>3), int(a & 31), a&0x80 == 0, nil
}
