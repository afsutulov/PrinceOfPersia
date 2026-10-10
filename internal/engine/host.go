// SPDX-License-Identifier: GPL-3.0-or-later
package engine

import (
	"bytes"
	"image"
	"io"
	"modernc.org/libc"
	"os"
	"sync"
	"time"
	"unsafe"
)

type WindowState struct {
	Title                                string
	Width, Height                        int
	Fullscreen, HideCursor, IntegerScale bool
	Aspect                               float64
	Icon                                 image.Image
	Revision                             uint64
}
type Frame struct {
	Pixels        []byte
	Width, Height int
	Revision      uint64
}
type Rumble struct {
	ID           int32
	Strong, Weak float64
	Duration     time.Duration
}
type Host struct {
	mu                        sync.Mutex
	window                    WindowState
	frame                     Frame
	events                    []SDL_Event
	keys                      [512]byte
	pads                      []int32
	rumbles                   []Rumble
	text                      bool
	outputWidth, outputHeight int32
	Audio                     *PCMStream
}

func NewHost() *Host {
	return &Host{window: WindowState{Title: "PrinceGo", Width: 640, Height: 480, Aspect: 4.0 / 3}, outputWidth: 640, outputHeight: 480, Audio: &PCMStream{paused: true}}
}

var currentHost = NewHost()

func SetHost(h *Host)               { currentHost = h }
func (h *Host) Window() WindowState { h.mu.Lock(); defer h.mu.Unlock(); return h.window }
func (h *Host) Frame() Frame        { h.mu.Lock(); defer h.mu.Unlock(); return h.frame }
func (h *Host) Resize(w, ht int) {
	h.mu.Lock()
	h.outputWidth, h.outputHeight = int32(w), int32(ht)
	h.mu.Unlock()
}
func (h *Host) TextEnabled() bool { h.mu.Lock(); defer h.mu.Unlock(); return h.text }
func (h *Host) push(e SDL_Event)  { h.mu.Lock(); h.events = append(h.events, e); h.mu.Unlock() }
func eventFrom[T any](v *T) SDL_Event {
	var e SDL_Event
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&e)), 56), unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))))
	return e
}
func (h *Host) Key(scan int32, down, repeat bool, mod uint16) {
	if scan < 0 || scan >= 512 {
		return
	}
	h.mu.Lock()
	h.keys[scan] = 0
	if down {
		h.keys[scan] = 1
	}
	h.mu.Unlock()
	typ := uint32(SDL_KEYUP)
	state := uint8(0)
	if down {
		typ = SDL_KEYDOWN
		state = 1
	}
	rep := uint8(0)
	if repeat {
		rep = 1
	}
	e := SDL_KeyboardEvent{Ftype1: typ, Fstate: state, Frepeat: rep, Fkeysym: SDL_Keysym{Fscancode: scan, Fmod: mod}}
	h.push(eventFrom(&e))
}
func (h *Host) Text(text string) {
	for len(text) > 0 {
		n := min(31, len(text))
		e := SDL_TextInputEvent{Ftype1: SDL_TEXTINPUT}
		for i, b := range []byte(text[:n]) {
			e.Ftext[i] = int8(b)
		}
		h.push(eventFrom(&e))
		text = text[n:]
	}
}
func (h *Host) Focus(gained bool) {
	event := uint8(SDL_WINDOWEVENT_FOCUS_LOST)
	if gained {
		event = SDL_WINDOWEVENT_FOCUS_GAINED
	}
	e := SDL_WindowEvent{Ftype1: SDL_WINDOWEVENT, Fevent: event}
	h.push(eventFrom(&e))
}
func (h *Host) Close() { h.push(SDL_Event{Ftype1: SDL_QUIT}) }
func (h *Host) Gamepads(ids []int32) {
	h.mu.Lock()
	h.pads = append([]int32(nil), ids...)
	h.mu.Unlock()
}
func (h *Host) ControllerDevice(id int32, added bool) {
	typ := uint32(SDL_CONTROLLERDEVICEREMOVED)
	if added {
		typ = SDL_CONTROLLERDEVICEADDED
	}
	e := SDL_ControllerDeviceEvent{Ftype1: typ, Fwhich: id}
	h.push(eventFrom(&e))
}
func (h *Host) ControllerAxis(id int32, axis uint8, value int16) {
	e := SDL_ControllerAxisEvent{Ftype1: SDL_CONTROLLERAXISMOTION, Fwhich: id, Faxis: axis, Fvalue: value}
	h.push(eventFrom(&e))
}
func (h *Host) ControllerButton(id int32, button uint8, down bool) {
	typ := uint32(SDL_CONTROLLERBUTTONUP)
	state := uint8(0)
	if down {
		typ = SDL_CONTROLLERBUTTONDOWN
		state = 1
	}
	e := SDL_ControllerButtonEvent{Ftype1: typ, Fwhich: id, Fbutton: button, Fstate: state}
	h.push(eventFrom(&e))
}
func (h *Host) Rumbles() []Rumble {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := h.rumbles
	h.rumbles = nil
	return out
}

