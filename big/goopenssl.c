// Copyright (c) Pedersen authors.
// 
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

// go:build !openssldev
//  +build !openssldev

#include "goopenssl.h"

#include <dlfcn.h>
#include <stdio.h>

static unsigned long
version_num(void *handle)
{
	unsigned long (*fn)(void);
	// OPENSSL_version_num is defined in OpenSSL 1.1.0 and 1.1.1.
	fn = (unsigned long (*)(void))dlsym(handle, "OpenSSL_version_num");
	if (fn != NULL)
		return fn();

	// SSLeay is defined in OpenSSL 1.0.2.
	fn = (unsigned long (*)(void))dlsym(handle, "SSLeay");
	if (fn != NULL)
		return fn();

	return 0;
}

int go_openssl_version_major(void *handle)
{
	unsigned int (*fn)(void);
	// OPENSSL_version_major is supported since OpenSSL 3.
	fn = (unsigned int (*)(void))dlsym(handle, "OPENSSL_version_major");
	if (fn != NULL)
		return (int)fn();

	// If OPENSSL_version_major is not defined, try with OpenSSL 1 functions.
	unsigned long num = version_num(handle);
	if (num < 0x10000000L || num >= 0x20000000L)
		return -1;

	return 1;
}

int go_openssl_version_minor(void *handle)
{
	unsigned int (*fn)(void);
	// OPENSSL_version_major is supported since OpenSSL 3.
	fn = (unsigned int (*)(void))dlsym(handle, "OPENSSL_version_minor");
	if (fn != NULL)
		return (int)fn();

	// If OPENSSL_version_major is not defined, try with OpenSSL 1 functions.
	unsigned long num = version_num(handle);
	// OpenSSL version number follows this schema:
	// MNNFFPPS: major minor fix patch status.
	if (num < 0x10000000L || num >= 0x10200000L)
	{
		// We only support minor version 0 and 1,
		// so there is no need to implement an algorithm
		// that decodes the version number into individual components.
		return -1;
	}

	if (num >= 0x10100000L)
		return 1;

	return 0;
}

// Approach taken from .Net System.Security.Cryptography.Native
// https://github.com/dotnet/runtime/blob/f64246ce08fb7a58221b2b7c8e68f69c02522b0d/src/libraries/Native/Unix/System.Security.Cryptography.Native/opensslshim.c

#define DEFINEFUNC(ret, func, args, argscall) ret(*_g_##func) args;
#define DEFINEFUNC_LEGACY_1_0(ret, func, args, argscall, res) DEFINEFUNC(ret, func, args, argscall)
#define DEFINEFUNC_LEGACY_1(ret, func, args, argscall, res) DEFINEFUNC(ret, func, args, argscall)
#define DEFINEFUNC_1_1(ret, func, args, argscall, res) DEFINEFUNC(ret, func, args, argscall)
#define DEFINEFUNC_3_0(ret, func, args, argscall, res) DEFINEFUNC(ret, func, args, argscall)
#define DEFINEFUNC_RENAMED_1_1(ret, func, oldfunc, args, argscall) DEFINEFUNC(ret, func, args, argscall)
#define DEFINEFUNC_RENAMED_3_0(ret, func, oldfunc, args, argscall) DEFINEFUNC(ret, func, args, argscall)

FOR_ALL_OPENSSL_FUNCTIONS

#undef DEFINEFUNC
#undef DEFINEFUNC_LEGACY_1_0
#undef DEFINEFUNC_LEGACY_1
#undef DEFINEFUNC_1_1
#undef DEFINEFUNC_3_0
#undef DEFINEFUNC_RENAMED_1_1
#undef DEFINEFUNC_RENAMED_3_0

