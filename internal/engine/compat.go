// SPDX-License-Identifier: GPL-3.0-or-later
// Portable file and directory operations preserve the engine's generated memory layout.
package engine

import (
	"fmt"
	"io"
	"modernc.org/libc"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// The translated engine seeds its DOS RNG with Unix seconds. Use Go's clock:
// modernc.org/libc's Windows Xtime implementation is an unimplemented stub.
func unixSeconds() int64 { return time.Now().Unix() }

var compatMutex sync.Mutex
var allocaFrames = map[*libc.TLS][][]uintptr{}

func allocaEntry(t *libc.TLS) {
	compatMutex.Lock()
	allocaFrames[t] = append(allocaFrames[t], nil)
	compatMutex.Unlock()
}
func allocaExit(t *libc.TLS) {
	compatMutex.Lock()
	a := allocaFrames[t]
	last := a[len(a)-1]
	allocaFrames[t] = a[:len(a)-1]
	compatMutex.Unlock()
	for _, p := range last {
		libc.Xfree(t, p)
	}
}
func goAlloca(t *libc.TLS, n uint64) uintptr {
	p := libc.Xmalloc(t, n)
	compatMutex.Lock()
	a := allocaFrames[t]
	i := len(a) - 1
	a[i] = append(a[i], p)
	compatMutex.Unlock()
	return p
}

var ctypeOnce sync.Once
var ctypePointer uintptr

func goCtype(t *libc.TLS) uintptr {
	ctypeOnce.Do(func() {
		ctypePointer = libc.Xcalloc(t, 1, 8)
		p := libc.Xcalloc(t, 256, 2)
		*(*uintptr)(unsafe.Pointer(ctypePointer)) = p
		for _, c := range []byte{' ', '\t', '\r', '\n', '\v', '\f'} {
			*(*uint16)(unsafe.Pointer(p + uintptr(c)*2)) = 8192
		}
	})
	return ctypePointer
}
func goStrtol(t *libc.TLS, p, end uintptr, base int32) int64 {
	s := libc.GoString(p)
	i := 0
	for i < len(s) && strings.ContainsRune(" \t\r\n\v\f", rune(s[i])) {
		i++
	}
	start := i
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	digits := i
	b := int(base)
	if b == 0 {
		b = 10
		if i < len(s) && s[i] == '0' {
			b = 8
			if i+2 < len(s) && (s[i+1] == 'x' || s[i+1] == 'X') {
				b = 16
				i += 2
				digits = i
			}
		}
	} else if b == 16 && i+2 < len(s) && s[i] == '0' && (s[i+1] == 'x' || s[i+1] == 'X') {
		i += 2
		digits = i
	}
	for i < len(s) {
		c := s[i]
		d := int(c) - int('0')
		if c >= 'a' && c <= 'z' {
			d = int(c-'a') + 10
		} else if c >= 'A' && c <= 'Z' {
			d = int(c-'A') + 10
		}
		if d < 0 || d >= b {
			break
		}
		i++
	}
	if i == digits {
		if end != 0 {
			*(*uintptr)(unsafe.Pointer(end)) = p
		}
		return 0
	}
	v, _ := strconv.ParseInt(s[start:i], int(base), 64)
	if end != 0 {
		*(*uintptr)(unsafe.Pointer(end)) = p + uintptr(i)
	}
	return v
}

type compatFile struct {
	f   *os.File
	eof bool
	err error
	id  int32
}

var compatFiles = map[uintptr]*compatFile{}
var compatDescriptors = map[int32]*compatFile{}
var compatFileNext int32

func goFopen(t *libc.TLS, path, mode uintptr) uintptr {
	m := libc.GoString(mode)
	flag := os.O_RDONLY
	if strings.HasPrefix(m, "w") {
		flag = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	} else if strings.HasPrefix(m, "a") {
		flag = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	}
	if strings.Contains(m, "+") {
		flag = (flag &^ os.O_WRONLY) | os.O_RDWR
	}
	f, e := os.OpenFile(libc.GoString(path), flag, 0600)
	if e != nil {
		return 0
	}
	p := libc.Xcalloc(t, 1, 8)
	compatMutex.Lock()
	compatFileNext++
	cf := &compatFile{f: f, id: compatFileNext}
	compatFiles[p] = cf
	compatDescriptors[cf.id] = cf
	compatMutex.Unlock()
	return p
}
func getFile(p uintptr) *compatFile {
	compatMutex.Lock()
	defer compatMutex.Unlock()
	return compatFiles[p]
}
func goFclose(t *libc.TLS, p uintptr) int32 {
	compatMutex.Lock()
	f := compatFiles[p]
	delete(compatFiles, p)
	if f != nil {
		delete(compatDescriptors, f.id)
	}
	compatMutex.Unlock()
	if f == nil {
		return -1
	}
	e := f.f.Close()
	libc.Xfree(t, p)
	if e != nil {
		return -1
	}
	return 0
}
func goFread(t *libc.TLS, p uintptr, size, n uint64, h uintptr) uint64 {
	f := getFile(h)
	if f == nil || size == 0 || n == 0 {
		return 0
	}
	b := unsafe.Slice((*byte)(unsafe.Pointer(p)), int(size*n))
	got, e := io.ReadFull(f.f, b)
	if e == io.EOF || e == io.ErrUnexpectedEOF {
		f.eof = true
	} else if e != nil {
		f.err = e
	}
	return uint64(got) / size
}
func goFwrite(t *libc.TLS, p uintptr, size, n uint64, h uintptr) uint64 {
	f := getFile(h)
	if f == nil || size == 0 || n == 0 {
		return 0
	}
	got, e := f.f.Write(unsafe.Slice((*byte)(unsafe.Pointer(p)), int(size*n)))
	if e != nil {
		f.err = e
	}
	return uint64(got) / size
}
func goFseek(t *libc.TLS, h uintptr, offset int64, whence int32) int32 {
	f := getFile(h)
	if f == nil {
		return -1
	}
	_, e := f.f.Seek(offset, int(whence))
	if e != nil {
		f.err = e
		return -1
	}
	f.eof = false
	return 0
}
func goFtell(t *libc.TLS, h uintptr) int64 {
	f := getFile(h)
	if f == nil {
		return -1
	}
	v, e := f.f.Seek(0, io.SeekCurrent)
	if e != nil {
		return -1
	}
	return v
}
func goFeof(t *libc.TLS, h uintptr) int32 {
	f := getFile(h)
	if f != nil && f.eof {
		return 1
	}
	return 0
}
func goFerror(t *libc.TLS, h uintptr) int32 {
	f := getFile(h)
	if f == nil || f.err != nil {
		return 1
	}
	return 0
}
func goFgetc(t *libc.TLS, h uintptr) int32 {
	var b [1]byte
	f := getFile(h)
	if f == nil {
		return -1
	}
	_, e := io.ReadFull(f.f, b[:])
	if e != nil {
		if e == io.EOF {
			f.eof = true
		} else {
			f.err = e
		}
		return -1
	}
	return int32(b[0])
}
func goFgets(t *libc.TLS, p uintptr, n int32, h uintptr) uintptr {
	if n <= 1 {
		return 0
	}
	b := unsafe.Slice((*byte)(unsafe.Pointer(p)), int(n))
	i := 0
	for i < int(n)-1 {
		c := goFgetc(t, h)
		if c < 0 {
			break
		}
		b[i] = byte(c)
		i++
		if c == 10 {
			break
		}
	}
	b[i] = 0
	if i == 0 {
		return 0
	}
	return p
}
func goFileno(t *libc.TLS, h uintptr) int32 {
	f := getFile(h)
	if f == nil {
		return -1
	}
	return f.id
}
func fillStat(info os.FileInfo, p uintptr) {
	s := (*stat1)(unsafe.Pointer(p))
	*s = stat1{}
	s.Fst_size = info.Size()
	s.Fst_mode = 0100000 | uint32(info.Mode().Perm())
	if info.IsDir() {
		s.Fst_mode = 0040000 | uint32(info.Mode().Perm())
	}
	s.Fst_mtim.Ftv_sec = info.ModTime().Unix()
}
func goStat(t *libc.TLS, path, dst uintptr) int32 {
	i, e := os.Stat(libc.GoString(path))
	if e != nil {
		return -1
	}
	fillStat(i, dst)
	return 0
}
func goFstat(t *libc.TLS, fd int32, dst uintptr) int32 {
	compatMutex.Lock()
	f := compatDescriptors[fd]
	compatMutex.Unlock()
	if f == nil {
		return -1
	}
	i, e := f.f.Stat()
	if e != nil {
		return -1
	}
	fillStat(i, dst)
	return 0
}
func goMkdir(t *libc.TLS, p uintptr, mode uint32) int32 {
	e := os.Mkdir(libc.GoString(p), os.FileMode(mode))
	if e != nil {
		return -1
	}
	return 0
}
func goAccess(t *libc.TLS, p uintptr, mode int32) int32 {
	_, e := os.Stat(libc.GoString(p))
	if e != nil {
		return -1
	}
	return 0
}
func goRemove(t *libc.TLS, p uintptr) int32 {
	if os.Remove(libc.GoString(p)) != nil {
		return -1
	}
	return 0
}

type compatDir struct {
	entries []os.DirEntry
	index   int
	entry   uintptr
}

var compatDirs = map[uintptr]*compatDir{}

func goOpendir(t *libc.TLS, p uintptr) uintptr {
	entries, e := os.ReadDir(libc.GoString(p))
	if e != nil {
		return 0
	}
	handle := libc.Xcalloc(t, 1, 8)
	entry := libc.Xcalloc(t, 1, uint64(unsafe.Sizeof(dirent{})))
	compatMutex.Lock()
	compatDirs[handle] = &compatDir{entries: entries, entry: entry}
	compatMutex.Unlock()
	return handle
}
func goReaddir(t *libc.TLS, p uintptr) uintptr {
	compatMutex.Lock()
	defer compatMutex.Unlock()
	d := compatDirs[p]
	if d == nil || d.index == len(d.entries) {
		return 0
	}
	e := d.entries[d.index]
	d.index++
	v := (*dirent)(unsafe.Pointer(d.entry))
	*v = dirent{}
	if e.IsDir() {
		v.Fd_type = 4
	} else {
		v.Fd_type = 8
	}
	for i, b := range []byte(e.Name()) {
		if i >= 255 {
			break
		}
		v.Fd_name[i] = int8(b)
	}
	return d.entry
}
func goClosedir(t *libc.TLS, p uintptr) int32 {
	compatMutex.Lock()
	d := compatDirs[p]
	delete(compatDirs, p)
	compatMutex.Unlock()
	if d == nil {
		return -1
	}
	libc.Xfree(t, d.entry)
	libc.Xfree(t, p)
	return 0
}
func goFscanf(t *libc.TLS, h, format, args uintptr) int32 {
	// The only stream scan in the inherited engine reads music/name identifiers.
	if libc.GoString(format) != "%d=%255s\n" {
		panic("unsupported stream scan")
	}
	var line strings.Builder
	for {
		c := goFgetc(t, h)
		if c < 0 || c == 10 {
			break
		}
		line.WriteByte(byte(c))
	}
	for {
		c := goFgetc(t, h)
		if c < 0 {
			break
		}
		if !strings.ContainsRune(" \t\r\n\v\f", rune(c)) {
			goFseek(t, h, -1, io.SeekCurrent)
			break
		}
	}
	parts := strings.SplitN(strings.TrimSpace(line.String()), "=", 2)
	if len(parts) != 2 {
		return -1
	}
	index, e := strconv.Atoi(parts[0])
	if e != nil {
		return 0
	}
	a := unsafe.Slice((*uintptr)(unsafe.Pointer(args)), 2)
	*(*int32)(unsafe.Pointer(a[0])) = int32(index)
	name := []byte(parts[1])
	if len(name) > 255 {
		name = name[:255]
	}
	dst := unsafe.Slice((*byte)(unsafe.Pointer(a[1])), len(name)+1)
	copy(dst, name)
	dst[len(name)] = 0
	return 2
}

var _ = fmt.Sprintf

func goStrnlen(t *libc.TLS, p uintptr, n uint64) uint64 {
	for i := uint64(0); i < n; i++ {
		if *(*byte)(unsafe.Pointer(p + uintptr(i))) == 0 {
			return i
		}
	}
	return n
}
func goStrncasecmp(t *libc.TLS, a, b uintptr, n uint64) int32 {
	for i := uint64(0); i < n; i++ {
		x, y := *(*byte)(unsafe.Pointer(a + uintptr(i))), *(*byte)(unsafe.Pointer(b + uintptr(i)))
		if x >= 'A' && x <= 'Z' {
			x += 32
		}
		if y >= 'A' && y <= 'Z' {
			y += 32
		}
		if x != y {
			return int32(x) - int32(y)
		}
		if x == 0 {
			return 0
		}
	}
	return 0
}
func goStrcasecmp(t *libc.TLS, a, b uintptr) int32 { return goStrncasecmp(t, a, b, ^uint64(0)) }
func goAssert(t *libc.TLS, expr, file uintptr, line int32, function uintptr) {
	panic(fmt.Sprintf("%s:%d: %s: assertion failed: %s", libc.GoString(file), line, libc.GoString(function), libc.GoString(expr)))
}
