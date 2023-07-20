// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package pedersen_test

import (
	"crypto/rand"
	mrand "math/rand"
	"testing"
	"time"

	"github.com/matteoarella/pedersen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pedersenCombineTestCases struct {
	pedersenTestCases
	shares *pedersen.Shares
	secret []byte
}

func TestPedersenCombineInvalid(t *testing.T) {
	group := getTestSchnorrGroup(t)

	testCases := []pedersenCombineTestCases{
		{
			pedersenTestCases: pedersenTestCases{
				description: "nil shares",
				parameters: pedersenParameters{
					parts:     5,
					threshold: 3,
				},
				options: []pedersen.Option{
					pedersen.CyclicGroup(group),
				},
			},
			shares: nil,
		},
		{
			pedersenTestCases: pedersenTestCases{
				description: "wrong shares parts len",
				parameters: pedersenParameters{
					parts:     5,
					threshold: 3,
				},
				options: []pedersen.Option{
					pedersen.CyclicGroup(group),
				},
			},
			shares: &pedersen.Shares{
				Parts: [][]pedersen.SecretPart{
					{
						pedersen.SecretPart{},
						pedersen.SecretPart{},
					},
				},
			},
		},
	}

	for _, scenario := range testCases {
		t.Run(scenario.description, func(t *testing.T) {
			p, err := pedersen.NewPedersen(scenario.parameters.parts, scenario.parameters.threshold, scenario.options...)
			require.NoError(t, err)
			require.NotNil(t, p)

			_, err = p.Combine(scenario.shares)
			require.Error(t, err)
		})
	}
}

func getRandomIndexSubset(n, size int) []int {
	mrand.Seed(time.Now().Unix())
	return mrand.Perm(n)[0:size]
}

func getSharesSubset(s *pedersen.Shares, threshold uint) *pedersen.Shares {
	n := len(s.Parts)
	shares := &pedersen.Shares{
		Abscissae:   s.Abscissae,
		Parts:       make([][]pedersen.SecretPart, n),
		Commitments: s.Commitments,
	}

	chunksCount := len(s.Parts[0])

	for idx := 0; idx < n; idx++ {
		shares.Parts[idx] = make([]pedersen.SecretPart, chunksCount)
	}

	for chunkIndex := 0; chunkIndex < chunksCount; chunkIndex++ {
		indexes := getRandomIndexSubset(n, int(threshold))

		for _, idx := range indexes {
			shares.Parts[idx][chunkIndex] = s.Parts[idx][chunkIndex]
		}
	}

	return shares
}

func TestPedersenCombineValid(t *testing.T) {
	group := getTestSchnorrGroup(t)

	randomSecret := make([]byte, 128)
	_, err := rand.Read(randomSecret)
	require.NoError(t, err)

	testCases := []pedersenCombineTestCases{
		{
			pedersenTestCases: pedersenTestCases{
				description: "valid combine random secret",
				parameters: pedersenParameters{
					parts:     10,
					threshold: 5,
				},
				options: []pedersen.Option{
					pedersen.CyclicGroup(group),
				},
			},
			secret: randomSecret,
		},
		{
			pedersenTestCases: pedersenTestCases{
				description: "valid combine secret with heading zeros",
				parameters: pedersenParameters{
					parts:     10,
					threshold: 5,
				},
				options: []pedersen.Option{
					pedersen.CyclicGroup(group),
				},
			},
			secret: []byte{0, 0, 0, 1, 2, 3, 4, 5},
		},
		{
			pedersenTestCases: pedersenTestCases{
				description: "valid combine secret with heading zero chunks",
				parameters: pedersenParameters{
					parts:     10,
					threshold: 5,
				},
				options: []pedersen.Option{
					pedersen.CyclicGroup(group),
				},
			},
			secret: []byte{0x0, 0x0, 0x2d, 0x33, 0x0, 0x0, 0xe7, 0x0, 0x0, 0x1c, 0x82, 0xa4, 0x4c, 0xcb, 0x11, 0x88},
		},
	}

	for _, scenario := range testCases {
		t.Run(scenario.description, func(t *testing.T) {
			p, err := pedersen.NewPedersen(scenario.parameters.parts, scenario.parameters.threshold, scenario.options...)
			require.NoError(t, err)
			require.NotNil(t, p)

			shares, err := p.Split(scenario.secret, nil)
			require.NoError(t, err)
			require.NotNil(t, shares)

			// combine full shares
			secret, err := p.Combine(shares)
			require.NoError(t, err)
			require.NotNil(t, secret)
			assert.Equal(t, scenario.secret, secret)

			// combine subset of shares
			subset := getSharesSubset(shares, uint(p.GetThreshold()))

			secret, err = p.Combine(subset)
			require.NoError(t, err)
			require.NotNil(t, secret)
			assert.Equal(t, scenario.secret, secret)
		})
	}
}

func benchmarkCombineCase(b *testing.B, groupSize, parts, threshold int) {
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
		_, err = p.Combine(shares)
		require.NoError(b, err)
	}
}

func BenchmarkPedersenCombine_1024_5_3(b *testing.B) {
	benchmarkCombineCase(b, 1024, 5, 3)
}

func BenchmarkPedersenCombine_1024_7_4(b *testing.B) {
	benchmarkCombineCase(b, 1024, 7, 4)
}

func BenchmarkPedersenCombine_1024_10_5(b *testing.B) {
	benchmarkCombineCase(b, 1024, 10, 5)
}

func BenchmarkPedersenCombine_2048_5_3(b *testing.B) {
	benchmarkCombineCase(b, 2048, 5, 3)
}

func BenchmarkPedersenCombine_2048_7_4(b *testing.B) {
	benchmarkCombineCase(b, 2048, 7, 4)
}

func BenchmarkPedersenCombine_2048_10_5(b *testing.B) {
	benchmarkCombineCase(b, 2048, 10, 5)
}
