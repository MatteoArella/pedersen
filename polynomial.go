// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package pedersen

import (
	"github.com/matteoarella/pedersen/big"
)

type polynomial struct {
	coefficients []*big.Int
	order        *big.Int
}

func genRandNum(min, max *big.Int) (*big.Int, error) {
	bg, err := big.NewInt()
	if err != nil {
		return nil, err
	}

	if err := bg.Set(max); err != nil {
		return nil, err
	}

	if err := bg.Sub(bg, min); err != nil {
		return nil, err
	}

	if err := bg.RandRange(max); err != nil {
		return nil, err
	}

	if err := bg.Add(bg, min); err != nil {
		return nil, err
	}

	return bg, nil
}

func newPolynomial(intercept *big.Int, degree int, order *big.Int) (polynomial, error) {
	var err error
	p := polynomial{
		coefficients: make([]*big.Int, degree+1),
	}
	p.order, err = big.NewInt()
	if err != nil {
		return polynomial{}, err
	}
	if err := p.order.Set(order); err != nil {
		return polynomial{}, err
	}

	min, err := big.NewInt()
	if err != nil {
		return polynomial{}, err
	}

	if err := min.SetUInt64(0); err != nil {
		return polynomial{}, err
	}

	if intercept == nil {
		p.coefficients[0], _ = genRandNum(min, order)
	} else {
		coefficient, err := big.NewInt()
		if err != nil {
			return polynomial{}, err
		}

		if err := coefficient.Set(intercept); err != nil {
			return polynomial{}, err
		}

		p.coefficients[0] = coefficient
	}

	if err := randInts(p.coefficients[1:], min, order, false); err != nil {
		return p, err
	}

	return p, nil
}

func (p *polynomial) evaluate(ctx *big.IntContext, x *big.Int) (*big.Int, error) {
	zero, err := ctx.GetInt()
	if err != nil {
		return nil, err
	}

	if err := zero.SetUInt64(0); err != nil {
		return nil, err
	}

	out, err := big.NewInt()
	if err != nil {
		return nil, err
	}

	if x.Cmp(zero) == 0 {
		if err := out.Set(p.coefficients[0]); err != nil {
			return nil, err
		}

		return out, nil
	}

	degree := len(p.coefficients) - 1

	if err := out.Set(p.coefficients[degree]); err != nil {
		return nil, err
	}

	for i := degree - 1; i >= 0; i-- {
		if err := out.Mul(ctx, out, x); err != nil {
			return nil, err
		}

		if err := out.Add(out, p.coefficients[i]); err != nil {
			return nil, err
		}
	}

	if err := out.Mod(ctx, out, p.order); err != nil {
		return nil, err
	}

	return out, nil
}

func interpolatePolynomial(ctx *big.IntContext, xSamples, ySamples []*big.Int, x, order *big.Int) (*big.Int, error) {
	limit := len(xSamples)
	result, err := big.NewInt()
	if err != nil {
		return nil, err
	}

	if err := result.SetUInt64(0); err != nil {
		return nil, err
	}

	ctx.Attach()
	defer ctx.Detach()

	for j := 0; j < limit; j++ {
		basis, err := ctx.GetInt()
		if err != nil {
			return nil, err
		}
		if err := basis.SetUInt64(1); err != nil {
			return nil, err
		}

		for k := 0; k < limit; k++ {
			if j == k {
				continue
			}

			num, err := ctx.GetInt()
			if err != nil {
				return nil, err
			}

			if err := num.Set(xSamples[k]); err != nil {
				return nil, err
			}

			if err := num.Sub(num, x); err != nil {
				return nil, err
			}

			denom, err := ctx.GetInt()
			if err != nil {
				return nil, err
			}

			if err := denom.Set(xSamples[k]); err != nil {
				return nil, err
			}

			if err := denom.Sub(denom, xSamples[j]); err != nil {
				return nil, err
			}

			if err := denom.ModInverse(ctx, denom, order); err != nil {
				return nil, err
			}

			if err := num.Mul(ctx, num, denom); err != nil {
				return nil, err
			}

			if err := basis.Mul(ctx, basis, num); err != nil {
				return nil, err
			}

			if err := basis.Mod(ctx, basis, order); err != nil {
				return nil, err
			}
		}

		if err := basis.Mul(ctx, basis, ySamples[j]); err != nil {
			return nil, err
		}

		if err := result.Add(result, basis); err != nil {
			return nil, err
		}
	}

	if err := result.Mod(ctx, result, order); err != nil {
		return nil, err
	}

	return result, nil
}

func randInts(slice []*big.Int, min, max *big.Int, distinct bool) error {
	n := len(slice)

	checkMap := map[string]bool{}

	for i := 0; i < n; i++ {
		val, err := genRandNum(min, max)
		if err != nil {
			return err
		}

		if distinct {
			samp := val.String()

			for {
				if exists := checkMap[samp]; !exists {
					break
				}

				val, err = genRandNum(min, max)
				if err != nil {
					return err
				}

				samp = val.String()
			}

			checkMap[samp] = true
		}

		slice[i] = val
	}

	return nil
}
