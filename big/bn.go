// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

//go:build cgo
// +build cgo

package big

// #include "goopenssl.h"
import "C"
import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"math"
	"runtime"
	"unsafe"
)

var (
	ErrInvalidParse = errors.New("invalid parse")
)

// An Int represents a signed multi-precision integer.
type Int struct {
	bn *C.GO_BIGNUM
}

func (z *Int) wrapInt(finalize bool) *Int {
	runtime.SetFinalizer(z, func(bn *Int) {
		if bn.bn == nil {
			return
		}

		if finalize {
			C.go_openssl_BN_free(bn.bn)
		}

		bn.bn = nil
	})

	return z
}

func (z *Int) init() error {
	if z.bn == nil {
		if !isGeq11() {
			z.bn = C.go_openssl_BN_new()
			if z.bn == nil {
				return newOpenSSLError("BN_new")
			}
		} else {
			z.bn = C.go_openssl_BN_secure_new()
			if z.bn == nil {
				return newOpenSSLError("BN_secure_new")
			}
		}

		z.wrapInt(true)
	}

	return nil
}

// Allocates and initialize an Int struct.
func NewInt() (*Int, error) {
	i := &Int{}

	err := i.init()
	if err != nil {
		return nil, err
	}

	return i, nil
}

// Returns a constant Int with value 1.
func One() *Int {
	one := &Int{
		bn: C.go_openssl_BN_value_one(),
	}

	return one.wrapInt(false)
}

func (i *Int) SetConstantTime() *Int {
	err := i.init()
	if err != nil {
		return i
	}

	if isGeq11() {
		C.go_openssl_BN_set_flags(i.bn, C.GO_BN_FLG_CONSTTIME)
	} else {
		C.legacy_1_0_BN_set_flags(i.bn, C.GO_BN_FLG_CONSTTIME)
	}

	return i
}

// String returns the decimal representation of i.
func (i *Int) String() string {
	return C.GoString(C.go_openssl_BN_bn2dec(i.bn))
}

// String returns the hexadecimal representation of i.
func (i *Int) Hex() string {
	return C.GoString(C.go_openssl_BN_bn2hex(i.bn))
}

func (i *Int) unmarshalString(data string) error {
	length := len(data)

	if length < 1 {
		return nil
	}

	if length > 2 && data[0] == '0' && (data[1] == 'x' || data[1] == 'X') {
		return i.SetHexString(data[2:])
	}

	// try decimal
	err := i.SetDecString(data)
	if err == nil {
		return nil
	}

	// try exadecimal
	return i.SetHexString(data)
}

// MarshalJSON implements the json.Marshaler interface.
func (i *Int) MarshalJSON() ([]byte, error) {
	return json.Marshal("0x" + i.Hex())
}

