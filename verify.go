// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package pedersen

import (
	"errors"

	"github.com/matteoarella/pedersen/big"
	"golang.org/x/sync/errgroup"
)

var (
	ErrNilAbscissa             = errors.New("abscissa cannot be nil")
	ErrNilShares               = errors.New("shares cannot be nil")
	ErrNilShare                = errors.New("s_share or t_share cannot be nil")
	ErrNilCommitment           = errors.New("commitment cannot be nil")
	ErrInsufficientSharesParts = errors.New("insufficient shares parts")
	ErrInsufficientCommitments = errors.New("commitments length cannot be different from threshold")
	ErrWrongSharesLen          = errors.New("shares parts length and commitments parts length must be equal")
	ErrWrongSecretPart         = errors.New("wrong secret part")
)

// validateShares validates if the provided shares have a correct shape.
func (p *Pedersen) validateShares(s *Shares) error {
	if s == nil {
		return ErrNilShares
	}

	if len(s.Abscissae) != p.parts {
		return ErrInsufficientAbscissae
	}

	if len(s.Parts) != p.parts {
		return ErrInsufficientSharesParts
	}

	for i := 0; i < p.parts; i++ {
		if s.Abscissae[i] == nil {
			return ErrNilAbscissa
		}
	}

	partsCount := len(s.Parts[0])

	if len(s.Commitments) != partsCount {
		return ErrWrongSharesLen
	}

	for partIdx := 0; partIdx < partsCount; partIdx++ {
		if len(s.Commitments[partIdx]) != p.threshold {
			return ErrInsufficientCommitments
		}

		parts := 0

		for i := 0; i < p.parts; i++ {
			if len(s.Parts[i]) != partsCount {
				return ErrWrongSharesLen
			}

			if (SecretPart{}) == s.Parts[i][partIdx] {
				continue
			}

			if s.Parts[i][partIdx].SShare == nil || s.Parts[i][partIdx].TShare == nil {
				return ErrNilShare
			}

			parts++
		}

		if parts < p.threshold {
			return ErrInsufficientSharesParts
		}

		for i := 0; i < p.threshold; i++ {
			if s.Commitments[partIdx][i] == nil {
				return ErrNilCommitment
			}
		}
	}

	return nil
}

func (p *Pedersen) vandermondeAbscissa(ctx *big.IntContext,
	abscissa *big.Int,
) ([]*big.Int, error) {
	abscissae := make([]*big.Int, p.threshold)

	abscissae[0] = big.One()

	for i := 1; i < p.threshold; i++ {
		a, err := ctx.GetInt()
		if err != nil {
			return nil, err
		}

		if err := a.ModMul(ctx, abscissae[i-1], abscissa, p.group.Q); err != nil {
			return nil, err
		}

		abscissae[i] = a
	}

	return abscissae, nil
}

func (p *Pedersen) verifyWithContext(mont *big.MontgomeryContext,
	ctx *big.IntContext,
	vandermondeAbscissa []*big.Int,
	part SecretPart,
	commitments []*big.Int,
) error {
	if part.SShare == nil || part.TShare == nil {
		return ErrNilShare
	}

	if len(commitments) != p.threshold {
		return ErrInsufficientCommitments
	}

	ctx.Attach()
	defer ctx.Detach()

	rhs, err := ctx.GetInt()
	if err != nil {
		return err
	}

	if err := rhs.Set(commitments[0]); err != nil {
		return err
	}

	// rhs = c_0 * c_1^x * ... * c_j^{x^j}
	for j := 1; j < p.threshold; j++ {
		term, err := ctx.GetInt()
		if err != nil {
			return err
		}

		if err := term.ModExpMont(mont, ctx, commitments[j], vandermondeAbscissa[j], p.group.P); err != nil {
			return err
		}

		if err := rhs.ModMul(ctx, rhs, term, p.group.P); err != nil {
			return err
		}
	}

	lhs, err := p.commit(mont, ctx, part.SShare, part.TShare)
	if err != nil {
		return err
	}

	equals, err := lhs.ConstantTimeEq(rhs)
	if err != nil {
		return err
	}

	if !equals {
		return ErrWrongSecretPart
	}

	return nil
}

// Verify verifies if the provided secret part is valid, according to the provided abscissa value and
// commitments vector.
func (p *Pedersen) Verify(abscissa *big.Int, part SecretPart, commitments []*big.Int) error {
	if abscissa == nil {
		return ErrNilAbscissa
	}

	ctx, err := big.NewIntContext()
	if err != nil {
		return err
	}
	defer ctx.Destroy()

	mont, err := big.NewMontgomeryContext()
	if err != nil {
		return err
	}
	defer mont.Destroy()

	if err := mont.Set(p.group.P, ctx); err != nil {
		return err
	}

	vandermondeAbscissa, err := p.vandermondeAbscissa(ctx, abscissa)
	if err != nil {
		return err
	}

	return p.verifyWithContext(mont, ctx, vandermondeAbscissa, part, commitments)
}

// VerifyShares verifies if every secret part is valid.
func (p *Pedersen) VerifyShares(s *Shares) error {
	err := p.validateShares(s)
	if err != nil {
		return err
	}

	partsCount := len(s.Parts[0])
	concLimit := p.adjustConcLimit(partsCount)
	chunksIndex := p.balanceIndices(p.parts, concLimit)
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

			mont, err := big.NewMontgomeryContext()
			if err != nil {
				return err
			}
			defer mont.Destroy()

			if err := mont.Set(p.group.P, ctx); err != nil {
				return err
			}

			for idx := chunk.start; idx < chunk.end; idx++ {
				// compute Vandermonde abscissa
				vandermondeAbscissa, err := p.vandermondeAbscissa(ctx, s.Abscissae[idx])
				if err != nil {
					return err
				}

				for partIndex := 0; partIndex < partsCount; partIndex++ {
					if (SecretPart{}) == s.Parts[idx][partIndex] {
						continue
					}

					err := p.verifyWithContext(
						mont,
						ctx,
						vandermondeAbscissa, s.Parts[idx][partIndex], s.Commitments[partIndex])
					if err != nil {
						return err
					}
				}
			}

			return nil
		})
	}

	return group.Wait()
}
