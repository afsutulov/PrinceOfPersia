// SPDX-License-Identifier: GPL-3.0-or-later
package engine

import (
	"io"
	"modernc.org/libc"
	"sync"
	"time"
	"unsafe"
)

// PCMStream calls the existing Go synthesizer directly. It never installs a
// callback in a native library. Its stream format is stereo signed 16-bit LE.
type PCMStream struct {
	mu               sync.Mutex
	callback         func(*libc.TLS, uintptr, uintptr, int32)
	tls              *libc.TLS
	userdata, buffer uintptr
	rate             int
	chunk, offset    int
	paused, closed   bool
}

func (a *PCMStream) SampleRate() int { a.mu.Lock(); defer a.mu.Unlock(); return a.rate }
func (a *PCMStream) Read(p []byte) (int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return 0, io.EOF
	}
	written := 0
	for written < len(p) {
		if a.offset == a.chunk {
			a.offset = 0
			if a.paused || a.callback == nil {
				if a.chunk == 0 {
					clear(p[written:])
					return len(p), nil
				}
				clear(unsafe.Slice((*byte)(unsafe.Pointer(a.buffer)), a.chunk))
			} else {
				a.callback(a.tls, a.userdata, a.buffer, int32(a.chunk))
			}
		}
		n := copy(p[written:], unsafe.Slice((*byte)(unsafe.Pointer(a.buffer+uintptr(a.offset))), a.chunk-a.offset))
		a.offset += n
		written += n
	}
	return written, nil
}
func (a *PCMStream) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.closed = true
	if a.tls != nil {
		libc.Xfree(a.tls, a.buffer)
		a.tls.Close()
		a.tls = nil
	}
}
func SDL_OpenAudio(t *libc.TLS, desired, obtained uintptr) int32 {
	s := (*SDL_AudioSpec)(unsafe.Pointer(desired))
	if s.Fformat != AUDIO_S16LSB1 || s.Fchannels != 2 || s.Ffreq <= 0 {
		return -1
	}
	a := currentHost.Audio
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tls = libc.NewTLS()
	a.rate = int(s.Ffreq)
	a.userdata = s.Fuserdata
	a.callback = *(*func(*libc.TLS, uintptr, uintptr, int32))(unsafe.Pointer(&struct{ uintptr }{s.Fcallback}))
	a.chunk = int(s.Fsamples) * 4
	if a.chunk == 0 {
		a.chunk = 4096
	}
	a.offset = a.chunk
	a.buffer = libc.Xcalloc(a.tls, 1, uint64(a.chunk))
	a.paused = true
	if obtained != 0 {
		*(*SDL_AudioSpec)(unsafe.Pointer(obtained)) = *s
	}
	return 0
}
func SDL_PauseAudio(t *libc.TLS, pause int32) {
	a := currentHost.Audio
	a.mu.Lock()
	a.paused = pause != 0
	a.offset = a.chunk
	a.mu.Unlock()
}
func SDL_LockAudio(t *libc.TLS) {
	a := currentHost.Audio
	// SDL's audio lock is recursive. The synthesizer can lock again when a
	// MIDI track ends; its private TLS identifies a callback already under mu.
	if t != a.tls {
		a.mu.Lock()
	}
}
func SDL_UnlockAudio(t *libc.TLS) {
	a := currentHost.Audio
	if t != a.tls {
		a.mu.Unlock()
	}
}

var timerMutex sync.Mutex
var timerNext int32
var goTimers = map[int32]*time.Timer{}

func SDL_AddTimer(t *libc.TLS, interval uint32, target, param uintptr) int32 {
	cb := *(*func(*libc.TLS, uint32, uintptr) uint32)(unsafe.Pointer(&struct{ uintptr }{target}))
	tls := libc.NewTLS()
	timerMutex.Lock()
	timerNext++
	id := timerNext
	var run func()
	run = func() {
		next := cb(tls, interval, param)
		timerMutex.Lock()
		defer timerMutex.Unlock()
		if next == 0 {
			delete(goTimers, id)
			tls.Close()
		} else {
			interval = next
			goTimers[id] = time.AfterFunc(time.Duration(next)*time.Millisecond, run)
		}
	}
	goTimers[id] = time.AfterFunc(time.Duration(interval)*time.Millisecond, run)
	timerMutex.Unlock()
	return id
}