// UnmarshalJSON implements the json.Unmarshaler interface..
func (i *Int) UnmarshalJSON(data []byte) error {
	var x string
	err := json.Unmarshal(data, &x)
	if err != nil {
		return err
	}

	return i.unmarshalString(x)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (i *Int) MarshalText() ([]byte, error) {
	return []byte("0x" + i.Hex()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (i *Int) UnmarshalText(data []byte) error {
	return i.unmarshalString(string(data))
}

// GeneratePrime generates a pseudo-random prime number of at least bit length bits
// using the IntContext provided in ctx.
// The returned number is probably prime with a negligible error. The maximum error
// rate is 2^-128. It's 2^-287 for a 512 bit prime, 2^-435 for a 1024 bit prime, 2^-648
// for a 2048 bit prime, and lower than 2^-882 for primes larger than 2048 bit.
// If safe is true, it will be a safe prime (i.e. a prime p so that (p-1)/2 is also prime).
// ctx is a previously allocated IntContext used for temporary variables.
func GeneratePrime(ctx *IntContext, bits int, safe bool) (*Int, error) {
	p, err := NewInt()
	if err != nil {
		return nil, err
	}

	safePrime := C.int(0)
	if safe {
		safePrime = C.int(1)
	}

	if !is30() {
		r := C.go_openssl_BN_generate_prime_ex(p.bn, C.int(bits), safePrime, nil, nil, nil)
		if r != 1 {
			return nil, newOpenSSLError("BN_generate_prime_ex")
		}

		return p, nil
	}

	newCtx := ctx
	if newCtx == nil {
		newCtx, err = NewIntContext()
		if err != nil {
			return nil, err
		}

		defer newCtx.Destroy()
	}

	r := C.go_openssl_BN_generate_prime_ex2(p.bn, C.int(bits), safePrime, nil, nil, nil, newCtx.ctx)
	if r != 1 {
		return nil, newOpenSSLError("BN_generate_prime_ex2")
	}

	return p, nil
}

// ProbablyPrime tests if the number z is prime. The functions tests until one of the tests
// shows that z is composite, or all the tests passed. If z passes all these tests, it is
// considered a probable prime.
// The test performed on z are trial division by a number of small primes and rounds of the
// of the Miller-Rabin probabilistic primality test.
// The functions do at least 64 rounds of the Miller-Rabin test giving a maximum false
// positive rate of 2^-128.
// If the size of z is more than 2048 bits, they do at least 128 rounds giving a maximum
// false positive rate of 2^-256.
// ctx is a previously allocated IntContext used for temporary variables.
func (z *Int) ProbablyPrime(ctx *IntContext) (bool, error) {
	newCtx := ctx
	if newCtx == nil {
		newCtx, err := NewIntContext()
		if err != nil {
			return false, err
		}

		defer newCtx.Destroy()
	}

	if isLegacy1() {
		nChecks := C.go_openssl_BN_prime_checks_for_size(C.int(z.BitLen()))
		ret := C.go_openssl_BN_is_prime_ex(z.bn, nChecks, newCtx.ctx, nil)
		if ret == -1 {
			return false, newOpenSSLError("BN_is_prime_ex")
		}

		return ret == 1, nil
	} else {

		ret := C.go_openssl_BN_check_prime(z.bn, newCtx.ctx, nil)
		if ret == -1 {
			return false, newOpenSSLError("BN_check_prime")
		}

		return ret == 1, nil
	}
}

// Add sets z to the sum x+y and places the result in z.
func (z *Int) Add(x, y *Int) error {
	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_add(z.bn, x.bn, y.bn)
	if ret != 1 {
		return newOpenSSLError("BN_add")
	}

	return nil
}

// Sub sets z to the difference x-y and places the result in z.
func (z *Int) Sub(x, y *Int) error {
	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_sub(z.bn, x.bn, y.bn)
	if ret != 1 {
		return newOpenSSLError("BN_sub")
	}

	return nil
}

// Mul multiplies x and y and places the result in z.
// For multiplication by powers of 2, use [Lsh].
// ctx is a previously allocated IntContext used for temporary variables.
func (z *Int) Mul(ctx *IntContext, x, y *Int) error {
	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_mul(z.bn, x.bn, y.bn, ctx.ctx)
	if ret != 1 {
		return newOpenSSLError("BN_mul")
	}

	return nil
}

// ModMul multiplies x by y and finds the nonnegative remainder respective to modulus m (z=(x*y) mod m).
// For more efficient algorithms for repeated computations using the same modulus, see [ModMulMontgomery].
// ctx is a previously allocated IntContext used for temporary variables.
func (z *Int) ModMul(ctx *IntContext, x, y, m *Int) error {
	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_mod_mul(z.bn, x.bn, y.bn, m.bn, ctx.ctx)
	if ret != 1 {
		return newOpenSSLError("BN_mod_mul")
	}

	return nil
}

// ModMulMontgomery implement Montgomery multiplication.
// It computes Mont(x,y):=x*y*R^-1 and places the result in z.
// ctx is a previously allocated IntContext used for temporary variables.
func (z *Int) ModMulMontgomery(mont *MontgomeryContext, ctx *IntContext, x, y *Int) error {
	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_mod_mul_montgomery(z.bn, x.bn, y.bn, mont.ctx, ctx.ctx)
	if ret != 1 {
		return newOpenSSLError("BN_mod_mul_montgomery")
	}

	return nil
}

// Div divides z by y and places the result in z.
// For division by powers of 2, use [Rsh].
// ctx is a previously allocated IntContext used for temporary variables.
func (z *Int) Div(ctx *IntContext, x, y *Int) error {
	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_div(z.bn, nil, x.bn, y.bn, ctx.ctx)
	if ret != 1 {
		return newOpenSSLError("BN_div")
	}

	return nil
}

// Exp raises x to the y-th power and places the result in z (z=x^y).
// This function is faster than repeated applications of [Mul].
// ctx is a previously allocated IntContext used for temporary variables.
func (z *Int) Exp(ctx *IntContext, x, y *Int) error {
	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_exp(z.bn, x.bn, y.bn, ctx.ctx)
	if ret != 1 {
		return newOpenSSLError("BN_exp")
	}

	return nil
}

// ModExp computes x to the y-th power modulo m (z=x^y % m).
// This function uses less time and space than [Exp].
// Do not call this function when m is even and any of the parameters
// have the constant-time flag set.
// ctx is a previously allocated IntContext used for temporary variables.
func (z *Int) ModExp(ctx *IntContext, x, y, m *Int) error {
	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_mod_exp(z.bn, x.bn, y.bn, m.bn, ctx.ctx)
	if ret != 1 {
		return newOpenSSLError("BN_mod_exp")
	}

	return nil
}

// ModExpMont computes z to the y-th power modulo m (z=x^y % m) using Montgomery multiplication.
// mont is a Montgomery context and can be nil. In the case mont is nil, it will be initialized
// within the function, so you can save time on initialization if you provide it in advance.
// If any of the parameters x, y or m have the constant-time flag set, this function uses fixed
// windows and the special precomputation memory layout to limit data-dependency to a minimum
// to protect secret exponents.
// ctx is a previously allocated IntContext used for temporary variables.
func (z *Int) ModExpMont(mont *MontgomeryContext, ctx *IntContext, x, y, m *Int) error {
	var montCtx *C.GO_BN_MONT_CTX

	if mont != nil {
		montCtx = mont.ctx
	}

	var modulus *C.GO_BIGNUM
	if m != nil {
		modulus = m.bn
	}

	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_mod_exp_mont(z.bn, x.bn, y.bn, modulus, ctx.ctx, montCtx)
	if ret != 1 {
		return newOpenSSLError("BN_mod_exp_mont")
	}

	return nil
}

// Mod sets z to the modulus x%y for y != 0.
// ctx is a previously allocated IntContext used for temporary variables.
func (z *Int) Mod(ctx *IntContext, x, y *Int) error {
	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_mod(z.bn, x.bn, y.bn, ctx.ctx)
	if ret != 1 {
		return newOpenSSLError("BN_mod")
	}

	return nil
}

// ModInverse sets z to the multiplicative inverse of g in the ring ℤ/nℤ.
// ctx is a previously allocated IntContext used for temporary variables.
func (z *Int) ModInverse(ctx *IntContext, g, n *Int) error {
	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_mod_inverse(z.bn, g.bn, n.bn, ctx.ctx)
	if ret == nil {
		return newOpenSSLError("BN_mod_inverse")
	}

	return nil
}

// BitLen returns the length of the absolute value of z in bits. The bit length of 0 is 0.
func (z *Int) BitLen() int {
	if z.bn == nil {
		return 0
	}

	return int(C.go_openssl_BN_num_bits(z.bn))
}

// BytesLen returns the size of z in bytes.
func (z *Int) BytesLen() int {
	if z.bn == nil {
		return 0
	}

	return int(C.go_openssl_BN_num_bytes(z.bn))
}

// SetBytes interprets buf as the bytes of a big-endian unsigned integer, sets z to that value, and returns z.
func (z *Int) SetBytes(buf []byte) *Int {
	err := z.init()
	if err != nil {
		return z
	}

	C.go_openssl_BN_bin2bn((*C.uchar)(unsafe.Pointer(&buf[0])), C.int(len(buf)), z.bn)
	return z
}

// SetDecString sets z to the value of s interpreted in the decimal base.
func (z *Int) SetDecString(s string) error {
	err := z.init()
	if err != nil {
		return err
	}

	str := C.CString(s)
	defer C.free(unsafe.Pointer(str))

	ret := C.go_openssl_BN_dec2bn(&z.bn, str)
	if ret == C.int(0) {
		return newOpenSSLError("BN_dec2bn")
	} else if ret < C.int(len(s)) {
		return ErrInvalidParse
	}

	return nil
}

// SetHexString sets z to the value of s interpreted in the hexadecimal base.
func (z *Int) SetHexString(s string) error {
	err := z.init()
	if err != nil {
		return err
	}

	str := C.CString(s)
	defer C.free(unsafe.Pointer(str))

	ret := C.go_openssl_BN_hex2bn(&z.bn, str)
	if ret == 0 {
		return newOpenSSLError("BN_hex2bn")
	} else if ret < C.int(len(s)) {
		return ErrInvalidParse
	}

	return nil
}

// SetUInt64 sets z to the value of x.
func (z *Int) SetUInt64(x uint64) error {
	err := z.init()
	if err != nil {
		return err
	}

	ok := C.go_openssl_BN_set_word(z.bn, C.GO_BN_ULONG(x))
	if ok != 1 {
		return newOpenSSLError("BN_set_word")
	}

	return nil
}

// Bytes sets buf to the absolute value of z as a big-endian byte slice.
// If the absolute value of z doesn't fit in buf, FillBytes will panic.
func (z *Int) FillBytes(buf []byte) error {
	bytesLen := z.BytesLen()
	bufLen := len(buf)

	if bufLen < bytesLen {
		panic("bn.Int: buffer too small to fit value")
	}

	if isGeq11() {
		ret := C.go_openssl_BN_bn2binpad(z.bn, (*C.uchar)(unsafe.Pointer(&buf[0])), C.int(bufLen))
		if ret != C.int(bufLen) {
			return newOpenSSLError("BN_bn2binpad")
		}
	} else {
		ret := C.go_openssl_BN_bn2bin(z.bn, (*C.uchar)(unsafe.Pointer(&buf[0])))
		if ret != C.int(bufLen) {
			return newOpenSSLError("BN_bn2bin")
		}
	}

	return nil
}

// Bytes returns the absolute value of z as a big-endian byte slice.
// To use a fixed length slice, or a preallocated one, use [FillBytes].
func (z *Int) Bytes() ([]byte, error) {
	bytesLen := z.BytesLen()
	buf := make([]byte, bytesLen)

	if bytesLen <= 0 {
		return buf, nil
	}

	if isGeq11() {
		ret := C.go_openssl_BN_bn2binpad(z.bn, (*C.uchar)(unsafe.Pointer(&buf[0])), C.int(bytesLen))
		if ret != C.int(bytesLen) {
			return nil, newOpenSSLError("BN_bn2binpad")
		}
	} else {
		ret := C.go_openssl_BN_bn2bin(z.bn, (*C.uchar)(unsafe.Pointer(&buf[0])))
		if ret != C.int(bytesLen) {
			return nil, newOpenSSLError("BN_bn2bin")
		}
	}

	return buf, nil
}

// Lsh shifts x left by n bits and places the result in z (z=x*2^n).
// Note that n must be nonnegative.
func (z *Int) Lsh(x *Int, n uint) error {
	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_lshift(z.bn, x.bn, C.int(n))
	if ret != 1 {
		return newOpenSSLError("BN_lshift")
	}

	return nil
}

// Rsh shifts x right by n bits and places the result in z (z=x/2^n).
// Note that n must be nonnegative.
func (z *Int) Rsh(x *Int, n uint) error {
	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_rshift(z.bn, x.bn, C.int(n))
	if ret != 1 {
		return newOpenSSLError("BN_rshift")
	}

	return nil
}

// Or sets z = x | y.
func (z *Int) Or(x, y *Int) error {
	max, min := x, y

	maxLen, minLen := max.BytesLen(), min.BytesLen()

	if maxLen < minLen {
		max, min = y, x
		maxLen, minLen = max.BytesLen(), min.BytesLen()
	}

	minBytes, err := min.Bytes()
	if err != nil {
		return err
	}

	maxBytes, err := max.Bytes()
	if err != nil {
		return err
	}

	zBytes := make([]byte, maxLen)

	offset := maxLen - minLen
	for i, j := offset, 0; i < maxLen; i, j = i+1, j+1 {
		zBytes[i] = minBytes[j] | maxBytes[i]
	}
	copy(zBytes[0:offset], maxBytes[0:offset])

	z.SetBytes(zBytes)

	return nil
}

// And sets z = x & y.
func (z *Int) And(x, y *Int) error {
	max, min := x, y

	maxLen, minLen := max.BytesLen(), min.BytesLen()

	if maxLen < minLen {
		max, min = y, x
		maxLen, minLen = max.BytesLen(), min.BytesLen()
	}

	minBytes, err := min.Bytes()
	if err != nil {
		return err
	}

	maxBytes, err := max.Bytes()
	if err != nil {
		return err
	}

	zBytes := make([]byte, maxLen)

	offset := maxLen - minLen
	for i, j := offset, 0; i < maxLen; i, j = i+1, j+1 {
		zBytes[i] = minBytes[j] & maxBytes[i]
	}

	z.SetBytes(zBytes)

	return nil
}

// Uint64 returns the uint64 representation of z. If z cannot be represented in a uint64, the function
// returns [math.MaxUint64].
func (z *Int) Uint64() uint64 {
	if z.bn == nil {
		return math.MaxUint64
	}

	return uint64(C.go_openssl_BN_get_word(z.bn))
}

// Set sets z to x.
func (z *Int) Set(x *Int) error {
	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_copy(z.bn, x.bn)
	if ret == nil {
		return newOpenSSLError("BN_copy")
	}

	return nil
}

// RandRange generates a cryptographically strong pseudo-random number z in
// the range 0 <= z < max.
func (z *Int) RandRange(max *Int) error {
	err := z.init()
	if err != nil {
		return err
	}

	ret := C.go_openssl_BN_rand_range(z.bn, max.bn)
	if ret != 1 {
		return newOpenSSLError("BN_rand_range")
	}

	return nil
}

// Cmp compares z and x and returns:
//
//	-1 if z <  x
//	 0 if z == x
//	+1 if z >  x
func (z *Int) Cmp(x *Int) int {
	return int(C.go_openssl_BN_cmp(z.bn, x.bn))
}

// ConstantTimeEq compares z and x and returns true if they are equal, false otherwise.
// The time taken is a function of the bytes length of the numbers and is independent
// of the contents.
func (z *Int) ConstantTimeEq(x *Int) (bool, error) {
	zBytes, err := z.Bytes()
	if err != nil {
		return false, err
	}

	xBytes, err := x.Bytes()
	if err != nil {
		return false, err
	}

	return subtle.ConstantTimeCompare(zBytes, xBytes) == 1, nil
}
