// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package xml

import (
	"encoding/xml"
	iofs "io/fs"
	"path/filepath"

	perrors "github.com/matteoarella/pedersen/internal/errors"
	"github.com/matteoarella/pedersen/internal/io"
	"github.com/spf13/afero"
)

type xmlIO struct {
	io.BaseIO
}

func New(fs afero.Fs) io.IO {
	return xmlIO{
		BaseIO: io.BaseIO{
			Fs: fs,
		},
	}
}

func Ext() string {
	return ".xml"
}

func (x xmlIO) Ext() string {
	return Ext()
}

func (x xmlIO) ReadFile(name string, v interface{}) error {
	fileData, err := x.BaseIO.ReadFile(name, Ext())
	if err != nil {
		return perrors.WrapErrorf(err, "[xmlIO.ReadFile]")
	}

	return xml.Unmarshal(fileData, v)
}

func (x xmlIO) WriteFile(name string, v interface{}, perm iofs.FileMode) error {
	derData, err := xml.Marshal(v)
	if err != nil {
		return perrors.WrapErrorf(err, "[xmlIO.WriteFile]")
	}

	ext := filepath.Ext(name)
	if len(ext) < 1 {
		name += x.Ext()
	}

	return x.BaseIO.WriteFile(name, Ext(), derData, perm)
}
