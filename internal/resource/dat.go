// SPDX-License-Identifier: GPL-3.0-or-later
// The binary format is documented by SDLPoP types.h and seg009.c.
package resource

import (
	"encoding/binary"
	"fmt"
)

type Entry struct {
	ID         uint16 `json:"id"`
	Offset     uint32 `json:"offset"`
	Size       uint16 `json:"size"`
	ChecksumOK bool   `json:"checksum_ok"`
}

type DAT struct {
	Entries []Entry
	data    []byte
	byID    map[uint16]int
}

// Parse validates every range and checksum before exposing resource bytes.
// Each resource is preceded by one checksum byte, not included in Size.
func Parse(data []byte) (*DAT, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("DAT header truncated")
	}
	off := uint64(binary.LittleEndian.Uint32(data))
	size := uint64(binary.LittleEndian.Uint16(data[4:]))
	if off < 6 || size < 2 || off+size > uint64(len(data)) {
		return nil, fmt.Errorf("DAT table outside file")
	}
	n := int(binary.LittleEndian.Uint16(data[off:]))
	if uint64(2+8*n) != size {
		return nil, fmt.Errorf("DAT table size %d inconsistent with %d entries", size, n)
	}
	d := &DAT{data: append([]byte(nil), data...), byID: make(map[uint16]int, n)}
	for i := 0; i < n; i++ {
		p := data[int(off)+2+8*i:]
		e := Entry{ID: binary.LittleEndian.Uint16(p), Offset: binary.LittleEndian.Uint32(p[2:]), Size: binary.LittleEndian.Uint16(p[6:])}
		start, end := uint64(e.Offset), uint64(e.Offset)+1+uint64(e.Size)
		if start < 6 || end > off {
			return nil, fmt.Errorf("resource %d outside payload area", e.ID)
		}
		if _, ok := d.byID[e.ID]; ok {
			return nil, fmt.Errorf("duplicate resource %d", e.ID)
		}
		var sum byte
		for _, b := range data[start:end] {
			sum += b
		}
		e.ChecksumOK = sum == 255
		if !e.ChecksumOK {
			return nil, fmt.Errorf("resource %d checksum mismatch", e.ID)
		}
		d.byID[e.ID] = i
		d.Entries = append(d.Entries, e)
	}
	return d, nil
}

func (d *DAT) Resource(id uint16) ([]byte, error) {
	i, ok := d.byID[id]
	if !ok {
		return nil, fmt.Errorf("resource %d missing", id)
	}
	e := d.Entries[i]
	start := int(e.Offset) + 1
	return append([]byte(nil), d.data[start:start+int(e.Size)]...), nil
}
