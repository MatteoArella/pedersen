// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package schema

import (
	"github.com/matteoarella/pedersen"
	"github.com/matteoarella/pedersen/big"
)

type Group struct {
	P *big.Int `json:"p" yaml:"p" xml:"p"`
	Q *big.Int `json:"q" yaml:"q" xml:"q"`
	G *big.Int `json:"g" yaml:"g" xml:"g"`
	H *big.Int `json:"h" yaml:"h" xml:"h"`
}

type Shares struct {
	Abscissa *big.Int              `json:"abscissa" yaml:"abscissa" xml:"abscissa"`
	Parts    []pedersen.SecretPart `json:"parts" yaml:"parts" xml:"parts"`
}

type Commitments struct {
	Commitments [][]*big.Int `json:"commitments" yaml:"commitments" xml:"commitments"`
}
