// SPDX-License-Identifier: GPL-3.0-or-later
// Ebitengine is the device layer; the recovered engine owns gameplay timing.
package frontend

import (
	"bytes"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image"
	"image/png"
	"math"
	"Prince/internal/engine"
	"runtime/debug"
	"time"
)

type game struct {
	host                        *engine.Host
	root                        string
	args                        []string
	started, closing            bool
	done                        chan error
	result                      error
	image                       *ebiten.Image
	frameVersion, windowVersion uint64
	focus                       bool
	player                      *audio.Player
	windowWidth, windowHeight   int
	pads                        map[int32]*padState
}
type padState struct {
	axes    [6]int16
	buttons [15]bool
}

func Run(root string, args []string, iconPNG []byte) error {
	icon, err := png.Decode(bytes.NewReader(iconPNG))
	if err != nil {
		return fmt.Errorf("application icon: %w", err)
	}
	ebiten.SetWindowIcon([]image.Image{icon})
	h := engine.NewHost()
	engine.SetHost(h)
	g := &game{host: h, root: root, args: args, done: make(chan error, 1), focus: true, pads: map[int32]*padState{}}
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("PrinceGo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowClosingHandled(true)
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetTPS(120)
	err = ebiten.RunGame(g)
	if g.player != nil {
		g.player.Close()
	}
	h.Audio.Close()
	if err != nil {
		return err
	}
	return g.result
}
func (g *game) Update() error {
	g.input()
	if !g.started {
		g.started = true
		go func() {
			defer func() {
				if p := recover(); p != nil {
					g.done <- fmt.Errorf("engine panic: %v\n%s", p, debug.Stack())
				}
			}()
			g.done <- engine.Run(g.root, g.args)
		}()
	}
	select {
	case g.result = <-g.done:
		return ebiten.Termination
	default:
	}
	if ebiten.IsWindowBeingClosed() && !g.closing {
		g.closing = true
		g.host.Close()
	}
	win := g.host.Window()
	if win.Revision != g.windowVersion {
		g.windowVersion = win.Revision
		ebiten.SetWindowTitle(win.Title)
		if win.Width > 0 && win.Height > 0 && (win.Width != g.windowWidth || win.Height != g.windowHeight) {
			g.windowWidth, g.windowHeight = win.Width, win.Height
			ebiten.SetWindowSize(win.Width, win.Height)
		}
		ebiten.SetFullscreen(win.Fullscreen)
		if win.HideCursor {
			ebiten.SetCursorMode(ebiten.CursorModeHidden)
		} else {
			ebiten.SetCursorMode(ebiten.CursorModeVisible)
		}
		if win.Icon != nil {
			ebiten.SetWindowIcon([]image.Image{win.Icon})
		}
	}
	if g.player == nil {
		if rate := g.host.Audio.SampleRate(); rate > 0 {
			ctx := audio.NewContext(rate)
			p, e := ctx.NewPlayer(g.host.Audio)
			if e != nil {
				return e
			}
			p.SetBufferSize(time.Second * 1024 / time.Duration(rate))
			p.Play()
			g.player = p
		}
	}
	for _, r := range g.host.Rumbles() {
		ebiten.VibrateGamepad(ebiten.GamepadID(r.ID), &ebiten.VibrateGamepadOptions{Duration: r.Duration, StrongMagnitude: r.Strong, WeakMagnitude: r.Weak})
	}
	return nil
}
func (g *game) Layout(w, h int) (int, int) { g.host.Resize(w, h); return 320, 200 }
func (g *game) Draw(screen *ebiten.Image) {
	f := g.host.Frame()
	if f.Revision != g.frameVersion {
		g.frameVersion = f.Revision
		if f.Width > 0 && f.Height > 0 {
			if g.image == nil || g.image.Bounds().Dx() != f.Width || g.image.Bounds().Dy() != f.Height {
				if g.image != nil {
					g.image.Dispose()
				}
				g.image = ebiten.NewImage(f.Width, f.Height)
			}
			g.image.WritePixels(f.Pixels)
		}
	}
	if g.image != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(320/float64(g.image.Bounds().Dx()), 200/float64(g.image.Bounds().Dy()))
		screen.DrawImage(g.image, op)
	}
}
func (g *game) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, _ ebiten.GeoM) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	state := g.host.Window()
	aspect := state.Aspect
	if aspect <= 0 {
		aspect = 4.0 / 3
	}
	dh := float64(h)
	dw := dh * aspect
	if dw > float64(w) {
		dw = float64(w)
		dh = dw / aspect
	}
	if state.IntegerScale && dw >= 1600 && dh >= 1200 {
		scale := math.Floor(math.Min(dw/1600, dh/1200))
		dw, dh = 1600*scale, 1200*scale
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(dw/320, dh/200)
	op.GeoM.Translate((float64(w)-dw)/2, (float64(h)-dh)/2)
	op.Filter = ebiten.FilterNearest
	screen.DrawImage(offscreen, op)
}
func (g *game) input() {
	mod := uint16(0)
	for _, m := range []struct {
		k    ebiten.Key
		mask uint16
	}{{ebiten.KeyShiftLeft, engine.KMOD_LSHIFT}, {ebiten.KeyShiftRight, engine.KMOD_RSHIFT}, {ebiten.KeyControlLeft, engine.KMOD_LCTRL}, {ebiten.KeyControlRight, engine.KMOD_RCTRL}, {ebiten.KeyAltLeft, engine.KMOD_LALT}, {ebiten.KeyAltRight, engine.KMOD_RALT}, {ebiten.KeyMetaLeft, engine.KMOD_LGUI}, {ebiten.KeyMetaRight, engine.KMOD_RGUI}} {
		if ebiten.IsKeyPressed(m.k) {
			mod |= m.mask
		}
	}
	for _, key := range keys {
		duration := inpututil.KeyPressDuration(key.key)
		if duration == 1 {
			g.host.Key(key.scan, true, false, mod)
		} else if duration >= 48 && (duration-48)%4 == 0 {
			g.host.Key(key.scan, true, true, mod)
		}
		if inpututil.IsKeyJustReleased(key.key) {
			g.host.Key(key.scan, false, false, mod)
		}
	}
	focused := ebiten.IsFocused()
	if focused != g.focus {
		g.focus = focused
		g.host.Focus(focused)
	}
	if g.host.TextEnabled() {
		g.host.Text(string(ebiten.AppendInputChars(nil)))
	}
	ids := ebiten.AppendGamepadIDs(nil)
	list := make([]int32, len(ids))
	connected := map[int32]bool{}
	for i, id := range ids {
		list[i] = int32(id)
		connected[int32(id)] = true
	}
	g.host.Gamepads(list)
	for index, id := range ids {
		p := g.pads[int32(id)]
		if p == nil {
			p = &padState{}
			g.pads[int32(id)] = p
			g.host.ControllerDevice(int32(index), true)
		}
		standard := ebiten.IsStandardGamepadLayoutAvailable(id)
		for i := 0; i < 6; i++ {
			v := float64(0)
			if standard {
				if i < 4 {
					v = ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxis(i))
				} else {
					button := ebiten.StandardGamepadButtonFrontBottomLeft
					if i == 5 {
						button = ebiten.StandardGamepadButtonFrontBottomRight
					}
					v = ebiten.StandardGamepadButtonValue(id, button)
				}
			} else if i < ebiten.GamepadAxisCount(id) {
				v = ebiten.GamepadAxisValue(id, i)
			}
			axis := int16(math.Max(-32768, math.Min(32767, v*32767)))
			if axis != p.axes[i] {
				p.axes[i] = axis
				g.host.ControllerAxis(int32(id), uint8(i), axis)
			}
		}
		for i, button := range padButtons {
			pressed := false
			if standard {
				pressed = ebiten.IsStandardGamepadButtonPressed(id, button)
			} else if i < ebiten.GamepadButtonCount(id) {
				pressed = ebiten.IsGamepadButtonPressed(id, ebiten.GamepadButton(i))
			}
			if pressed != p.buttons[i] {
				p.buttons[i] = pressed
				g.host.ControllerButton(int32(id), uint8(i), pressed)
			}
		}
	}
	for id := range g.pads {
		if !connected[id] {
			delete(g.pads, id)
			g.host.ControllerDevice(id, false)
		}
	}
}

var padButtons = []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightBottom, ebiten.StandardGamepadButtonRightRight, ebiten.StandardGamepadButtonRightLeft, ebiten.StandardGamepadButtonRightTop, ebiten.StandardGamepadButtonCenterLeft, ebiten.StandardGamepadButtonCenterCenter, ebiten.StandardGamepadButtonCenterRight, ebiten.StandardGamepadButtonLeftStick, ebiten.StandardGamepadButtonRightStick, ebiten.StandardGamepadButtonFrontTopLeft, ebiten.StandardGamepadButtonFrontTopRight, ebiten.StandardGamepadButtonLeftTop, ebiten.StandardGamepadButtonLeftBottom, ebiten.StandardGamepadButtonLeftLeft, ebiten.StandardGamepadButtonLeftRight}
