// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package pedersen

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/matteoarella/pedersen/big"
)

const (
	defaultGroupPrimeBitLen = 128
	minThreshold            = 2
)

var (
	defaultConcLimit = runtime.NumCPU()
)

// errors
var (
	ErrInvalidOptions   = errors.New("invalid options")
	ErrInvalidThreshold = fmt.Errorf("threshold must be at least %d", minThreshold)
)

// Option represents an option for configuring a Pedersen struct.
type Option func(*Pedersen)

// The CyclicGroup option sets the cyclic group to be used.
func CyclicGroup(group *Group) Option {
	return func(p *Pedersen) {
		p.group = group
	}
}

// The ConcLimit option sets the maximum number of concurrent operations.
// If a negative number is provided, the number of concurrent operations
// is set to the number of CPUs.
func ConcLimit(concLimit int) Option {
	return func(p *Pedersen) {
		if concLimit < 1 {
			concLimit = defaultConcLimit
		}

		p.concLimit = concLimit
	}
}

// A Pedersen struct used for splitting, reconstructing, and verifying secrets.
type Pedersen struct {
	group *Group

	threshold int
	parts     int
	concLimit int
}

func (p *Pedersen) validate() error {
	if p.threshold < minThreshold {
		return ErrInvalidThreshold
	}

	if p.parts < p.threshold {
		return ErrInsufficientSharesParts
	}

	return nil
}

// NewPedersen creates a new Pedersen struct with the provided (threshold, parts) scheme.
// With such a scheme a secret is split into parts shares, of which at least threshold
// are required to reconstruct the secret.
// A new randomly generated cyclic group is used if none is provided.
func NewPedersen(parts, threshold int, options ...Option) (*Pedersen, error) {
	defaultPedersenOptions := []Option{
		ConcLimit(defaultConcLimit),
	}

	p := &Pedersen{
		parts:     parts,
		threshold: threshold,
	}

	for _, o := range defaultPedersenOptions {
		o(p)
	}

	// override default options
	for _, o := range options {
		o(p)
	}

	if p.group != nil {
		if err := p.group.validate(); err != nil {
			return nil, err
		}
	} else {
		group, err := NewSchnorrGroup(defaultGroupPrimeBitLen)
		if err != nil {
			return nil, err
		}

		p.group = group
	}

	if err := p.validate(); err != nil {
		return nil, err
	}

	return p, nil
}

// GetThreshold returns the threshold of the Pedersen struct.
func (p *Pedersen) GetThreshold() int {
	return p.threshold
}

// GetThreshold returns the parts of the Pedersen struct.
func (p *Pedersen) GetParts() int {
	return p.parts
}

// GetGroup returns the cyclic group of the Pedersen struct.
func (p *Pedersen) GetGroup() *Group {
	return p.group
}

// GetConcLimit returns the maximum number of concurrent operations
// of the Pedersen struct.
func (p *Pedersen) GetConcLimit() int {
	return p.concLimit
}

func (p *Pedersen) adjustConcLimit(num int) int {
	concLimit := p.GetConcLimit()
	if concLimit > num {
		return num
	}

	return concLimit
}

func (p *Pedersen) balanceIndices(length, numChunks int) []chunkRange {
	var ranges []chunkRange

	chunkSize := (length + numChunks - 1) / numChunks

	for start := 0; start < length; start += chunkSize {
		end := start + chunkSize

		if end > length {
			end = length
		}

		ranges = append(ranges, chunkRange{
			start: start,
			end:   end,
		})
	}

	return ranges
}

func (p *Pedersen) commit(
	mont *big.MontgomeryContext,
	ctx *big.IntContext,
	s *big.Int,
	t *big.Int,
) (*big.Int, error) {
	gs, err := big.NewInt()
	if err != nil {
		return nil, err
	}
	ht, err := ctx.GetInt()
	if err != nil {
		return nil, err
	}

	if err := gs.ModExpMont(mont, ctx, p.group.G, s, p.group.P); err != nil {
		return nil, err
	}

	if err := ht.ModExpMont(mont, ctx, p.group.H, t, p.group.P); err != nil {
		return nil, err
	}

	if err := gs.ModMul(ctx, gs, ht, p.group.P); err != nil {
		return nil, err
	}

	return gs, nil
}
