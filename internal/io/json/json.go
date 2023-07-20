// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package json

import (
	"encoding/json"
	iofs "io/fs"
	"path/filepath"

	perrors "github.com/matteoarella/pedersen/internal/errors"
	"github.com/matteoarella/pedersen/internal/io"
	"github.com/spf13/afero"
)

type jsonIO struct {
	io.BaseIO
}

func New(fs afero.Fs) io.IO {
	return jsonIO{
		BaseIO: io.BaseIO{
			Fs: fs,
		},
	}
}

func Ext() string {
	return ".json"
}

func (j jsonIO) Ext() string {
	return Ext()
}

func (j jsonIO) ReadFile(name string, v interface{}) error {
	fileData, err := j.BaseIO.ReadFile(name, Ext())
	if err != nil {
		return perrors.WrapErrorf(err, "[jsonIO.ReadFile]")
	}

	return json.Unmarshal(fileData, v)
}

func (j jsonIO) WriteFile(name string, v interface{}, perm iofs.FileMode) error {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return perrors.WrapErrorf(err, "[jsonIO.WriteFile]")
	}

	if len(filepath.Ext(name)) < 1 {
		name += j.Ext()
	}

	return j.BaseIO.WriteFile(name, Ext(), jsonData, perm)
}