// Load all the functions stored in FOR_ALL_OPENSSL_FUNCTIONS
// and assign them to their corresponding function pointer
// defined in goopenssl.h.
void go_openssl_load_functions(void *handle, int major, int minor)
{
#define DEFINEFUNC_INTERNAL(name, func)                                                                               \
	_g_##name = dlsym(handle, func);                                                                              \
	if (_g_##name == NULL)                                                                                        \
	{                                                                                                             \
		fprintf(stderr, "Cannot get required symbol " #func " from libcrypto version %d.%d\n", major, minor); \
		abort();                                                                                              \
	}
#define DEFINEFUNC(ret, func, args, argscall) \
	DEFINEFUNC_INTERNAL(func, #func)
#define DEFINEFUNC_LEGACY_1_0(ret, func, args, argscall, res) \
	if (major == 1 && minor == 0)                         \
	{                                                     \
		DEFINEFUNC_INTERNAL(func, #func)              \
	}
#define DEFINEFUNC_LEGACY_1(ret, func, args, argscall, res) \
	if (major == 1)                                     \
	{                                                   \
		DEFINEFUNC_INTERNAL(func, #func)            \
	}
#define DEFINEFUNC_1_1(ret, func, args, argscall, res) \
	if (major == 3 || (major == 1 && minor == 1))  \
	{                                              \
		DEFINEFUNC_INTERNAL(func, #func)       \
	}
#define DEFINEFUNC_3_0(ret, func, args, argscall, res) \
	if (major == 3)                                \
	{                                              \
		DEFINEFUNC_INTERNAL(func, #func)       \
	}
#define DEFINEFUNC_RENAMED_1_1(ret, func, oldfunc, args, argscall) \
	if (major == 1 && minor == 0)                              \
	{                                                          \
		DEFINEFUNC_INTERNAL(func, #oldfunc)                \
	}                                                          \
	else                                                       \
	{                                                          \
		DEFINEFUNC_INTERNAL(func, #func)                   \
	}
#define DEFINEFUNC_RENAMED_3_0(ret, func, oldfunc, args, argscall) \
	if (major == 1)                                            \
	{                                                          \
		DEFINEFUNC_INTERNAL(func, #oldfunc)                \
	}                                                          \
	else                                                       \
	{                                                          \
		DEFINEFUNC_INTERNAL(func, #func)                   \
	}

	FOR_ALL_OPENSSL_FUNCTIONS

#undef DEFINEFUNC
#undef DEFINEFUNC_LEGACY_1_0
#undef DEFINEFUNC_LEGACY_1
#undef DEFINEFUNC_1_1
#undef DEFINEFUNC_3_0
#undef DEFINEFUNC_RENAMED_1_1
#undef DEFINEFUNC_RENAMED_3_0
}

// LEGACY
struct bignum_st
{
	GO_BN_ULONG *d; /* Pointer to an array of 'BN_BITS2' bit
			 * chunks. */
	int top;	/* Index of last used d +1. */
	/* The next are internal book keeping for bn_expand. */
	int dmax; /* Size of the d array. */
	int neg;  /* one if the number is negative */
	int flags;
};

#define LEGACY_1_0_BN_set_flags(b, n) ((b)->flags |= (n))
#define LEGACY_1_0_BN_get_flags(b, n) ((b)->flags & (n))

void legacy_1_0_BN_set_flags(GO_BIGNUM *arg0, int arg1)
{
	LEGACY_1_0_BN_set_flags(arg0, arg1);
}

int legacy_1_0_BN_get_flags(const GO_BIGNUM *arg0, int arg1)
{
	return LEGACY_1_0_BN_get_flags(arg0, arg1);
}

int go_openssl_BN_num_bytes(const GO_BIGNUM *a)
{
	return (go_openssl_BN_num_bits(a) + 7) / 8;
}

int go_openssl_BN_mod(GO_BIGNUM *rem, const GO_BIGNUM *a, const GO_BIGNUM *m, GO_BN_CTX *ctx)
{
	return go_openssl_BN_div(NULL, rem, a, m, ctx);
}

/*
 * BN_prime_checks_for_size() returns the number of Miller-Rabin iterations
 * that will be done for checking that a random number is probably prime. The
 * error rate for accepting a composite number as prime depends on the size of
 * the prime |b|. The error rates used are for calculating an RSA key with 2 primes,
 * and so the level is what you would expect for a key of double the size of the
 * prime.
 *
 * This table is generated using the algorithm of FIPS PUB 186-4
 * Digital Signature Standard (DSS), section F.1, page 117.
 * (https://dx.doi.org/10.6028/NIST.FIPS.186-4)
 *
 * The following magma script was used to generate the output:
 * securitybits:=125;
 * k:=1024;
 * for t:=1 to 65 do
 *   for M:=3 to Floor(2*Sqrt(k-1)-1) do
 *     S:=0;
 *     // Sum over m
 *     for m:=3 to M do
 *       s:=0;
 *       // Sum over j
 *       for j:=2 to m do
 *         s+:=(RealField(32)!2)^-(j+(k-1)/j);
 *       end for;
 *       S+:=2^(m-(m-1)*t)*s;
 *     end for;
 *     A:=2^(k-2-M*t);
 *     B:=8*(Pi(RealField(32))^2-6)/3*2^(k-2)*S;
 *     pkt:=2.00743*Log(2)*k*2^-k*(A+B);
 *     seclevel:=Floor(-Log(2,pkt));
 *     if seclevel ge securitybits then
 *       printf "k: %5o, security: %o bits  (t: %o, M: %o)\n",k,seclevel,t,M;
 *       break;
 *     end if;
 *   end for;
 *   if seclevel ge securitybits then break; end if;
 * end for;
 *
 * It can be run online at:
 * http://magma.maths.usyd.edu.au/calc
 *
 * And will output:
 * k:  1024, security: 129 bits  (t: 6, M: 23)
 *
 * k is the number of bits of the prime, securitybits is the level we want to
 * reach.
 *
 * prime length | RSA key size | # MR tests | security level
 * -------------+--------------|------------+---------------
 *  (b) >= 6394 |     >= 12788 |          3 |        256 bit
 *  (b) >= 3747 |     >=  7494 |          3 |        192 bit
 *  (b) >= 1345 |     >=  2690 |          4 |        128 bit
 *  (b) >= 1080 |     >=  2160 |          5 |        128 bit
 *  (b) >=  852 |     >=  1704 |          5 |        112 bit
 *  (b) >=  476 |     >=   952 |          5 |         80 bit
 *  (b) >=  400 |     >=   800 |          6 |         80 bit
 *  (b) >=  347 |     >=   694 |          7 |         80 bit
 *  (b) >=  308 |     >=   616 |          8 |         80 bit
 *  (b) >=   55 |     >=   110 |         27 |         64 bit
 *  (b) >=    6 |     >=    12 |         34 |         64 bit
 */
#define BN_prime_checks_for_size(b) ((b) >= 3747 ? 3 : (b) >= 1345 ? 4  \
						   : (b) >= 476	   ? 5  \
						   : (b) >= 400	   ? 6  \
						   : (b) >= 347	   ? 7  \
						   : (b) >= 308	   ? 8  \
						   : (b) >= 55	   ? 27 \
								   : /* b >= 6 */ 34)

int go_openssl_BN_prime_checks_for_size(int size)
{
	return BN_prime_checks_for_size(size);
}
