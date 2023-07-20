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

	"github.com/stretchr/testify/require"
)

type pedersenParameters struct {
	parts     int
	threshold int
}

type pedersenTestCases struct {
	description string
	parameters  pedersenParameters
	options     []pedersen.Option
}

func getTestSchnorrGroup(t *testing.T) *pedersen.Group {
	P, err := big.NewInt()
	require.NoError(t, err)
	Q, err := big.NewInt()
	require.NoError(t, err)
	G, err := big.NewInt()
	require.NoError(t, err)
	H, err := big.NewInt()
	require.NoError(t, err)

	err = P.SetDecString("17634709279010524619")
	require.NoError(t, err)
	err = Q.SetDecString("8817354639505262309")
	require.NoError(t, err)
	err = G.SetDecString("8414335786771157015")
	require.NoError(t, err)
	err = H.SetDecString("15078279289296123424")
	require.NoError(t, err)

	return &pedersen.Group{
		P: P,
		Q: Q,
		G: G,
		H: H,
	}
}

func TestPedersenInvalid(t *testing.T) {
	group := getTestSchnorrGroup(t)

	for _, scenario := range []pedersenTestCases{
		{
			parameters: pedersenParameters{
				parts:     0,
				threshold: 0,
			},
			options: []pedersen.Option{
				pedersen.CyclicGroup(group),
			},
			description: "zero threshold and zero parts",
		},
		{
			parameters: pedersenParameters{
				parts:     3,
				threshold: 5,
			},
			options: []pedersen.Option{
				pedersen.CyclicGroup(group),
			},
			description: "threshold greater than parts",
		},
		{
			parameters: pedersenParameters{
				parts:     5,
				threshold: 1,
			},
			options: []pedersen.Option{
				pedersen.CyclicGroup(group),
			},
			description: "threshold value too small",
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			_, err := pedersen.NewPedersen(scenario.parameters.parts, scenario.parameters.threshold, scenario.options...)
			require.Error(t, err)
		})
	}
}

func TestPedersenValid(t *testing.T) {
	group := getTestSchnorrGroup(t)

	for _, scenario := range []pedersenTestCases{
		{
			parameters: pedersenParameters{
				parts:     5,
				threshold: 3,
			},
			options: []pedersen.Option{
				pedersen.CyclicGroup(group),
			},
			description: "valid pedersen parameters",
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			p, err := pedersen.NewPedersen(scenario.parameters.parts, scenario.parameters.threshold, scenario.options...)
			require.NoError(t, err)

			require.NotNil(t, p)
		})
	}
}
