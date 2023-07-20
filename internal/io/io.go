// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package io

import (
	"errors"
	"io/fs"
	"path"
	"path/filepath"

	perrors "github.com/matteoarella/pedersen/internal/errors"
	"github.com/spf13/afero"
)

var (
	ErrUnknownFileExtension = errors.New("unknown file extension")
)

type IO interface {
	ReadFile(name string, v interface{}) error
	WriteFile(name string, v interface{}, perm fs.FileMode) error
	Ext() string
}

type BaseIO struct {
	Fs afero.Fs
}

func (b BaseIO) ReadFile(name, ext string) ([]byte, error) {
	if len(filepath.Ext(name)) < 1 {
		name += ext
	}

	data, err := afero.ReadFile(b.Fs, name)
	if err != nil {
		return nil, perrors.WrapErrorf(err, "[BaseIO.ReadFile]")
	}

	return data, nil
}

func (b BaseIO) WriteFile(name, ext string, data []byte, perm fs.FileMode) error {
	if filepath.Ext(name) != ext {
		name += ext
	}

	dir := path.Dir(name)
	// Make sure the directory permission has the executable bit set
	dirPerm := perm | 0o111

	err := b.Fs.MkdirAll(dir, dirPerm)
	if err != nil {
		return perrors.WrapErrorf(err, "[BaseIO.MkdirAll]")
	}

	err = afero.WriteFile(b.Fs, name, data, perm)
	if err != nil {
		return perrors.WrapErrorf(err, "[BaseIO.WriteFile]")
	}

	return nil
}
