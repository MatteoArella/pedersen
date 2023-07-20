// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package pedersen

import "C"
import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/matteoarella/pedersen/big"
)

const (
	minPrimeBitLen = 64
)

var (
	ErrNilPrime         = errors.New("prime cannot be nil")
	ErrInvalidPrimeSize = fmt.Errorf("prime number size must be at least %d bits", minPrimeBitLen)
	ErrInvalidPrime     = errors.New("invalid prime")
	ErrNilGenerator     = errors.New("generator cannot be nil")
	ErrInvalidGenerator = errors.New("invalid generator")
)

// Group represents a cyclic group.
// P and Q are large primes s.t. p=mq+1 where m is an integer.
// G and H are two generators of the unique subgroup of ℤ*q.
type Group struct {
	P *big.Int
	Q *big.Int
	G *big.Int
	H *big.Int
}

func (g *Group) String() string {
	data, _ := json.Marshal(g)
	return string(data)
}

func (g *Group) validatePrimes(ctx *big.IntContext) error {
	if g.P == nil || g.Q == nil {
		return ErrNilPrime
	}

	if g.P.BitLen() < minPrimeBitLen {
		return ErrInvalidPrimeSize
	}

	// check that p and q are primes s.t. p=mq+1
	// where m is an integer
	ok, err := g.P.ProbablyPrime(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidPrime
	}

	ok, err = g.Q.ProbablyPrime(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidPrime
	}

	pMinus, err := ctx.GetInt()
	if err != nil {
		return err
	}

	if err := pMinus.Sub(g.P, big.One()); err != nil {
		return err
	}

	if err := pMinus.Mod(ctx, pMinus, g.Q); err != nil {
		return err
	}

	zero, err := ctx.GetInt()
	if err != nil {
		return err
	}
	if err := zero.SetUInt64(0); err != nil {
		return err
	}

	// not safe prime
	if pMinus.Cmp(zero) != 0 {
		return ErrInvalidPrime
	}

	return nil
}

func (g *Group) validateGenerator(ctx *big.IntContext, generator *big.Int) error {
	if generator == nil {
		return ErrNilGenerator
	}

	exp, err := big.NewInt()
	if err != nil {
		return err
	}

	if err := exp.Sub(g.P, big.One()); err != nil {
		return err
	}

	if err := exp.Div(ctx, exp, g.Q); err != nil {
		return err
	}

	expected, err := big.NewInt()
	if err != nil {
		return err
	}

	if err := expected.ModExp(ctx, generator, exp, g.P); err != nil {
		return err
	}

	// g^((p-1)/q) mod p != 1
	if expected.Cmp(big.One()) == 0 {
		return ErrInvalidGenerator
	}

	// g^q mod p = 1
	if err := expected.ModExp(ctx, generator, g.Q, g.P); err != nil {
		return err
	}

	if expected.Cmp(big.One()) != 0 {
		return ErrInvalidGenerator
	}

	return nil
}

func (g *Group) validate() error {
	ctx, err := big.NewIntContext()
	if err != nil {
		return err
	}
	defer ctx.Destroy()

	if err := g.validatePrimes(ctx); err != nil {
		return err
	}

	if err := g.validateGenerator(ctx, g.G); err != nil {
		return err
	}

	if err := g.validateGenerator(ctx, g.H); err != nil {
		return err
	}

	return nil
}

func getGenerator(ctx *big.IntContext, p, q *big.Int) (*big.Int, error) {
	ctx.Attach()
	defer ctx.Detach()

	mont, err := big.NewMontgomeryContext()
	if err != nil {
		return nil, err
	}
	defer mont.Destroy()

	if err := mont.Set(p, ctx); err != nil {
		return nil, err
	}

	pMinus, err := ctx.GetInt()
	if err != nil {
		return nil, err
	}

	if err := pMinus.Sub(p, big.One()); err != nil {
		return nil, err
	}

	two, err := ctx.GetInt()
	if err != nil {
		return nil, err
	}

	if err := two.SetUInt64(2); err != nil {
		return nil, err
	}

	exp, err := big.NewInt()
	if err != nil {
		return nil, err
	}
	exp.SetConstantTime()

	if err := exp.Set(pMinus); err != nil {
		return nil, err
	}

	if err := exp.Div(ctx, exp, q); err != nil {
		return nil, err
	}

	for {
		g, err := genRandNum(two, pMinus)
		if err != nil {
			return nil, err
		}
		g.SetConstantTime()

		if err := g.ModExpMont(mont, ctx, g, exp, p); err != nil {
			return nil, err
		}

		if g.Cmp(big.One()) != 0 {
			return g, nil
		}
	}
}

// Generate a new Schnorr group of given bits size.
func NewSchnorrGroup(bits int) (*Group, error) {
	if bits < minPrimeBitLen {
		return nil, ErrInvalidPrimeSize
	}

	// Generate a large prime of size 'bits'
	p, err := big.GeneratePrime(nil, bits, true)
	if err != nil {
		return nil, err
	}
	p.SetConstantTime()

	// Calculate the safe prime q=(p-1)/2 of order 'bits'
	q, err := big.NewInt()
	if err != nil {
		return nil, err
	}
	q.SetConstantTime()

	if err := q.Sub(p, big.One()); err != nil {
		return nil, err
	}

	// divide by 2
	if err := q.Rsh(q, 1); err != nil {
		return nil, err
	}

	ctx, err := big.NewIntContext()
	if err != nil {
		return nil, err
	}
	defer ctx.Destroy()

	g, err := getGenerator(ctx, p, q)
	if err != nil {
		return nil, err
	}

	h, err := getGenerator(ctx, p, q)
	if err != nil {
		return nil, err
	}

	return &Group{
		P: p,
		Q: q,
		G: g,
		H: h,
	}, nil
}
