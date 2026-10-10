// SPDX-License-Identifier: GPL-3.0-or-later
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"Prince/assets"
	"Prince/internal/frontend"
	"runtime"
)

func init() { runtime.LockOSThread() }

//go:embed icon.png
var appIcon []byte

//go:generate go run github.com/akavel/rsrc@v0.10.2 -ico icon.ico -arch amd64 -o icon_windows_amd64.syso

func main() {
	profile := flag.String("profile", "", "save directory (default: user config/PrinceGo)")
	version := flag.Bool("version", false, "show implementation status")
	flag.Parse()
	if *version {
		fmt.Println("PrinceGo 1.1.0 — Go engine")
		return
	}
	if *profile == "" {
		d, err := os.UserConfigDir()
		if err != nil {
			fatal(err)
		}
		*profile = filepath.Join(d, "PrinceGo")
	}
	p, err := filepath.Abs(*profile)
	if err != nil {
		fatal(err)
	}
	if len(p) > 180 {
		fatal(fmt.Errorf("profile path too long for the inherited engine (max 180 UTF-8 bytes)"))
	}
	if err = os.MkdirAll(p, 0700); err != nil {
		fatal(err)
	}
	root, err := assets.Materialize(p)
	if err != nil {
		fatal(err)
	}
	if err = os.Setenv("SDLPOP_SAVE_PATH", p); err != nil {
		fatal(err)
	}
	args := append([]string{"Prince"}, flag.Args()...)
	if err = frontend.Run(root, args, appIcon); err != nil {
		fatal(err)
	}
}

func fatal(err error) { fmt.Fprintln(os.Stderr, "PrinceGo:", err); os.Exit(1) }
