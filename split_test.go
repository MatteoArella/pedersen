// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package pedersen_test

import (
	"crypto/rand"
	"testing"

	"github.com/matteoarella/pedersen"
	"github.com/matteoarella/pedersen/big"

	"github.com/stretchr/testify/require"
)

type pedersenSplitTestCases struct {
	pedersenTestCases
	secret    []byte
	abscissae []int
}

func convertIntAbscissae(abscissae []int) ([]*big.Int, error) {
	if abscissae == nil {
		return nil, nil
	}

	result := make([]*big.Int, len(abscissae))

	for idx, ab := range abscissae {
		abscissa, err := big.NewInt()
		if err != nil {
			return nil, err
		}

		if err := abscissa.SetUInt64(uint64(ab)); err != nil {
			return nil, err
		}

		result[idx] = abscissa
	}

	return result, nil
}

func TestPedersenSplitInvalid(t *testing.T) {
	group := getTestSchnorrGroup(t)

	testCases := []pedersenSplitTestCases{
		{
			pedersenTestCases: pedersenTestCases{
				description: "nil secret",
				parameters: pedersenParameters{
					parts:     5,
					threshold: 3,
				},
				options: []pedersen.Option{
					pedersen.CyclicGroup(group),
				},
			},
			secret: nil,
		},
		{
			pedersenTestCases: pedersenTestCases{
				description: "wrong abscissae size",
				parameters: pedersenParameters{
					parts:     5,
					threshold: 3,
				},
				options: []pedersen.Option{
					pedersen.CyclicGroup(group),
				},
			},
			secret:    []byte("test"),
			abscissae: []int{1, 2, 3, 4},
		},
	}

	for _, scenario := range testCases {
		t.Run(scenario.description, func(t *testing.T) {
			p, err := pedersen.NewPedersen(scenario.parameters.parts, scenario.parameters.threshold, scenario.options...)
			require.NoError(t, err)
			require.NotNil(t, p)

			abscissae, err := convertIntAbscissae(scenario.abscissae)
			require.NoError(t, err)

			_, err = p.Split(scenario.secret, abscissae)
			require.Error(t, err)
		})
	}
}

func TestPedersenSplitValid(t *testing.T) {
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
			secret: []byte("test"),
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
			secret:    []byte("test"),
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
		})
	}
}

func benchmarkSplitCase(b *testing.B, groupSize, parts, threshold int) {
	b.Helper()

	group, err := pedersen.NewSchnorrGroup(groupSize)
	require.NoError(b, err)

	randomSecret := make([]byte, 256)
	_, err = rand.Read(randomSecret)
	require.NoError(b, err)

	p, err := pedersen.NewPedersen(parts, threshold, pedersen.CyclicGroup(group))
	require.NoError(b, err)
	require.NotNil(b, p)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = p.Split(randomSecret, nil)
		require.NoError(b, err)
	}
}

func BenchmarkPedersenSplit_1024_5_3(b *testing.B) {
	benchmarkSplitCase(b, 1024, 5, 3)
}

func BenchmarkPedersenSplit_1024_7_4(b *testing.B) {
	benchmarkSplitCase(b, 1024, 7, 4)
}

func BenchmarkPedersenSplit_1024_10_5(b *testing.B) {
	benchmarkSplitCase(b, 1024, 10, 5)
}

func BenchmarkPedersenSplit_2048_5_3(b *testing.B) {
	benchmarkSplitCase(b, 2048, 5, 3)
}

func BenchmarkPedersenSplit_2048_7_4(b *testing.B) {
	benchmarkSplitCase(b, 2048, 7, 4)
}

func BenchmarkPedersenSplit_2048_10_5(b *testing.B) {
	benchmarkSplitCase(b, 2048, 10, 5)
}
