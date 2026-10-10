// SPDX-License-Identifier: GPL-3.0-or-later
// Port of SDLPoP seg009.c image decompression algorithms by David Nagy.
package sprite

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
)

type Image struct {
	Width, Height, Depth, Method int
	Pixels                       []byte
}

func Decode(b []byte) (*Image, error) {
	if len(b) < 6 {
		return nil, fmt.Errorf("image header truncated")
	}
	h, w, f := int(binary.LittleEndian.Uint16(b)), int(binary.LittleEndian.Uint16(b[2:])), int(binary.LittleEndian.Uint16(b[4:]))
	depth, method := ((f>>12)&7)+1, (f>>8)&15
	if w < 1 || h < 1 || w > 4096 || h > 4096 {
		return nil, fmt.Errorf("invalid image dimensions %dx%d", w, h)
	}
	if depth != 1 && depth != 2 && depth != 4 && depth != 8 {
		return nil, fmt.Errorf("unsupported depth %d", depth)
	}
	stride := (w*depth + 7) / 8
	packed, err := unpack(b[6:], stride*h, method)
	if err != nil {
		return nil, err
	}
	if method == 2 || method == 4 {
		raster := make([]byte, len(packed))
		for i, v := range packed {
			raster[(i%h)*stride+i/h] = v
		}
		packed = raster
	}
	img := &Image{Width: w, Height: h, Depth: depth, Method: method, Pixels: make([]byte, w*h)}
	mask := (1 << depth) - 1
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bit := x * depth
			img.Pixels[y*w+x] = (packed[y*stride+bit/8] >> uint(8-depth-bit%8)) & byte(mask)
		}
	}
	return img, nil
}

func unpack(src []byte, n, method int) ([]byte, error) {
	out := make([]byte, 0, n)
	p := 0
	read := func() (byte, error) {
		if p >= len(src) {
			return 0, fmt.Errorf("compressed image truncated at byte %d", p)
		}
		v := src[p]
		p++
		return v, nil
	}
	switch method {
	case 0:
		if len(src) < n {
			return nil, fmt.Errorf("raw image truncated")
		}
		return append(out, src[:n]...), nil
	case 1, 2:
		for len(out) < n {
			control, e := read()
			if e != nil {
				return nil, e
			}
			count := int(int8(control))
			if count >= 0 {
				count++
				for i := 0; i < count && len(out) < n; i++ {
					v, e := read()
					if e != nil {
						return nil, e
					}
					out = append(out, v)
				}
			} else {
				v, e := read()
				if e != nil {
					return nil, e
				}
				for i := 0; i < -count && len(out) < n; i++ {
					out = append(out, v)
				}
			}
		}
	case 3, 4:
		var window [1024]byte
		wp := 1024 - 66
		var mask uint16
		emit := func(v byte) { out = append(out, v); window[wp] = v; wp = (wp + 1) & 1023 }
		for len(out) < n {
			mask >>= 1
			if mask&0xff00 == 0 {
				v, e := read()
				if e != nil {
					return nil, e
				}
				mask = uint16(v) | 0xff00
			}
			if mask&1 != 0 {
				v, e := read()
				if e != nil {
					return nil, e
				}
				emit(v)
			} else {
				hi, e := read()
				if e != nil {
					return nil, e
				}
				lo, e := read()
				if e != nil {
					return nil, e
				}
				info := int(hi)<<8 | int(lo)
				rp, count := info&1023, (info>>10)+3
				for i := 0; i < count && len(out) < n; i++ {
					emit(window[rp])
					rp = (rp + 1) & 1023
				}
			}
		}
	default:
		return nil, fmt.Errorf("unsupported compression method %d", method)
	}
	return out, nil
}

// Palette decodes the VGA component of a 100-byte SHPL resource.
func Palette(b []byte) (color.Palette, error) {
	if len(b) != 100 {
		return nil, fmt.Errorf("SHPL size %d; expected 100", len(b))
	}
	p := make(color.Palette, 16)
	for i := 0; i < 16; i++ {
		k := 4 + 3*i
		p[i] = color.NRGBA{b[k] << 2, b[k+1] << 2, b[k+2] << 2, 255}
	}
	p[0] = color.NRGBA{}
	return p, nil
}

func (s *Image) Paletted(p color.Palette) (*image.Paletted, error) {
	for _, v := range s.Pixels {
		if int(v) >= len(p) {
			return nil, fmt.Errorf("pixel index %d exceeds palette", v)
		}
	}
	i := image.NewPaletted(image.Rect(0, 0, s.Width, s.Height), p)
	copy(i.Pix, s.Pixels)
	return i, nil
}
