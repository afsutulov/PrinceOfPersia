// SPDX-License-Identifier: GPL-3.0-or-later
package engine

import (
	"fmt"
	"modernc.org/libc"
	"os"
	"path/filepath"
	"Prince/internal/dosrng"
	doslevel "Prince/internal/level"
	"Prince/internal/resource"
	"Prince/internal/sprite"
	"unsafe"
)

var levels *resource.DAT

type engineExit struct{ code int32 }

func goExit(t *libc.TLS, code int32) { panic(engineExit{code}) }

func Run(root string, args []string) (err error) {
	defer func() {
		if v := recover(); v != nil {
			if e, ok := v.(engineExit); ok {
				if e.code != 0 {
					err = fmt.Errorf("engine exited with code %d", e.code)
				}
			} else {
				panic(v)
			}
		}
	}()
	b, e := os.ReadFile(filepath.Join(root, "mods", "dos14", "LEVELS.DAT"))
	if e != nil {
		return e
	}
	levels, e = resource.Parse(b)
	if e != nil {
		return e
	}
	if e = os.Chdir(root); e != nil {
		return e
	}
	if len(args) > 128 {
		return fmt.Errorf("too many arguments")
	}
	tls := libc.NewTLS()
	defer tls.Close()
	argv := libc.Xcalloc(tls, uint64(len(args)+1), 8)
	defer libc.Xfree(tls, argv)
	ptrs := unsafe.Slice((*uintptr)(unsafe.Pointer(argv)), len(args)+1)
	for i, a := range args {
		p, e := libc.CString(a)
		if e != nil {
			return e
		}
		ptrs[i] = p
		defer libc.Xfree(tls, p)
	}
	g_argc = int32(len(args))
	g_argv = argv
	// DOS restart uses a single setjmp target. Unwind the Go stack, then re-enter the same start point.
	boot := true
	for {
		restart := false
		func() {
			defer func() {
				if v := recover(); v != nil {
					if j, ok := v.(libc.LongjmpRetval); ok && j.JumpBuffer == uintptr(unsafe.Pointer(&setjmp_buf)) {
						restart = true
					} else {
						panic(v)
					}
				}
			}()
			if boot {
				boot = false
				pop_main(tls)
			} else {
				first_start = 1
				start_game(tls)
			}
		}()
		if !restart {
			return nil
		}
	}
}
func GoRandom(tls *libc.TLS, p uintptr, max uint16) uint16 {
	s, v := dosrng.Next(*(*uint32)(unsafe.Pointer(p)), max)
	*(*uint32)(unsafe.Pointer(p)) = s
	return v
}
func GoLoadLevel(tls *libc.TLS, number int32, dst uintptr, n int32) int32 {
	if levels == nil || number < 0 || number > 15 || n != doslevel.Size {
		return -1
	}
	b, e := levels.Resource(uint16(2000 + number))
	if e != nil {
		return -1
	}
	l, e := doslevel.Decode(b)
	if e != nil {
		return -1
	}
	copy(unsafe.Slice((*byte)(unsafe.Pointer(dst)), doslevel.Size), l.Raw[:])
	return 0
}
func decode_image_go(tls *libc.TLS, src uintptr, n int32, pal uintptr) uintptr {
	if n < 6 {
		return 0
	}
	img, e := sprite.Decode(unsafe.Slice((*byte)(unsafe.Pointer(src)), int(n)))
	if e != nil {
		panic(e)
	}
	p := SDL_CreateRGBSurface(tls, 0, int32(img.Width), int32(img.Height), 8, 0, 0, 0, 0)
	if p == 0 {
		return 0
	}
	s := (*SDL_Surface)(unsafe.Pointer(p))
	if SDL_LockSurface(tls, p) != 0 {
		SDL_FreeSurface(tls, p)
		return 0
	}
	for y := 0; y < img.Height; y++ {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(s.Fpixels+uintptr(y)*uintptr(s.Fpitch))), img.Width), img.Pixels[y*img.Width:(y+1)*img.Width])
	}
	SDL_UnlockSurface(tls, p)
	cp := libc.Xmalloc(tls, 64)
	defer libc.Xfree(tls, cp)
	colors := unsafe.Slice((*SDL_Color)(unsafe.Pointer(cp)), 16)
	palette := (*dat_pal_type)(unsafe.Pointer(pal))
	for i := range colors {
		c := palette.Fvga[i]
		colors[i] = SDL_Color{Fr: c.Fr << 2, Fg: c.Fg << 2, Fb: c.Fb << 2, Fa: 255}
	}
	colors[0] = SDL_Color{}
	format := (*SDL_PixelFormat)(unsafe.Pointer(s.Fformat))
	SDL_SetPaletteColors(tls, format.Fpalette, cp, 0, 16)
	return p
}
