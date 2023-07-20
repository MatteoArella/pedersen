// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package pedersen

import (
	"encoding/json"

	"github.com/matteoarella/pedersen/big"
)

// SecretPart represents a secret part associated to a shareholder.
type SecretPart struct {
	SShare *big.Int
	TShare *big.Int
}

// Shares represents the shares obtained from splitting a secret.
type Shares struct {
	// Abscissae is the abscissae vector used for computing the ordinate values of
	// the secret parts.
	// There is one abscissa for each shareholder, so if shareholderIdx represents
	// the index of one shareholder, Abscissae[shareholderIdx] is the abscissa
	// related to that shareholder.
	Abscissae []*big.Int

	// Parts is the matrix of secret parts.
	// If the secret that has to be split is not representable in the cyclic group,
	// the secret is split into chunks, and each chunk is split into secret parts according
	// to Pedersen verifiable secret sharing.
	// The first index of Parts represents the shareholder index, while the second index
	// represents the chunk index (Parts[shareholderIdx][chunkIdx]).
	Parts [][]SecretPart

	// Commitments is the matrix of commitments.
	// The first index of Commitments represents the chunk index so Commitments[chunkIdx]
	// is the vector of commitments related to the chunk with index chunkIdx.
	Commitments [][]*big.Int
}

// Returns a string representation of a SecretPart struct.
func (p *SecretPart) String() string {
	data, _ := json.Marshal(p)
	return string(data)
}

// Returns a string representation of a Shares struct.
func (s *Shares) String() string {
	data, _ := json.Marshal(s)
	return string(data)
}
