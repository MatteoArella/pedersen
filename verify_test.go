// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package pedersen_test

import (
	"crypto/rand"
	"testing"

	"github.com/matteoarella/pedersen/big"

	"github.com/matteoarella/pedersen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pedersenVerifyTestCases struct {
	pedersenTestCases
	shares *pedersen.Shares
	error  error
}

func TestPedersenVerifyInvalid(t *testing.T) {
	group := getTestSchnorrGroup(t)

	testCases := []pedersenVerifyTestCases{
		{
			pedersenTestCases: pedersenTestCases{
				description: "nil abscissa",
				parameters: pedersenParameters{
					parts:     5,
					threshold: 3,
				},
				options: []pedersen.Option{
					pedersen.CyclicGroup(group),
				},
			},
			shares: &pedersen.Shares{
				Abscissae: []*big.Int{big.One(), nil, nil, big.One(), nil},
				Parts: [][]pedersen.SecretPart{
					{{SShare: big.One(), TShare: big.One()}},
					{{SShare: big.One(), TShare: big.One()}},
					{{SShare: big.One(), TShare: big.One()}},
					{{SShare: big.One(), TShare: big.One()}},
					{{SShare: big.One(), TShare: big.One()}},
				},
				Commitments: [][]*big.Int{
					{big.One(), big.One(), big.One()},
				},
			},
			error: pedersen.ErrNilAbscissa,
		},
		{
			pedersenTestCases: pedersenTestCases{
				description: "empty secret part",
				parameters: pedersenParameters{
					parts:     5,
					threshold: 3,
				},
				options: []pedersen.Option{
					pedersen.CyclicGroup(group),
				},
			},
			shares: &pedersen.Shares{
				Abscissae: []*big.Int{
					big.One(),
					big.One(),
					big.One(),
					big.One(),
					big.One(),
				},
				Parts: [][]pedersen.SecretPart{{{}}, {{}}, {{}}, {{}}, {{}}},
				Commitments: [][]*big.Int{
					{big.One(), big.One(), big.One()},
				},
			},
			error: pedersen.ErrInsufficientSharesParts,
		},
		{
			pedersenTestCases: pedersenTestCases{
				description: "wrong secret commitments len",
				parameters: pedersenParameters{
					parts:     5,
					threshold: 3,
				},
				options: []pedersen.Option{
					pedersen.CyclicGroup(group),
				},
			},
			shares: &pedersen.Shares{
				Abscissae: []*big.Int{
					big.One(),
					big.One(),
					big.One(),
					big.One(),
					big.One(),
				},
				Parts: [][]pedersen.SecretPart{
					{{SShare: big.One(), TShare: big.One()}},
					{{SShare: big.One(), TShare: big.One()}},
					{{SShare: big.One(), TShare: big.One()}},
					{{SShare: big.One(), TShare: big.One()}},
					{{SShare: big.One(), TShare: big.One()}},
				},
				Commitments: [][]*big.Int{
					{
						big.One(),
						big.One(),
					},
				},
			},
			error: pedersen.ErrInsufficientCommitments,
		},
	}

	for _, scenario := range testCases {
		t.Run(scenario.description, func(t *testing.T) {
			p, err := pedersen.NewPedersen(scenario.parameters.parts, scenario.parameters.threshold, scenario.options...)
			require.NoError(t, err)
			require.NotNil(t, p)

			err = p.VerifyShares(scenario.shares)
			assert.Error(t, err)
			assert.EqualError(t, err, scenario.error.Error())
		})
	}
}

func TestPedersenVerifyValid(t *testing.T) {
	group := getTestSchnorrGroup(t)

	randomSecret := make([]byte, 128)
	_, err := rand.Read(randomSecret)
	require.NoError(t, err)

	testCases := []pedersenSplitTestCases{
		{
			pedersenTestCases: pedersenTestCases{
				description: "valid small secret split",
				parameters: pedersenParameters{
					parts:     5,
					threshold: 3,
				},
				options: []pedersen.Option{
					pedersen.CyclicGroup(group),
				},
			},
			secret: []byte("s"),
		},
		{
			pedersenTestCases: pedersenTestCases{
				description: "valid big secret split",
				parameters: pedersenParameters{
					parts:     5,
					threshold: 3,
				},
				options: []pedersen.Option{
					pedersen.CyclicGroup(group),
				},
			},
			secret: randomSecret,
		},
		{
			pedersenTestCases: pedersenTestCases{
				description: "valid secret split parameters with abscissae",
				parameters: pedersenParameters{
					parts:     5,
					threshold: 3,
				},
				options: []pedersen.Option{
					pedersen.CyclicGroup(group),
				},
			},
			secret:    randomSecret,
			abscissae: []int{1, 2, 3, 4, 5},
		},
	}

	for _, scenario := range testCases {
		t.Run(scenario.description, func(t *testing.T) {
			p, err := pedersen.NewPedersen(scenario.parameters.parts, scenario.parameters.threshold, scenario.options...)
			require.NoError(t, err)
			require.NotNil(t, p)

			abscissae, err := convertIntAbscissae(scenario.abscissae)
			require.NoError(t, err)

			shares, err := p.Split(scenario.secret, abscissae)
			require.NoError(t, err)
			require.NotNil(t, shares)

			err = p.VerifyShares(shares)
			require.NoError(t, err)
		})
	}
}

func benchmarkVerifyCase(b *testing.B, groupSize, parts, threshold int) {
	b.Helper()

	group, err := pedersen.NewSchnorrGroup(groupSize)
	require.NoError(b, err)

	randomSecret := make([]byte, 256)
	_, err = rand.Read(randomSecret)
	require.NoError(b, err)

	p, err := pedersen.NewPedersen(parts, threshold, pedersen.CyclicGroup(group))
	require.NoError(b, err)
	require.NotNil(b, p)

	shares, err := p.Split(randomSecret, nil)
	require.NoError(b, err)
	require.NotNil(b, shares)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = p.VerifyShares(shares)
		require.NoError(b, err)
	}
}

func BenchmarkPedersenVerify_1024_5_3(b *testing.B) {
	benchmarkVerifyCase(b, 1024, 5, 3)
}

func BenchmarkPedersenVerify_1024_7_4(b *testing.B) {
	benchmarkVerifyCase(b, 1024, 7, 4)
}

func BenchmarkPedersenVerify_1024_10_5(b *testing.B) {
	benchmarkVerifyCase(b, 1024, 10, 5)
}

func BenchmarkPedersenVerify_2048_5_3(b *testing.B) {
	benchmarkVerifyCase(b, 2048, 5, 3)
}

func BenchmarkPedersenVerify_2048_7_4(b *testing.B) {
	benchmarkVerifyCase(b, 2048, 7, 4)
}

func BenchmarkPedersenVerify_2048_10_5(b *testing.B) {
	benchmarkVerifyCase(b, 2048, 10, 5)
}