var hostStart = time.Now()
var keyboardSnapshot [512]byte
var hostError = []byte("Go host error\x00")

func SDL_Init(t *libc.TLS, flags uint32) int32          { return 0 }
func SDL_InitSubSystem(t *libc.TLS, flags uint32) int32 { return 0 }
func SDL_Quit(t *libc.TLS)                              {}
func SDL_GetError(t *libc.TLS) uintptr                  { return uintptr(unsafe.Pointer(&hostError[0])) }
func SDL_GetVersion(t *libc.TLS, p uintptr) {
	*(*SDL_version)(unsafe.Pointer(p)) = SDL_version{Fmajor: 2, Fminor: 30, Fpatch: 0}
}
func SDL_Delay(t *libc.TLS, ms uint32) { time.Sleep(time.Duration(ms) * time.Millisecond) }
func SDL_GetPerformanceCounter(t *libc.TLS) uint64 {
	return uint64(time.Since(hostStart).Nanoseconds())
}
func SDL_GetPerformanceFrequency(t *libc.TLS) uint64 { return 1_000_000_000 }
func SDL_GetKeyboardState(t *libc.TLS, n uintptr) uintptr {
	currentHost.mu.Lock()
	keyboardSnapshot = currentHost.keys
	currentHost.mu.Unlock()
	if n != 0 {
		*(*int32)(unsafe.Pointer(n)) = 512
	}
	return uintptr(unsafe.Pointer(&keyboardSnapshot[0]))
}
func SDL_PollEvent(t *libc.TLS, p uintptr) int32 {
	h := currentHost
	h.mu.Lock()
	defer h.mu.Unlock()
	keyboardSnapshot = h.keys
	if len(h.events) == 0 {
		return 0
	}
	*(*SDL_Event)(unsafe.Pointer(p)) = h.events[0]
	h.events = h.events[1:]
	return 1
}
func SDL_PushEvent(t *libc.TLS, p uintptr) int32 {
	currentHost.push(*(*SDL_Event)(unsafe.Pointer(p)))
	return 1
}
func SDL_CreateWindow(t *libc.TLS, title uintptr, x, y, w, ht int32, flags uint32) uintptr {
	h := currentHost
	h.mu.Lock()
	h.window.Title = libc.GoString(title)
	h.window.Width = int(w)
	h.window.Height = int(ht)
	h.window.Fullscreen = flags&SDL_WINDOW_FULLSCREEN_DESKTOP != 0
	h.window.Revision++
	h.mu.Unlock()
	return 1
}
func SDL_GetWindowFlags(t *libc.TLS, p uintptr) uint32 {
	if currentHost.Window().Fullscreen {
		return SDL_WINDOW_FULLSCREEN_DESKTOP
	}
	return SDL_WINDOW_SHOWN | SDL_WINDOW_RESIZABLE | SDL_WINDOW_ALLOW_HIGHDPI
}
func SDL_SetWindowFullscreen(t *libc.TLS, p uintptr, flags uint32) int32 {
	h := currentHost
	h.mu.Lock()
	h.window.Fullscreen = flags&SDL_WINDOW_FULLSCREEN_DESKTOP != 0
	h.window.Revision++
	h.mu.Unlock()
	return 0
}
func SDL_ShowCursor(t *libc.TLS, on int32) int32 {
	h := currentHost
	h.mu.Lock()
	h.window.HideCursor = on == 0
	h.window.Revision++
	h.mu.Unlock()
	return 1
}
func SDL_SetWindowIcon(t *libc.TLS, w, s uintptr) {
	surf := surface(s)
	pixels := surfaceRGBA(s)
	icon := &image.NRGBA{Pix: pixels, Stride: int(surf.Fw) * 4, Rect: image.Rect(0, 0, int(surf.Fw), int(surf.Fh))}
	h := currentHost
	h.mu.Lock()
	h.window.Icon = icon
	h.window.Revision++
	h.mu.Unlock()
}
func SDL_SetHint(t *libc.TLS, k, v uintptr) int32 { return 1 }
func SDL_StartTextInput(t *libc.TLS) {
	currentHost.mu.Lock()
	currentHost.text = true
	currentHost.mu.Unlock()
}
func SDL_StopTextInput(t *libc.TLS) {
	currentHost.mu.Lock()
	currentHost.text = false
	currentHost.mu.Unlock()
}
func SDL_SetTextInputRect(t *libc.TLS, p uintptr) {}

