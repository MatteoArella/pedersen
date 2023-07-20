// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package pedersen

import (
	"errors"
	"math"

	"github.com/matteoarella/pedersen/big"
	"golang.org/x/sync/errgroup"
)

var (
	ErrEmptySecret           = errors.New("cannot split an empty secret")
	ErrInsufficientAbscissae = errors.New("abscissae cannot be less than parts")
)

const (
	zerosInfoSizeBytes = 4
)

type splitValue struct {
	index       int
	secretParts []SecretPart
	secretComm  []*big.Int
}

func leadingZeros(buff []byte) uint64 {
	sum := uint64(0)

	for _, n := range buff {
		if n == 0 {
			sum++
		} else {
			break
		}
	}

	return sum
}

func bigIntPadding(ctx *big.IntContext, buff []byte) (*big.Int, error) {
	ctx.Attach()
	defer ctx.Detach()

	zeros, err := ctx.GetInt()
	if err != nil {
		return nil, err
	}
	zerosCount := leadingZeros(buff)

	if err := zeros.SetUInt64(zerosCount); err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(buff)

	// append leading zeros info to n
	if err := n.Lsh(n, zerosInfoSizeBytes*8); err != nil {
		return nil, err
	}

	if zerosCount > 0 {
		if err := n.Or(n, zeros); err != nil {
			return nil, err
		}
	}

	return n, nil
}

func splitSecret(ctx *big.IntContext, value []byte, max *big.Int) ([]*big.Int, error) {
	valueLen := len(value)

	// ensure every chunk value is smaller than max
	partLen := int(math.Floor(float64(max.BitLen())/float64(8))) - zerosInfoSizeBytes
	if partLen <= 0 {
		partLen = 1
	}

	partCount := valueLen / partLen
	if partCount*partLen < valueLen {
		partCount++
	}

	splitted := make([]*big.Int, partCount)
	var err error

	for i := 0; i < partCount; i++ {
		end := (i + 1) * partLen
		if end > valueLen {
			end = valueLen
		}

		valuePart := value[i*partLen : end]
		splitted[i], err = bigIntPadding(ctx, valuePart)
		if err != nil {
			return nil, err
		}
	}

	return splitted, nil
}

type chunkRange struct {
	start, end int
}

func (p *Pedersen) split(
	mont *big.MontgomeryContext,
	ctx *big.IntContext,
	index int,
	secret *big.Int,
	abscissae []*big.Int,
) (splitValue, error) {
	ctx.Attach()
	defer ctx.Detach()

	F, err := newPolynomial(secret, p.threshold-1, p.group.Q)
	if err != nil {
		return splitValue{}, err
	}

	K, err := newPolynomial(nil, p.threshold-1, p.group.Q)
	if err != nil {
		return splitValue{}, err
	}

	secretParts := make([]SecretPart, p.parts)
	for i := 0; i < p.parts; i++ {
		s, err := F.evaluate(ctx, abscissae[i])
		if err != nil {
			return splitValue{}, err
		}

		t, err := K.evaluate(ctx, abscissae[i])
		if err != nil {
			return splitValue{}, err
		}

		secretParts[i] = SecretPart{
			SShare: s,
			TShare: t,
		}
	}

	commitments := make([]*big.Int, p.threshold)

	for i := 0; i < p.threshold; i++ {
		commitment, err := p.commit(mont, ctx, F.coefficients[i], K.coefficients[i])
		if err != nil {
			return splitValue{}, err
		}

		commitments[i] = commitment
	}

	return splitValue{
		index:       index,
		secretParts: secretParts,
		secretComm:  commitments,
	}, nil
}

// Split takes a secret and generates a `parts`
// number of shares, `threshold` of which are required to reconstruct
// the secret.
// If the secret that has to be split is not representable in the cyclic group,
// the secret is split into chunks, and each chunk is split into secret parts according
// to Pedersen verifiable secret sharing.
// The abscissae are used to evaluate the polynomials at the given points.
// If abscissae is nil, random abscissae are generated.
func (p *Pedersen) Split(secret []byte, abscissae []*big.Int) (*Shares, error) {
	if len(secret) == 0 {
		return nil, ErrEmptySecret
	}

	if abscissae == nil {
		abscissae = make([]*big.Int, p.parts)

		err := randInts(abscissae, big.One(), p.group.Q, true)
		if err != nil {
			return nil, err
		}
	} else if len(abscissae) < p.parts {
		return nil, ErrInsufficientAbscissae
	}

	ctx, err := big.NewIntContext()
	if err != nil {
		return nil, err
	}
	defer ctx.Destroy()

	// split secret into many byte slices and process them
	splitted, err := splitSecret(ctx, secret, p.group.Q)
	if err != nil {
		return nil, err
	}

	splittedLen := len(splitted)
	concLimit := p.adjustConcLimit(splittedLen)
	chunksIndex := p.balanceIndices(splittedLen, concLimit)
	parts := make([][]SecretPart, p.parts)
	commitments := make([][]*big.Int, splittedLen)

	for shareIdx := 0; shareIdx < p.parts; shareIdx++ {
		parts[shareIdx] = make([]SecretPart, splittedLen)
	}

	group := errgroup.Group{}
	group.SetLimit(concLimit)

	for _, chunk := range chunksIndex {
		secrets := splitted[chunk.start:chunk.end]
		chunk := chunk

		group.Go(func() error {
			/* create new IntContext for each goroutine since it's not safe to
			use the same context concurrently */
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

			for idx, secret := range secrets {
				if secret.Cmp(p.group.Q) > 0 {
					return ErrInvalidPrimeSize
				}

				result, err := p.split(mont, ctx, chunk.start+idx, secret, abscissae)
				if err != nil {
					return err
				}

				for shareIdx := 0; shareIdx < p.parts; shareIdx++ {
					parts[shareIdx][result.index] = result.secretParts[shareIdx]
				}

				commitments[result.index] = result.secretComm
			}

			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		return nil, err
	}

	return &Shares{
		Abscissae:   abscissae,
		Parts:       parts,
		Commitments: commitments,
	}, nil
}
