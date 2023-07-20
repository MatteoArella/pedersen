// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package pedersen

import (
	"github.com/matteoarella/pedersen/big"
	"golang.org/x/sync/errgroup"
)

type combineValue struct {
	index int
	value *big.Int
}

func (p *Pedersen) combine(
	ctx *big.IntContext,
	index int,
	abscissae []*big.Int,
	parts []SecretPart,
) (combineValue, error) {
	var (
		xSamples []*big.Int
		ySamples []*big.Int
	)

	for idx, p := range parts {
		if (SecretPart{}) == p {
			continue
		}

		ySamples = append(ySamples, p.SShare)
		xSamples = append(xSamples, abscissae[idx])
	}

	ctx.Attach()
	defer ctx.Detach()

	zero, err := ctx.GetInt()
	if err != nil {
		return combineValue{}, err
	}

	if err := zero.SetUInt64(0); err != nil {
		return combineValue{}, err
	}

	secret, err := interpolatePolynomial(ctx, xSamples, ySamples, zero, p.group.Q)
	if err != nil {
		return combineValue{}, err
	}

	return combineValue{
		index: index,
		value: secret,
	}, nil
}

func bigIntUnpadding(ctx *big.IntContext, n *big.Int) ([]byte, error) {
	ctx.Attach()
	defer ctx.Detach()

	// extract leading zeros info from trailing bytes
	mask, err := ctx.GetInt()
	if err != nil {
		return nil, err
	}
	if err := mask.SetUInt64(0x00000000FFFFFFFF); err != nil {
		return nil, err
	}

	if err := mask.And(n, mask); err != nil {
		return nil, err
	}

	zeros := mask.Uint64()

	// cut trailer
	if err := n.Rsh(n, zerosInfoSizeBytes*8); err != nil {
		return nil, err
	}

	res := make([]byte, zeros)

	for i := 0; i < int(zeros); i++ {
		res[i] = 0
	}

	nBytes, err := n.Bytes()
	if err != nil {
		return nil, err
	}

	return append(res, nBytes...), nil // nozero
}

// Combine combines the secret shares into the original secret.
func (p *Pedersen) Combine(shares *Shares) ([]byte, error) {
	err := p.validateShares(shares)
	if err != nil {
		return nil, err
	}

	splittedLen := len(shares.Parts[0])
	values := make([]*big.Int, splittedLen)
	concLimit := p.adjustConcLimit(splittedLen)
	chunksIndex := p.balanceIndices(splittedLen, concLimit)
	group := errgroup.Group{}
	group.SetLimit(concLimit)

	for _, chunk := range chunksIndex {
		chunk := chunk

		group.Go(func() error {
			ctx, err := big.NewIntContext()
			if err != nil {
				return err
			}
			defer ctx.Destroy()

			parts := make([]SecretPart, p.parts)

			for idx := chunk.start; idx < chunk.end; idx++ {
				for shareIdx := 0; shareIdx < p.parts; shareIdx++ {
					parts[shareIdx] = shares.Parts[shareIdx][idx]
				}

				value, err := p.combine(ctx, idx, shares.Abscissae, parts)
				if err != nil {
					return err
				}

				values[value.index] = value.value
			}

			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		return nil, err
	}

	var res []byte
	ctx, err := big.NewIntContext()
	if err != nil {
		return nil, err
	}
	defer ctx.Destroy()

	for i := 0; i < splittedLen; i++ {
		chunk, err := bigIntUnpadding(ctx, values[i])
		if err != nil {
			return nil, err
		}

		res = append(res, chunk...)
	}

	return res, nil
}
