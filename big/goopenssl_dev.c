// Copyright (c) Pedersen authors.
// 
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

// go:build openssldev
//  +build openssldev

#include "goopenssl.h"

#include <stdio.h>

int go_openssl_version_major(void)
{
	return OPENSSL_VERSION_MAJOR;
}

int go_openssl_version_minor(void)
{
	return OPENSSL_VERSION_MINOR;
}

#define DEFINEFUNC(ret, func, args, argscall) \
	inline ret go_openssl_##func args     \
	{                                     \
		return func argscall;         \
	}

#define DEFINEFUNC_RENAMED(ret, func, renamed, args, argscall) \
	inline ret go_openssl_##func args                      \
	{                                                      \
		return renamed argscall;                       \
	}

#define DEFINEFUNC_DUMMY(ret, func, args, argscall, res)                                  \
	inline ret go_openssl_##func args                                                 \
	{                                                                                 \
		fprintf(stderr, "Cannot get required symbol " #func " from libcrypto\n"); \
		abort();                                                                  \
		return res;                                                               \
	}

#if OPENSSL_VERSION_MAJOR == 1 && OPENSSL_VERSION_MINOR == 0
#define DEFINEFUNC_LEGACY_1_0(ret, func, args, argscall, res) \
	DEFINEFUNC(ret, func, args, argscall)
#else
#define DEFINEFUNC_LEGACY_1_0(ret, func, args, argscall, res) \
	DEFINEFUNC_DUMMY(ret, func, args, argscall, res)
#endif

#if OPENSSL_VERSION_MAJOR == 1
#define DEFINEFUNC_LEGACY_1(ret, func, args, argscall, res) \
	DEFINEFUNC(ret, func, args, argscall)
#else
#define DEFINEFUNC_LEGACY_1(ret, func, args, argscall, res) \
	DEFINEFUNC_DUMMY(ret, func, args, argscall, res)
#endif

#if (OPENSSL_VERSION_MAJOR == 3 || (OPENSSL_VERSION_MAJOR == 1 && OPENSSL_VERSION_MINOR == 1))
#define DEFINEFUNC_1_1(ret, func, args, argscall, res) \
	DEFINEFUNC(ret, func, args, argscall)
#else
#define DEFINEFUNC_1_1(ret, func, args, argscall, res) \
	DEFINEFUNC_DUMMY(ret, func, args, argscall, res)
#endif

#if OPENSSL_VERSION_MAJOR == 3
#define DEFINEFUNC_3_0(ret, func, args, argscall, res) \
	DEFINEFUNC(ret, func, args, argscall)
#else
#define DEFINEFUNC_3_0(ret, func, args, argscall, res) \
	DEFINEFUNC_DUMMY(ret, func, args, argscall, res)
#endif

#if OPENSSL_VERSION_MAJOR == 1 && OPENSSL_VERSION_MINOR == 0
#define DEFINEFUNC_RENAMED_1_1(ret, func, oldfunc, args, argscall) \
	DEFINEFUNC_RENAMED(ret, func, oldfunc, args, argscall)
#else
#define DEFINEFUNC_RENAMED_1_1(ret, func, oldfunc, args, argscall) \
	DEFINEFUNC(ret, func, args, argscall)
#endif

#if OPENSSL_VERSION_MAJOR == 1
#define DEFINEFUNC_RENAMED_3_0(ret, func, oldfunc, args, argscall) \
	DEFINEFUNC_RENAMED(ret, func, oldfunc, args, argscall)
#else
#define DEFINEFUNC_RENAMED_3_0(ret, func, oldfunc, args, argscall) \
	DEFINEFUNC(ret, func, args, argscall)
#endif

FOR_ALL_OPENSSL_FUNCTIONS

#undef DEFINEFUNC
#undef DEFINEFUNC_RENAMED
#undef DEFINEFUNC_DUMMY
#undef DEFINEFUNC_LEGACY_1_0
#undef DEFINEFUNC_LEGACY_1
#undef DEFINEFUNC_1_1
#undef DEFINEFUNC_3_0
#undef DEFINEFUNC_RENAMED_1_1
#undef DEFINEFUNC_RENAMED_3_0

// LEGACY
void legacy_1_0_BN_set_flags(GO_BIGNUM *arg0, int arg1)
{
	BN_set_flags(arg0, arg1);
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
