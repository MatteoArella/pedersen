// Copyright (c) Pedersen authors.
// 
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

#ifndef GOOPENSSL_H
#define GOOPENSSL_H

// This header file describes the OpenSSL ABI as built for use in Go.

#include "openssl_funcs.h"

int go_openssl_thread_setup(void);

// Define pointers to all the used OpenSSL functions.
// Calling C function pointers from Go is currently not supported.
// It is possible to circumvent this by using a C function wrapper.
// https://pkg.go.dev/cmd/cgo
#ifndef GO_OPENSSL_DEV
int go_openssl_version_major(void *handle);
int go_openssl_version_minor(void *handle);
void go_openssl_load_functions(void *handle, int major, int minor);

#define DEFINEFUNC(ret, func, args, argscall)    \
	extern ret(*_g_##func) args;             \
	static inline ret go_openssl_##func args \
	{                                        \
		return _g_##func argscall;       \
	}
#else /* GO_OPENSSL_DEV */
int go_openssl_version_major(void);
int go_openssl_version_minor(void);

#define DEFINEFUNC(ret, func, args, argscall) \
	extern ret go_openssl_##func args;
#endif /* GO_OPENSSL_DEV */

#define DEFINEFUNC_LEGACY_1_0(ret, func, args, argscall, res) \
	DEFINEFUNC(ret, func, args, argscall)
#define DEFINEFUNC_LEGACY_1(ret, func, args, argscall, res) \
	DEFINEFUNC(ret, func, args, argscall)
#define DEFINEFUNC_1_1(ret, func, args, argscall, res) \
	DEFINEFUNC(ret, func, args, argscall)
#define DEFINEFUNC_3_0(ret, func, args, argscall, res) \
	DEFINEFUNC(ret, func, args, argscall)
#define DEFINEFUNC_RENAMED_1_1(ret, func, oldfunc, args, argscall) \
	DEFINEFUNC(ret, func, args, argscall)
#define DEFINEFUNC_RENAMED_3_0(ret, func, oldfunc, args, argscall) \
	DEFINEFUNC(ret, func, args, argscall)

FOR_ALL_OPENSSL_FUNCTIONS

#undef DEFINEFUNC
#undef DEFINEFUNC_LEGACY_1_0
#undef DEFINEFUNC_LEGACY_1
#undef DEFINEFUNC_1_1
#undef DEFINEFUNC_3_0
#undef DEFINEFUNC_RENAMED_1_1
#undef DEFINEFUNC_RENAMED_3_0

void legacy_1_0_BN_set_flags(GO_BIGNUM *arg0, int arg1);

int go_openssl_BN_num_bytes(const GO_BIGNUM *a);
int go_openssl_BN_mod(GO_BIGNUM *rem, const GO_BIGNUM *a, const GO_BIGNUM *m, GO_BN_CTX *ctx);
int go_openssl_BN_prime_checks_for_size(int size);

#endif /* GOOPENSSL_H */
