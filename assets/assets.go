// SPDX-License-Identifier: GPL-3.0-or-later
package assets

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func Materialize(profile string) (string, error) {
	root := filepath.Join(profile, "assets-v2")
	if err := os.MkdirAll(root, 0700); err != nil {
		return "", err
	}
	err := fs.WalkDir(Files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		target := filepath.Join(root, filepath.FromSlash(path))
		if d.IsDir() {
			return os.MkdirAll(target, 0700)
		}
		b, err := fs.ReadFile(Files, path)
		if err != nil {
			return err
		}
		if old, e := os.ReadFile(target); e == nil && sha256.Sum256(old) == sha256.Sum256(b) {
			return nil
		}
		// Atomic replacement allows two independent starts sharing a profile.
		tmp, err := os.CreateTemp(filepath.Dir(target), ".asset-*")
		if err != nil {
			return err
		}
		name := tmp.Name()
		defer os.Remove(name)
		if _, err = tmp.Write(b); err != nil {
			tmp.Close()
			return err
		}
		if err = tmp.Close(); err != nil {
			return err
		}
		if err = os.Rename(name, target); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		return nil
	})
	return root, err
}
