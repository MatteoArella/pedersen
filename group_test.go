// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package pedersen_test

import (
	"testing"

	"github.com/matteoarella/pedersen"

	"github.com/matteoarella/pedersen/big"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validateGenerator(t *testing.T, ctx *big.IntContext, group *pedersen.Group) {
	exp, err := big.NewInt()
	require.NoError(t, err)

	err = exp.Sub(group.P, big.One())
	require.NoError(t, err)

	err = exp.Div(ctx, exp, group.Q)
	require.NoError(t, err)

	expected, err := big.NewInt()
	require.NoError(t, err)

	err = expected.ModExp(ctx, group.G, exp, group.P)
	require.NoError(t, err)

	// g^((p-1)/q) mod p != 1
	assert.EqualValues(t, true, expected.Cmp(big.One()) != 0)

	// g^q mod p = 1
	g, err := big.NewInt()
	require.NoError(t, err)
	err = g.ModExp(ctx, group.G, group.Q, group.P)
	require.NoError(t, err)

	assert.EqualValues(t, true, g.Cmp(big.One()) == 0)
}

func TestSchnorrGroup(t *testing.T) {
	group, err := pedersen.NewSchnorrGroup(64)
	require.NoError(t, err)

	ctx, err := big.NewIntContext()
	require.NoError(t, err)
	defer ctx.Destroy()

	validateGenerator(t, ctx, group)
}
