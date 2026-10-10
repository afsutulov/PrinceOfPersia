// SPDX-License-Identifier: GPL-3.0-or-later
// Go raster operations for the recovered engine's original ABI structures.
// SDL_* names are internal compatibility names; no SDL library is used.
package engine

import (
	"math/bits"
	"modernc.org/libc"
	"unsafe"
)

type surfaceOptions struct {
	key   uint32
	keyed bool
	alpha uint8
	blend int32
}

var surfaceState = map[uintptr]*surfaceOptions{}

func surface(p uintptr) *SDL_Surface         { return (*SDL_Surface)(unsafe.Pointer(p)) }
func pixelFormat(p uintptr) *SDL_PixelFormat { return (*SDL_PixelFormat)(unsafe.Pointer(p)) }
func surfacePalette(p uintptr) *SDL_Palette  { return (*SDL_Palette)(unsafe.Pointer(p)) }
func SDL_CreateRGBSurface(t *libc.TLS, flags uint32, w, h, depth int32, r, g, b, a uint32) uintptr {
	if w < 0 || h < 0 || w > 32768 || h > 32768 || (depth != 8 && depth != 24 && depth != 32) {
		return 0
	}
	if depth != 8 && r|g|b == 0 {
		r, g, b = 0xff, 0xff00, 0xff0000
	}
	p := libc.Xcalloc(t, 1, uint64(unsafe.Sizeof(SDL_Surface{})))
	f := libc.Xcalloc(t, 1, uint64(unsafe.Sizeof(SDL_PixelFormat{})))
	s := surface(p)
	s.Fformat = f
	s.Fw = w
	s.Fh = h
	s.Fpitch = (w*(depth/8) + 3) &^ 3
	s.Frefcount = 1
	s.Fpixels = libc.Xcalloc(t, 1, uint64(s.Fpitch)*uint64(h))
	s.Fclip_rect = SDL_Rect{Fw: w, Fh: h}
	fm := pixelFormat(f)
	fm.FBitsPerPixel = uint8(depth)
	fm.FBytesPerPixel = uint8(depth / 8)
	fm.Frefcount = 1
	fm.FRmask = r
	fm.FGmask = g
	fm.FBmask = b
	fm.FAmask = a
	for i, v := range []uint32{r, g, b, a} {
		shift, loss := uint8(0), uint8(8)
		if v != 0 {
			shift = uint8(bits.TrailingZeros32(v))
			loss = uint8(8 - bits.OnesCount32(v))
		}
		switch i {
		case 0:
			fm.FRshift = shift
			fm.FRloss = loss
		case 1:
			fm.FGshift = shift
			fm.FGloss = loss
		case 2:
			fm.FBshift = shift
			fm.FBloss = loss
		case 3:
			fm.FAshift = shift
			fm.FAloss = loss
		}
	}
	if depth == 8 {
		fm.Fformat = SDL_PIXELFORMAT_INDEX8
		fm.Fpalette = libc.Xcalloc(t, 1, uint64(unsafe.Sizeof(SDL_Palette{})))
		pal := surfacePalette(fm.Fpalette)
		pal.Fncolors = 256
		pal.Frefcount = 1
		pal.Fcolors = libc.Xcalloc(t, 256, 4)
		for i := range unsafe.Slice((*SDL_Color)(unsafe.Pointer(pal.Fcolors)), 256) {
			(*SDL_Color)(unsafe.Pointer(pal.Fcolors + uintptr(i)*4)).Fa = 255
		}
	} else if depth == 24 {
		if r == 0xff {
			fm.Fformat = SDL_PIXELFORMAT_RGB24
		} else {
			fm.Fformat = SDL_PIXELFORMAT_BGR24
		}
	} else {
		switch r {
		case 0xff:
			fm.Fformat = SDL_PIXELFORMAT_ABGR8888
		case 0xff000000:
			fm.Fformat = SDL_PIXELFORMAT_RGBA8888
		default:
			fm.Fformat = SDL_PIXELFORMAT_ARGB8888
		}
	}
	blend := int32(SDL_BLENDMODE_NONE)
	if a != 0 {
		blend = SDL_BLENDMODE_BLEND
	}
	surfaceState[p] = &surfaceOptions{alpha: 255, blend: blend}
	return p
}
func releasePalette(t *libc.TLS, p uintptr) {
	if p == 0 {
		return
	}
	pal := surfacePalette(p)
	pal.Frefcount--
	if pal.Frefcount == 0 {
		libc.Xfree(t, pal.Fcolors)
		libc.Xfree(t, p)
	}
}
func SDL_FreeSurface(t *libc.TLS, p uintptr) {
	if p == 0 {
		return
	}
	s := surface(p)
	s.Frefcount--
	if s.Frefcount > 0 {
		return
	}
	releasePalette(t, pixelFormat(s.Fformat).Fpalette)
	libc.Xfree(t, s.Fformat)
	libc.Xfree(t, s.Fpixels)
	libc.Xfree(t, p)
	delete(surfaceState, p)
}
func SDL_SetPaletteColors(t *libc.TLS, p, colors uintptr, first, count int32) int32 {
	if p == 0 || first < 0 || count < 0 || first+count > surfacePalette(p).Fncolors {
		return -1
	}
	libc.Xmemcpy(t, surfacePalette(p).Fcolors+uintptr(first)*4, colors, uint64(count)*4)
	surfacePalette(p).Fversion++
	return 0
}
func SDL_SetSurfacePalette(t *libc.TLS, p, pal uintptr) int32 {
	f := pixelFormat(surface(p).Fformat)
	if f.FBitsPerPixel != 8 || pal == 0 {
		return -1
	}
	if f.Fpalette != pal {
		releasePalette(t, f.Fpalette)
		f.Fpalette = pal
		surfacePalette(pal).Frefcount++
	}
	return 0
}
func SDL_LockSurface(t *libc.TLS, p uintptr) int32 { surface(p).Flocked++; return 0 }
func SDL_UnlockSurface(t *libc.TLS, p uintptr)     { surface(p).Flocked-- }
func SDL_SetColorKey(t *libc.TLS, p uintptr, on int32, key uint32) int32 {
	o := surfaceState[p]
	o.keyed = on != 0
	o.key = key
	return 0
}
func SDL_SetSurfaceAlphaMod(t *libc.TLS, p uintptr, a uint8) int32 {
	surfaceState[p].alpha = a
	return 0
}
func SDL_SetSurfaceBlendMode(t *libc.TLS, p uintptr, m int32) int32 {
	surfaceState[p].blend = m
	return 0
}
func SDL_SetClipRect(t *libc.TLS, p, r uintptr) int32 {
	s := surface(p)
	rect := SDL_Rect{Fw: s.Fw, Fh: s.Fh}
	if r != 0 {
		rect = intersect(rect, *(*SDL_Rect)(unsafe.Pointer(r)))
	}
	s.Fclip_rect = rect
	if rect.Fw <= 0 || rect.Fh <= 0 {
		return 0
	}
	return 1
}
func intersect(a, b SDL_Rect) SDL_Rect {
	x := max(a.Fx, b.Fx)
	y := max(a.Fy, b.Fy)
	return SDL_Rect{Fx: x, Fy: y, Fw: max(0, min(a.Fx+a.Fw, b.Fx+b.Fw)-x), Fh: max(0, min(a.Fy+a.Fh, b.Fy+b.Fh)-y)}
}
func readPixel(s *SDL_Surface, x, y int32) uint32 {
	p := s.Fpixels + uintptr(y*s.Fpitch+x*int32(pixelFormat(s.Fformat).FBytesPerPixel))
	switch pixelFormat(s.Fformat).FBytesPerPixel {
	case 1:
		return uint32(*(*byte)(unsafe.Pointer(p)))
	case 3:
		b := unsafe.Slice((*byte)(unsafe.Pointer(p)), 3)
		return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16
	default:
		return *(*uint32)(unsafe.Pointer(p))
	}
}
func writePixel(s *SDL_Surface, x, y int32, v uint32) {
	p := s.Fpixels + uintptr(y*s.Fpitch+x*int32(pixelFormat(s.Fformat).FBytesPerPixel))
	switch pixelFormat(s.Fformat).FBytesPerPixel {
	case 1:
		*(*byte)(unsafe.Pointer(p)) = byte(v)
	case 3:
		b := unsafe.Slice((*byte)(unsafe.Pointer(p)), 3)
		b[0], b[1], b[2] = byte(v), byte(v>>8), byte(v>>16)
	default:
		*(*uint32)(unsafe.Pointer(p)) = v
	}
}
func unpack(f *SDL_PixelFormat, v uint32) (r, g, b, a uint8) {
	if f.Fpalette != 0 {
		c := (*SDL_Color)(unsafe.Pointer(surfacePalette(f.Fpalette).Fcolors + uintptr(v)*4))
		return c.Fr, c.Fg, c.Fb, c.Fa
	}
	r = uint8((v & f.FRmask) >> f.FRshift)
	g = uint8((v & f.FGmask) >> f.FGshift)
	b = uint8((v & f.FBmask) >> f.FBshift)
	a = 255
	if f.FAmask != 0 {
		a = uint8((v & f.FAmask) >> f.FAshift)
	}
	return
}
func SDL_MapRGB(t *libc.TLS, p uintptr, r, g, b uint8) uint32 { return SDL_MapRGBA(t, p, r, g, b, 255) }
func SDL_MapRGBA(t *libc.TLS, p uintptr, r, g, b, a uint8) uint32 {
	f := pixelFormat(p)
	if f.Fpalette != 0 {
		colors := unsafe.Slice((*SDL_Color)(unsafe.Pointer(surfacePalette(f.Fpalette).Fcolors)), int(surfacePalette(f.Fpalette).Fncolors))
		best, dist := 0, int(^uint(0)>>1)
		for i, c := range colors {
			dr, dg, db := int(r)-int(c.Fr), int(g)-int(c.Fg), int(b)-int(c.Fb)
			d := dr*dr + dg*dg + db*db
			if d < dist {
				best, dist = i, d
			}
		}
		return uint32(best)
	}
	return uint32(r)<<f.FRshift&f.FRmask | uint32(g)<<f.FGshift&f.FGmask | uint32(b)<<f.FBshift&f.FBmask | uint32(a)<<f.FAshift&f.FAmask
}
func SDL_FillRect(t *libc.TLS, p, r uintptr, color uint32) int32 {
	s := surface(p)
	rect := s.Fclip_rect
	if r != 0 {
		rect = intersect(rect, *(*SDL_Rect)(unsafe.Pointer(r)))
	}
	for y := rect.Fy; y < rect.Fy+rect.Fh; y++ {
		for x := rect.Fx; x < rect.Fx+rect.Fw; x++ {
			writePixel(s, x, y, color)
		}
	}
	return 0
}
func SDL_ConvertSurface(t *libc.TLS, p, f uintptr, flags uint32) uintptr {
	fm := pixelFormat(f)
	s := surface(p)
	out := SDL_CreateRGBSurface(t, flags, s.Fw, s.Fh, int32(fm.FBitsPerPixel), fm.FRmask, fm.FGmask, fm.FBmask, fm.FAmask)
	dst := surface(out)
	df := pixelFormat(dst.Fformat)
	if fm.Fpalette != 0 {
		SDL_SetPaletteColors(t, df.Fpalette, surfacePalette(fm.Fpalette).Fcolors, 0, surfacePalette(fm.Fpalette).Fncolors)
	}
	sf := pixelFormat(s.Fformat)
	o := surfaceState[p]
	if sf.Fpalette != 0 && fm.Fpalette == sf.Fpalette {
		for y := int32(0); y < s.Fh; y++ {
			libc.Xmemcpy(t, dst.Fpixels+uintptr(y*dst.Fpitch), s.Fpixels+uintptr(y*s.Fpitch), uint64(s.Fw))
		}
		options := *o
		surfaceState[out] = &options
		return out
	}
	for y := int32(0); y < s.Fh; y++ {
		for x := int32(0); x < s.Fw; x++ {
			v := readPixel(s, x, y)
			r, g, b, a := unpack(sf, v)
			if o.keyed && v & ^sf.FAmask == o.key & ^sf.FAmask && df.FAmask != 0 {
				a = 0
			}
			writePixel(dst, x, y, SDL_MapRGBA(t, dst.Fformat, r, g, b, a))
		}
	}
	copyOptions := *o
	surfaceState[out] = &copyOptions
	if o.keyed && df.FAmask != 0 {
		copyOptions.keyed = false
		copyOptions.blend = SDL_BLENDMODE_BLEND
	} else if o.keyed {
		r, g, b, a := unpack(sf, o.key)
		copyOptions.key = SDL_MapRGBA(t, dst.Fformat, r, g, b, a)
	}
	return out
}
func SDL_ConvertSurfaceFormat(t *libc.TLS, p uintptr, format, flags uint32) uintptr {
	r, g, b, a, depth := uint32(0xff), uint32(0xff00), uint32(0xff0000), uint32(0xff000000), int32(32)
	switch format {
	case SDL_PIXELFORMAT_RGB24:
		depth = 24
		a = 0
	case SDL_PIXELFORMAT_BGR24:
		depth = 24
		a = 0
		r, b = b, r
	case SDL_PIXELFORMAT_ARGB8888:
		r, g, b = 0xff0000, 0xff00, 0xff
	case SDL_PIXELFORMAT_RGBA8888:
		r, g, b, a = 0xff000000, 0xff0000, 0xff00, 0xff
	case SDL_PIXELFORMAT_INDEX8:
		depth = 8
		r, g, b, a = 0, 0, 0, 0
	}
	tmp := SDL_CreateRGBSurface(t, 0, 0, 0, depth, r, g, b, a)
	out := SDL_ConvertSurface(t, p, surface(tmp).Fformat, flags)
	SDL_FreeSurface(t, tmp)
	return out
}
func SDL_UpperBlit(t *libc.TLS, src, sr, dst, dr uintptr) int32 {
	return blit(t, src, sr, dst, dr, false)
}
func SDL_UpperBlitScaled(t *libc.TLS, src, sr, dst, dr uintptr) int32 {
	if pixelFormat(surface(src).Fformat).FBytesPerPixel == 1 {
		return -1
	}
	return blit(t, src, sr, dst, dr, true)
}
func blit(t *libc.TLS, src, sr, dst, dr uintptr, scaled bool) int32 {
	s, d := surface(src), surface(dst)
	srect := SDL_Rect{Fw: s.Fw, Fh: s.Fh}
	if sr != 0 {
		srect = *(*SDL_Rect)(unsafe.Pointer(sr))
	}
	drect := SDL_Rect{Fw: srect.Fw, Fh: srect.Fh}
	if scaled && dr == 0 {
		drect.Fw, drect.Fh = d.Fw, d.Fh
	}
	if dr != 0 {
		r := *(*SDL_Rect)(unsafe.Pointer(dr))
		drect.Fx, drect.Fy = r.Fx, r.Fy
		if scaled {
			drect.Fw, drect.Fh = r.Fw, r.Fh
		}
	}
	if srect.Fw <= 0 || srect.Fh <= 0 || drect.Fw <= 0 || drect.Fh <= 0 {
		return 0
	}
	clip := intersect(drect, d.Fclip_rect)
	sf, df := pixelFormat(s.Fformat), pixelFormat(d.Fformat)
	o := surfaceState[src]
	if sf.Fpalette != 0 && ((o.blend == SDL_BLENDMODE_NONE && o.alpha != 255) || (o.blend == SDL_BLENDMODE_BLEND && o.alpha == 255 && !o.keyed)) {
		return -1
	}
	for y := clip.Fy; y < clip.Fy+clip.Fh; y++ {
		sy := srect.Fy + (y-drect.Fy)*srect.Fh/drect.Fh
		if sy < 0 || sy >= s.Fh {
			continue
		}
		for x := clip.Fx; x < clip.Fx+clip.Fw; x++ {
			sx := srect.Fx + (x-drect.Fx)*srect.Fw/drect.Fw
			if sx < 0 || sx >= s.Fw {
				continue
			}
			v := readPixel(s, sx, sy)
			if o.keyed && v & ^sf.FAmask == o.key & ^sf.FAmask {
				continue
			}
			if sf.Fpalette != 0 && df.Fpalette == sf.Fpalette && o.blend == SDL_BLENDMODE_NONE {
				writePixel(d, x, y, v)
				continue
			}
			r, g, b, a := unpack(sf, v)
			a = uint8(uint32(a) * uint32(o.alpha) / 255)
			if o.blend == SDL_BLENDMODE_BLEND {
				if a == 0 {
					continue
				}
				if a != 255 {
					rd, gd, bd, ad := unpack(df, readPixel(d, x, y))
					if sf.FAmask != 0 && df.FBytesPerPixel == 4 && sf.FRmask == df.FRmask && sf.FGmask == df.FGmask && sf.FBmask == df.FBmask && o.alpha == 255 && !o.keyed {
						// Match the original x86 SDL packed-alpha path.
						r = uint8(int(r)*int(a)/256 + int(rd)*(255-int(a))/256)
						g = uint8(int(g)*int(a)/256 + int(gd)*(255-int(a))/256)
						b = uint8(int(b)*int(a)/256 + int(bd)*(255-int(a))/256)
						a = uint8(int(a) + int(ad)*(255-int(a))/256)
					} else if sf.FAmask != 0 && (o.alpha != 255 || o.keyed) {
						// SDL's general path rounds the two products separately.
						r = uint8(int(r)*int(a)/255 + int(rd)*(255-int(a))/255)
						g = uint8(int(g)*int(a)/255 + int(gd)*(255-int(a))/255)
						b = uint8(int(b)*int(a)/255 + int(bd)*(255-int(a))/255)
						a = uint8(int(a) + int(ad)*(255-int(a))/255)
					} else {
						r = uint8((int(r)-int(rd))*int(a)/255 + int(rd))
						g = uint8((int(g)-int(gd))*int(a)/255 + int(gd))
						b = uint8((int(b)-int(bd))*int(a)/255 + int(bd))
						a = uint8(int(a) + int(ad) - int(a)*int(ad)/255)
					}
				}
			} else if sf.FAmask == 0 {
				a = o.alpha
			}
			writePixel(d, x, y, SDL_MapRGBA(t, d.Fformat, r, g, b, a))
		}
	}
	if dr != 0 {
		*(*SDL_Rect)(unsafe.Pointer(dr)) = clip
	}
	return 0
}
func surfaceRGBA(p uintptr) []byte {
	s := surface(p)
	f := pixelFormat(s.Fformat)
	out := make([]byte, int(s.Fw*s.Fh)*4)
	for y := int32(0); y < s.Fh; y++ {
		for x := int32(0); x < s.Fw; x++ {
			r, g, b, a := unpack(f, readPixel(s, x, y))
			i := int(y*s.Fw+x) * 4
			out[i], out[i+1], out[i+2], out[i+3] = r, g, b, a
		}
	}
	return out
}
