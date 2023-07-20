// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package big

// #include "goopenssl.h"
import "C"
import (
	"runtime"
)

// A IntContext is a structure that holds [Int] temporary variables
// used by library functions.
// Since dynamic memory allocation to create [Int]s is rather expensive
// when used in conjunction with repeated subroutine calls,
// the IntContext structure is used.
// A given IntContext must only be used by a single thread of execution.
// No locking is performed, and the internal pool allocator will not properly
// handle multiple threads of execution.
type IntContext struct {
	ctx *C.GO_BN_CTX
}

type MontgomeryContext struct {
	ctx *C.GO_BN_MONT_CTX
}

func finalizeIntContext(bnCtx *IntContext) {
	if bnCtx.ctx == nil {
		return
	}

	C.go_openssl_BN_CTX_free(bnCtx.ctx)

	bnCtx.ctx = nil
}

func wrapIntContext(ctx *C.GO_BN_CTX) *IntContext {
	c := &IntContext{
		ctx: ctx,
	}

	runtime.SetFinalizer(c, finalizeIntContext)

	return c
}

// NewIntContext allocates and initializes a IntContext structure
func NewIntContext() (*IntContext, error) {
	var ctx *C.GO_BN_CTX

	switch {
	case is30():
		ctx = C.go_openssl_BN_CTX_secure_new_ex(nil)
		if ctx == nil {
			return nil, newOpenSSLError("BN_CTX_secure_new_ex")
		}
	case is11():
		ctx = C.go_openssl_BN_CTX_secure_new()
		if ctx == nil {
			return nil, newOpenSSLError("BN_CTX_secure_new")
		}
	default:
		ctx = C.go_openssl_BN_CTX_new()
		if ctx == nil {
			return nil, newOpenSSLError("BN_CTX_new")
		}
	}

	return wrapIntContext(ctx), nil
}

func (c *IntContext) Attach() {
	C.go_openssl_BN_CTX_start(c.ctx)
}

func (c *IntContext) Detach() {
	C.go_openssl_BN_CTX_end(c.ctx)
}

func (c *IntContext) GetInt() (*Int, error) {
	bn := C.go_openssl_BN_CTX_get(c.ctx)
	if bn == nil {
		return nil, newOpenSSLError("BN_CTX_get")
	}

	i := &Int{
		bn: bn,
	}

	return i.wrapInt(false), nil
}

// Destroy frees the components of the IntContext and the structure itself.
func (c *IntContext) Destroy() {
	finalizeIntContext(c)
}

func finalizeMontgomeryContext(montCtx *MontgomeryContext) {
	if montCtx.ctx == nil {
		return
	}

	C.go_openssl_BN_MONT_CTX_free(montCtx.ctx)

	montCtx.ctx = nil
}

func wrapMontgomeryContext(ctx *C.GO_BN_MONT_CTX) *MontgomeryContext {
	c := &MontgomeryContext{
		ctx: ctx,
	}

	runtime.SetFinalizer(c, finalizeMontgomeryContext)

	return c
}

func NewMontgomeryContext() (*MontgomeryContext, error) {
	ctx := C.go_openssl_BN_MONT_CTX_new()
	if ctx == nil {
		return nil, newOpenSSLError("BN_MONT_CTX_new")
	}

	return wrapMontgomeryContext(ctx), nil
}

func (c *MontgomeryContext) Set(m *Int, ctx *IntContext) error {
	ret := C.go_openssl_BN_MONT_CTX_set(c.ctx, m.bn, ctx.ctx)
	if ret != 1 {
		return newOpenSSLError("BN_MONT_CTX_set")
	}

	return nil
}

func (c *MontgomeryContext) Destroy() {
	finalizeMontgomeryContext(c)
}