type textureData struct {
	w, h   int32
	pixels []byte
}

var textures = map[uintptr]*textureData{}
var textureNext uintptr = 10
var renderingTarget uintptr
var presenting *textureData
var logicalW, logicalH int32 = 1600, 1200

func SDL_CreateRenderer(t *libc.TLS, w uintptr, i int32, flags uint32) uintptr { return 2 }
func SDL_GetRendererInfo(t *libc.TLS, r, p uintptr) int32 {
	info := (*SDL_RendererInfo)(unsafe.Pointer(p))
	*info = SDL_RendererInfo{Fflags: SDL_RENDERER_SOFTWARE, Fmax_texture_width: 32768, Fmax_texture_height: 32768}
	return 0
}
func SDL_GetRendererOutputSize(t *libc.TLS, r, w, h uintptr) int32 {
	host := currentHost
	host.mu.Lock()
	defer host.mu.Unlock()
	if w != 0 {
		*(*int32)(unsafe.Pointer(w)) = host.outputWidth
	}
	if h != 0 {
		*(*int32)(unsafe.Pointer(h)) = host.outputHeight
	}
	return 0
}
func SDL_RenderSetLogicalSize(t *libc.TLS, r uintptr, w, h int32) int32 {
	logicalW, logicalH = w, h
	host := currentHost
	host.mu.Lock()
	host.window.Aspect = float64(w) / float64(h)
	host.window.Revision++
	host.mu.Unlock()
	return 0
}
func SDL_RenderGetLogicalSize(t *libc.TLS, r, w, h uintptr) {
	if w != 0 {
		*(*int32)(unsafe.Pointer(w)) = logicalW
	}
	if h != 0 {
		*(*int32)(unsafe.Pointer(h)) = logicalH
	}
}
func SDL_RenderSetIntegerScale(t *libc.TLS, r uintptr, on int32) int32 {
	h := currentHost
	h.mu.Lock()
	h.window.IntegerScale = on != 0
	h.window.Revision++
	h.mu.Unlock()
	return 0
}
func SDL_CreateTexture(t *libc.TLS, r uintptr, f uint32, access, w, h int32) uintptr {
	textureNext++
	textures[textureNext] = &textureData{w: w, h: h, pixels: make([]byte, int(w*h)*4)}
	return textureNext
}
func SDL_UpdateTexture(t *libc.TLS, p, rect, pixels uintptr, pitch int32) int32 {
	tx := textures[p]
	if tx == nil {
		return -1
	}
	for y := int32(0); y < tx.h; y++ {
		src := unsafe.Slice((*byte)(unsafe.Pointer(pixels+uintptr(y*pitch))), int(tx.w)*3)
		for x := int32(0); x < tx.w; x++ {
			i := int(y*tx.w+x) * 4
			tx.pixels[i], tx.pixels[i+1], tx.pixels[i+2], tx.pixels[i+3] = src[x*3], src[x*3+1], src[x*3+2], 255
		}
	}
	return 0
}
func SDL_SetRenderTarget(t *libc.TLS, r, p uintptr) int32       { renderingTarget = p; return 0 }
func SDL_RenderClear(t *libc.TLS, r uintptr) int32              { return 0 }
func SDL_RenderCopy(t *libc.TLS, r, tx, src, dst uintptr) int32 { presenting = textures[tx]; return 0 }
func SDL_RenderPresent(t *libc.TLS, r uintptr) {
	if presenting == nil {
		return
	}
	h := currentHost
	h.mu.Lock()
	h.frame = Frame{Pixels: append([]byte(nil), presenting.pixels...), Width: int(presenting.w), Height: int(presenting.h), Revision: h.frame.Revision + 1}
	h.mu.Unlock()
}
func SDL_memset(t *libc.TLS, p uintptr, v int32, n uint64) uintptr { return libc.Xmemset(t, p, v, n) }

var rwNext uintptr = 1000
var rwFiles = map[uintptr]io.ReadWriteSeeker{}

