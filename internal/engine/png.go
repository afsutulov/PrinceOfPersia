// SPDX-License-Identifier: GPL-3.0-or-later
package engine

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"io"
	"modernc.org/libc"
	"os"
	"unsafe"
)

func loadPNG(tls *libc.TLS, r io.Reader) uintptr {
	img, e := png.Decode(r)
	if e != nil {
		return 0
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	nrgba := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.Draw(nrgba, nrgba.Bounds(), img, b.Min, draw.Src)
	p := SDL_CreateRGBSurface(tls, 0, int32(w), int32(h), 32, 0xff, 0xff00, 0xff0000, 0xff000000)
	if p == 0 {
		return 0
	}
	s := (*SDL_Surface)(unsafe.Pointer(p))
	if SDL_LockSurface(tls, p) != 0 {
		SDL_FreeSurface(tls, p)
		return 0
	}
	for y := 0; y < h; y++ {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(s.Fpixels+uintptr(y)*uintptr(s.Fpitch))), w*4), nrgba.Pix[y*nrgba.Stride:y*nrgba.Stride+w*4])
	}
	SDL_UnlockSurface(tls, p)
	return p
}
func IMG_Load(tls *libc.TLS, path uintptr) uintptr {
	f, e := os.Open(libc.GoString(path))
	if e != nil {
		return 0
	}
	defer f.Close()
	return loadPNG(tls, f)
}
func IMG_Load_RW(tls *libc.TLS, rw uintptr, free int32) uintptr {
	if rw == 0 {
		return 0
	}
	if free != 0 {
		defer SDL_RWclose(tls, rw)
	}
	n := nativeRWsize(rw)
	if n < 1 || n > 64*1024*1024 {
		return 0
	}
	p := libc.Xmalloc(tls, uint64(n))
	defer libc.Xfree(tls, p)
	if SDL_RWread(tls, rw, p, 1, uint64(n)) != uint64(n) {
		return 0
	}
	return loadPNG(tls, bytes.NewReader(unsafe.Slice((*byte)(unsafe.Pointer(p)), int(n))))
}
func IMG_SavePNG(tls *libc.TLS, surf, path uintptr) int32 {
	p := SDL_ConvertSurfaceFormat(tls, surf, uint32(SDL_PIXELFORMAT_ABGR8888), 0)
	if p == 0 {
		return -1
	}
	defer SDL_FreeSurface(tls, p)
	s := (*SDL_Surface)(unsafe.Pointer(p))
	if SDL_LockSurface(tls, p) != 0 {
		return -1
	}
	defer SDL_UnlockSurface(tls, p)
	img := image.NewNRGBA(image.Rect(0, 0, int(s.Fw), int(s.Fh)))
	for y := 0; y < int(s.Fh); y++ {
		copy(img.Pix[y*img.Stride:y*img.Stride+int(s.Fw)*4], unsafe.Slice((*byte)(unsafe.Pointer(s.Fpixels+uintptr(y)*uintptr(s.Fpitch))), int(s.Fw)*4))
	}
	f, e := os.Create(libc.GoString(path))
	if e != nil {
		return -1
	}
	e = png.Encode(f, img)
	closeErr := f.Close()
	if e != nil || closeErr != nil {
		return -1
	}
	return 0
}
