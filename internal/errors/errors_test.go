// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package errors_test

import (
	"errors"
	"testing"

	perrors "github.com/matteoarella/pedersen/internal/errors"
	"github.com/stretchr/testify/require"
)

func TestErrorsValid(t *testing.T) {
	err := errors.New("base error") //nolint:goerr113
	wrapped := perrors.WrapErrorf(err, "[Wrapped1]")
	require.EqualError(t, wrapped, "[Wrapped1]: base error")
	wrapped = perrors.WrapErrorf(wrapped, "[Wrapped2]")
	require.EqualError(t, wrapped, "[Wrapped2]: [Wrapped1]: base error")
	wrapped = perrors.UnwrapAll(wrapped)
	require.EqualError(t, wrapped, "base error")
}