func SDL_RWFromFile(t *libc.TLS, path, mode uintptr) uintptr {
	flags := os.O_RDONLY
	m := libc.GoString(mode)
	if len(m) > 0 && m[0] == 'w' {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	f, e := os.OpenFile(libc.GoString(path), flags, 0600)
	if e != nil {
		return 0
	}
	rwNext++
	rwFiles[rwNext] = f
	return rwNext
}

type readOnlyMemory struct{ *bytes.Reader }

func (r readOnlyMemory) Write(p []byte) (int, error) { return 0, os.ErrPermission }
func SDL_RWFromConstMem(t *libc.TLS, p uintptr, n int32) uintptr {
	if n < 0 {
		return 0
	}
	rwNext++
	rwFiles[rwNext] = readOnlyMemory{bytes.NewReader(unsafe.Slice((*byte)(unsafe.Pointer(p)), int(n)))}
	return rwNext
}
func nativeRWsize(p uintptr) int64 {
	f := rwFiles[p]
	if f == nil {
		return -1
	}
	pos, e := f.Seek(0, io.SeekCurrent)
	if e != nil {
		return -1
	}
	n, e := f.Seek(0, io.SeekEnd)
	f.Seek(pos, io.SeekStart)
	if e != nil {
		return -1
	}
	return n
}
func SDL_RWclose(t *libc.TLS, p uintptr) int32 {
	f := rwFiles[p]
	delete(rwFiles, p)
	if c, ok := f.(io.Closer); ok {
		if c.Close() != nil {
			return -1
		}
	}
	return 0
}
func SDL_RWread(t *libc.TLS, p, buf uintptr, size, count uint64) uint64 {
	if size == 0 || rwFiles[p] == nil {
		return 0
	}
	n, _ := io.ReadFull(rwFiles[p], unsafe.Slice((*byte)(unsafe.Pointer(buf)), int(size*count)))
	return uint64(n) / size
}
func SDL_RWwrite(t *libc.TLS, p, buf uintptr, size, count uint64) uint64 {
	if size == 0 || rwFiles[p] == nil {
		return 0
	}
	n, _ := rwFiles[p].Write(unsafe.Slice((*byte)(unsafe.Pointer(buf)), int(size*count)))
	return uint64(n) / size
}
func SDL_NumJoysticks(t *libc.TLS) int32 {
	currentHost.mu.Lock()
	defer currentHost.mu.Unlock()
	return int32(len(currentHost.pads))
}
func SDL_IsGameController(t *libc.TLS, i int32) int32 { return 1 }
func SDL_GameControllerOpen(t *libc.TLS, i int32) uintptr {
	currentHost.mu.Lock()
	defer currentHost.mu.Unlock()
	if i < 0 || int(i) >= len(currentHost.pads) {
		return 0
	}
	return uintptr(currentHost.pads[i]) + 1
}
func SDL_GameControllerFromInstanceID(t *libc.TLS, i int32) uintptr { return uintptr(i) + 1 }
func SDL_GameControllerClose(t *libc.TLS, p uintptr)                {}
func SDL_JoystickOpen(t *libc.TLS, i int32) uintptr                 { return SDL_GameControllerOpen(t, i) }
func SDL_GameControllerAddMappingsFromRW(t *libc.TLS, p uintptr, free int32) int32 {
	if free != 0 {
		SDL_RWclose(t, p)
	}
	return 0
}
func SDL_GameControllerRumble(t *libc.TLS, p uintptr, strong, weak uint16, ms uint32) int32 {
	h := currentHost
	h.mu.Lock()
	h.rumbles = append(h.rumbles, Rumble{ID: int32(p - 1), Strong: float64(strong) / 65535, Weak: float64(weak) / 65535, Duration: time.Duration(ms) * time.Millisecond})
	h.mu.Unlock()
	return 0
}
func SDL_JoystickRumble(t *libc.TLS, p uintptr, s, w uint16, ms uint32) int32 {
	return SDL_GameControllerRumble(t, p, s, w, ms)
}
func SDL_HapticOpen(t *libc.TLS, i int32) uintptr       { return SDL_GameControllerOpen(t, i) }
func SDL_HapticRumbleInit(t *libc.TLS, p uintptr) int32 { return 0 }
func SDL_HapticRumblePlay(t *libc.TLS, p uintptr, v float32, ms uint32) int32 {
	return SDL_GameControllerRumble(t, p, uint16(v*65535), uint16(v*65535), ms)
}
