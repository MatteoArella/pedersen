// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package cmd_test

import (
	"testing"

	"github.com/spf13/afero"
)

type CliCmdTestCase struct {
	scenario   string
	args       []string
	err        error
	validateFn func(t *testing.T, fs afero.Fs)
}
