// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package big_test

import (
	"testing"

	"github.com/matteoarella/pedersen/big"
	"github.com/stretchr/testify/require"
)

func TestBnValid(t *testing.T) {
	t.Run("valid bn", func(t *testing.T) {
		a, err := big.NewInt()
		require.NoError(t, err)
		err = a.SetUInt64(1)
		require.NoError(t, err)

		b, err := big.NewInt()
		require.NoError(t, err)
		err = b.SetUInt64(2)
		require.NoError(t, err)

		sum, err := big.NewInt()
		require.NoError(t, err)
		err = sum.SetUInt64(3)
		require.NoError(t, err)

		s, err := big.NewInt()
		require.NoError(t, err)
		err = s.Set(sum)
		require.NoError(t, err)

		err = b.Add(a, b)
		require.NoError(t, err)

		require.Equal(t, 0, sum.Cmp(b))
		require.Equal(t, 0, sum.Cmp(s))

		dec, err := big.NewInt()
		require.NoError(t, err)

		err = dec.SetDecString("10")
		require.NoError(t, err)

		expected, err := big.NewInt()
		require.NoError(t, err)
		err = expected.SetUInt64(10)
		require.NoError(t, err)

		require.Equal(t, 0, dec.Cmp(expected))

		mask, err := big.NewInt()
		require.NoError(t, err)
		err = mask.SetUInt64(5)
		require.NoError(t, err)

		ctx, err := big.NewIntContext()
		require.NoError(t, err)

		err = mask.Mod(ctx, mask, sum)
		require.NoError(t, err)

		err = expected.SetUInt64(3)
		require.NoError(t, err)

		err = a.SetUInt64(2)
		require.NoError(t, err)
		err = b.SetUInt64(3)
		require.NoError(t, err)
		err = s.SetUInt64(5)
		require.NoError(t, err)

		err = a.ModExp(ctx, a, b, s)
		require.NoError(t, err)

		require.Equal(t, 0, expected.Cmp(a))
	})
}

func TestExpMontValid(t *testing.T) {
	t.Run("valid exp mont bn", func(t *testing.T) {
		a, err := big.NewInt()
		require.NoError(t, err)
		err = a.SetUInt64(2)
		require.NoError(t, err)

		b, err := big.NewInt()
		require.NoError(t, err)
		err = b.SetUInt64(3)
		require.NoError(t, err)

		m, err := big.NewInt()
		require.NoError(t, err)
		err = m.SetUInt64(5)
		require.NoError(t, err)

		ctx, err := big.NewIntContext()
		require.NoError(t, err)
		defer ctx.Destroy()

		mont, err := big.NewMontgomeryContext()
		require.NoError(t, err)
		defer mont.Destroy()

		err = mont.Set(m, ctx)
		require.NoError(t, err)

		res, err := big.NewInt()
		require.NoError(t, err)

		err = res.ModExpMont(mont, ctx, a, b, m)
		require.NoError(t, err)

		expected, err := big.NewInt()
		require.NoError(t, err)
		err = expected.SetUInt64(3)
		require.NoError(t, err)

		require.Equal(t, 0, res.Cmp(expected))
	})
}
