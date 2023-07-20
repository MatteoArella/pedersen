// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package yaml

import (
	iofs "io/fs"

	perrors "github.com/matteoarella/pedersen/internal/errors"
	"github.com/matteoarella/pedersen/internal/io"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

type yamlIO struct {
	io.BaseIO
}

func New(fs afero.Fs) io.IO {
	return yamlIO{
		BaseIO: io.BaseIO{
			Fs: fs,
		},
	}
}

func Ext() string {
	return ".yaml"
}

func (y yamlIO) Ext() string {
	return Ext()
}

func (y yamlIO) ReadFile(name string, v interface{}) error {
	fileData, err := y.BaseIO.ReadFile(name, Ext())
	if err != nil {
		return perrors.WrapErrorf(err, "[yamlIO.ReadFile]")
	}

	return yaml.Unmarshal(fileData, v)
}

func (y yamlIO) WriteFile(name string, v interface{}, perm iofs.FileMode) error {
	yamlData, err := yaml.Marshal(v)
	if err != nil {
		return perrors.WrapErrorf(err, "[yamlIO.WriteFile]")
	}

	return y.BaseIO.WriteFile(name, Ext(), yamlData, perm)
}
