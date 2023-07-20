// Copyright (c) Pedersen authors.
// 
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

#ifndef OPENSSL_FUNCS_H
#define OPENSSL_FUNCS_H

#include <stdlib.h> // size_t
#include <stdint.h> // uint64_t

#ifdef GO_OPENSSL_DEV
#include <openssl/crypto.h>
#include <openssl/err.h>
#include <openssl/rand.h>
#include <openssl/bn.h>

#if !defined(OPENSSL_VERSION_MAJOR)
#define OPENSSL_VERSION_MAJOR (OPENSSL_VERSION_NUMBER >> 28)
#endif

#if !defined(OPENSSL_VERSION_MINOR)
#define OPENSSL_VERSION_MINOR ((OPENSSL_VERSION_NUMBER >> 20) & 0x000000FFL)
#endif

#if !(OPENSSL_VERSION_MAJOR == 1 && OPENSSL_VERSION_MINOR == 0)
enum
{
	GO_OPENSSL_INIT_LOAD_CRYPTO_STRINGS = OPENSSL_INIT_LOAD_CRYPTO_STRINGS,
	GO_OPENSSL_INIT_LOAD_CONFIG = OPENSSL_INIT_LOAD_CONFIG,
};
#else
// dummy values for legacy library compilation
enum
{
	GO_OPENSSL_INIT_LOAD_CRYPTO_STRINGS = 0,
	GO_OPENSSL_INIT_LOAD_CONFIG = 1,
};
#endif

typedef BN_ULONG GO_BN_ULONG;

enum
{
	/*
	 * avoid leaking exponent information through timing,
	 * BN_mod_exp_mont() will call BN_mod_exp_mont_consttime,
	 * BN_div() will call BN_div_no_branch,
	 * BN_mod_inverse() will call bn_mod_inverse_no_branch.
	 */
	GO_BN_FLG_CONSTTIME = BN_FLG_CONSTTIME,
};

#else /* !GO_OPENSSL_DEV */

enum
{
	GO_OPENSSL_INIT_LOAD_CRYPTO_STRINGS = 0x00000002L,
	GO_OPENSSL_INIT_LOAD_CONFIG = 0x00000040L,
};

#ifdef __UINT64_TYPE__
typedef __UINT64_TYPE__ GO_BN_ULONG;
#else
typedef __UINT32_TYPE__ GO_BN_ULONG;
#endif

enum
{
	/*
	 * avoid leaking exponent information through timing,
	 * BN_mod_exp_mont() will call BN_mod_exp_mont_consttime,
	 * BN_div() will call BN_div_no_branch,
	 * BN_mod_inverse() will call bn_mod_inverse_no_branch.
	 */
	GO_BN_FLG_CONSTTIME = 0x04,
};

#endif /* GO_OPENSSL_DEV */

typedef struct ossl_init_settings_st GO_OPENSSL_INIT_SETTINGS;
typedef struct ossl_lib_ctx_st GO_OSSL_LIB_CTX;
typedef struct bignum_st GO_BIGNUM;
typedef struct bignum_ctx GO_BN_CTX;
typedef struct bn_mont_ctx_st GO_BN_MONT_CTX;
typedef struct bn_gencb_st GO_BN_GENCB;

// List of all functions from the libcrypto that are used in this package.
// Forgetting to add a function here results in build failure with message reporting the function
// that needs to be added.
//
// The purpose of FOR_ALL_OPENSSL_FUNCTIONS is to define all libcrypto functions
// without depending on the openssl headers so it is easier to use this package
// with an openssl version different that the one used at build time.
//
// The following macros may not be defined at this point,
// they are not resolved here but just accumulated in FOR_ALL_OPENSSL_FUNCTIONS.
//
// DEFINEFUNC defines and loads openssl functions that can be directly called from Go as their signatures match
// the OpenSSL API and do not require special logic.
// The process will be aborted if the function can't be loaded.
//
// DEFINEFUNC_LEGACY_1_0 acts like DEFINEFUNC but only aborts the process if the function can't be loaded
// when using 1.0.x. This indicates the function is required when using 1.0.x, but is unused when using later versions.
// It also might not exist in later versions.
//
// DEFINEFUNC_LEGACY_1 acts like DEFINEFUNC but only aborts the process if the function can't be loaded
// when using 1.x. This indicates the function is required when using 1.x, but is unused when using later versions.
// It also might not exist in later versions.
//
// DEFINEFUNC_1_1 acts like DEFINEFUNC but only aborts the process if function can't be loaded
// when using 1.1.0 or higher.
//
// DEFINEFUNC_3_0 acts like DEFINEFUNC but only aborts the process if function can't be loaded
// when using 3.0.0 or higher.
//
// DEFINEFUNC_RENAMED_1_1 acts like DEFINEFUNC but tries to load the function using the new name when using >= 1.1.x
// and the old name when using 1.0.2. In both cases the function will have the new name.
//
// DEFINEFUNC_RENAMED_3_0 acts like DEFINEFUNC but tries to load the function using the new name when using >= 3.x
// and the old name when using 1.x. In both cases the function will have the new name.

#define FOR_ALL_OPENSSL_FUNCTIONS                                                                                                                                                                                            \
	DEFINEFUNC(unsigned long, ERR_get_error, (void), ())                                                                                                                                                                 \
	DEFINEFUNC(void, ERR_error_string_n, (unsigned long e, char *buf, size_t len), (e, buf, len))                                                                                                                        \
	DEFINEFUNC_RENAMED_1_1(const char *, OpenSSL_version, SSLeay_version, (int type), (type))                                                                                                                            \
	DEFINEFUNC(void, OPENSSL_init, (void), ())                                                                                                                                                                           \
	DEFINEFUNC_LEGACY_1_0(void, ERR_load_crypto_strings, (void), (), )                                                                                                                                                   \
	DEFINEFUNC_LEGACY_1_0(int, CRYPTO_num_locks, (void), (), -1)                                                                                                                                                         \
	DEFINEFUNC_LEGACY_1_0(void, CRYPTO_set_id_callback, (unsigned long (*id_function)(void)), (id_function), )                                                                                                           \
	DEFINEFUNC_LEGACY_1_0(void, CRYPTO_set_locking_callback, (void (*locking_function)(int mode, int n, const char *file, int line)), (locking_function), )                                                              \
	DEFINEFUNC_1_1(int, OPENSSL_init_crypto, (uint64_t ops, const GO_OPENSSL_INIT_SETTINGS *settings), (ops, settings), -1)                                                                                              \
	DEFINEFUNC(GO_BIGNUM *, BN_new, (void), ())                                                                                                                                                                          \
	DEFINEFUNC_1_1(GO_BIGNUM *, BN_secure_new, (void), (), NULL)                                                                                                                                                         \
	DEFINEFUNC(void, BN_free, (GO_BIGNUM * arg0), (arg0))                                                                                                                                                                \
	DEFINEFUNC(void, BN_clear_free, (GO_BIGNUM * arg0), (arg0))                                                                                                                                                          \
	DEFINEFUNC(const GO_BIGNUM *, BN_value_one, (void), ())                                                                                                                                                              \
	DEFINEFUNC(char *, BN_bn2dec, (const GO_BIGNUM *arg0), (arg0))                                                                                                                                                       \
	DEFINEFUNC(char *, BN_bn2hex, (const GO_BIGNUM *arg0), (arg0))                                                                                                                                                       \
	DEFINEFUNC(int, BN_generate_prime_ex, (GO_BIGNUM * ret, int bits, int safe, const GO_BIGNUM *add, const GO_BIGNUM *rem, GO_BN_GENCB *cb), (ret, bits, safe, add, rem, cb))                                           \
	DEFINEFUNC_3_0(int, BN_generate_prime_ex2, (GO_BIGNUM * arg0, int arg1, int arg2, const GO_BIGNUM *arg3, const GO_BIGNUM *arg4, GO_BN_GENCB *arg5, GO_BN_CTX *arg6), (arg0, arg1, arg2, arg3, arg4, arg5, arg6), -1) \
	DEFINEFUNC_LEGACY_1(int, BN_is_prime_ex, (const GO_BIGNUM *arg0, int arg1, GO_BN_CTX *arg2, GO_BN_GENCB *arg3), (arg0, arg1, arg2, arg3), -1)                                                                        \
	DEFINEFUNC_3_0(int, BN_check_prime, (const GO_BIGNUM *arg0, GO_BN_CTX *arg1, GO_BN_GENCB *arg2), (arg0, arg1, arg2), -1)                                                                                             \
	DEFINEFUNC(int, BN_add, (GO_BIGNUM * r, const GO_BIGNUM *a, const GO_BIGNUM *b), (r, a, b))                                                                                                                          \
	DEFINEFUNC(int, BN_sub, (GO_BIGNUM * r, const GO_BIGNUM *a, const GO_BIGNUM *b), (r, a, b))                                                                                                                          \
	DEFINEFUNC(int, BN_mul, (GO_BIGNUM * arg0, const GO_BIGNUM *arg1, const GO_BIGNUM *arg2, GO_BN_CTX *arg3), (arg0, arg1, arg2, arg3))                                                                                 \
	DEFINEFUNC(int, BN_mod_mul, (GO_BIGNUM * arg0, const GO_BIGNUM *arg1, const GO_BIGNUM *arg2, const GO_BIGNUM *arg3, GO_BN_CTX *arg4), (arg0, arg1, arg2, arg3, arg4))                                                \
	DEFINEFUNC(int, BN_mod_mul_montgomery, (GO_BIGNUM * arg0, const GO_BIGNUM *arg1, const GO_BIGNUM *arg2, GO_BN_MONT_CTX *arg3, GO_BN_CTX *arg4), (arg0, arg1, arg2, arg3, arg4))                                      \
	DEFINEFUNC(int, BN_div, (GO_BIGNUM * dv, GO_BIGNUM * rem, const GO_BIGNUM *m, const GO_BIGNUM *d, GO_BN_CTX *ctx), (dv, rem, m, d, ctx))                                                                             \
	DEFINEFUNC(int, BN_exp, (GO_BIGNUM * r, GO_BIGNUM * a, GO_BIGNUM * p, GO_BN_CTX * ctx), (r, a, p, ctx))                                                                                                              \
	DEFINEFUNC(int, BN_mod_exp, (GO_BIGNUM * r, const GO_BIGNUM *a, const GO_BIGNUM *p, const GO_BIGNUM *m, GO_BN_CTX *ctx), (r, a, p, m, ctx))                                                                          \
	DEFINEFUNC(int, BN_mod_exp_mont, (GO_BIGNUM * r, const GO_BIGNUM *a, const GO_BIGNUM *p, const GO_BIGNUM *m, GO_BN_CTX *ctx, GO_BN_MONT_CTX *m_ctx), (r, a, p, m, ctx, m_ctx))                                       \
	DEFINEFUNC(GO_BIGNUM *, BN_mod_inverse, (GO_BIGNUM * ret, const GO_BIGNUM *a, const GO_BIGNUM *n, GO_BN_CTX *ctx), (ret, a, n, ctx))                                                                                 \
	DEFINEFUNC(int, BN_num_bits, (const GO_BIGNUM *arg0), (arg0))                                                                                                                                                        \
	DEFINEFUNC(GO_BIGNUM *, BN_bin2bn, (const unsigned char *arg0, int arg1, GO_BIGNUM *arg2), (arg0, arg1, arg2))                                                                                                       \
	DEFINEFUNC(int, BN_dec2bn, (GO_BIGNUM * *arg0, const char *arg1), (arg0, arg1))                                                                                                                                      \
	DEFINEFUNC(int, BN_hex2bn, (GO_BIGNUM * *arg0, const char *arg1), (arg0, arg1))                                                                                                                                      \
	DEFINEFUNC(int, BN_set_word, (GO_BIGNUM * arg0, GO_BN_ULONG arg1), (arg0, arg1))                                                                                                                                     \
	DEFINEFUNC(int, BN_bn2bin, (const GO_BIGNUM *arg0, unsigned char *arg1), (arg0, arg1))                                                                                                                               \
	DEFINEFUNC_1_1(int, BN_bn2binpad, (const GO_BIGNUM *arg0, unsigned char *arg1, int arg2), (arg0, arg1, arg2), -1)                                                                                                    \
	DEFINEFUNC(int, BN_lshift, (GO_BIGNUM * arg0, const GO_BIGNUM *arg1, int arg2), (arg0, arg1, arg2))                                                                                                                  \
	DEFINEFUNC(int, BN_rshift, (GO_BIGNUM * arg0, const GO_BIGNUM *arg1, int arg2), (arg0, arg1, arg2))                                                                                                                  \
	DEFINEFUNC(GO_BN_ULONG, BN_get_word, (const GO_BIGNUM *arg0), (arg0))                                                                                                                                                \
	DEFINEFUNC(GO_BIGNUM *, BN_copy, (GO_BIGNUM * arg0, const GO_BIGNUM *arg1), (arg0, arg1))                                                                                                                            \
	DEFINEFUNC(int, BN_rand_range, (GO_BIGNUM * arg0, const GO_BIGNUM *arg1), (arg0, arg1))                                                                                                                              \
	DEFINEFUNC(int, BN_cmp, (GO_BIGNUM * arg0, const GO_BIGNUM *arg1), (arg0, arg1))                                                                                                                                     \
	DEFINEFUNC(GO_BN_CTX *, BN_CTX_new, (void), ())                                                                                                                                                                      \
	DEFINEFUNC_3_0(GO_BN_CTX *, BN_CTX_new_ex, (GO_OSSL_LIB_CTX * arg0), (arg0), NULL)                                                                                                                                   \
	DEFINEFUNC_1_1(GO_BN_CTX *, BN_CTX_secure_new, (void), (), NULL)                                                                                                                                                     \
	DEFINEFUNC_3_0(GO_BN_CTX *, BN_CTX_secure_new_ex, (GO_OSSL_LIB_CTX * arg0), (arg0), NULL)                                                                                                                            \
	DEFINEFUNC(void, BN_CTX_free, (GO_BN_CTX * arg0), (arg0))                                                                                                                                                            \
	DEFINEFUNC(void, BN_CTX_start, (GO_BN_CTX * arg0), (arg0))                                                                                                                                                           \
	DEFINEFUNC(void, BN_CTX_end, (GO_BN_CTX * arg0), (arg0))                                                                                                                                                             \
	DEFINEFUNC(GO_BIGNUM *, BN_CTX_get, (GO_BN_CTX * arg0), (arg0))                                                                                                                                                      \
	DEFINEFUNC_1_1(void, BN_set_flags, (GO_BIGNUM * arg0, int arg1), (arg0, arg1), )                                                                                                                                     \
	DEFINEFUNC(GO_BN_MONT_CTX *, BN_MONT_CTX_new, (void), ())                                                                                                                                                            \
	DEFINEFUNC(void, BN_MONT_CTX_free, (GO_BN_MONT_CTX * arg0), (arg0))                                                                                                                                                  \
	DEFINEFUNC(int, BN_MONT_CTX_set, (GO_BN_MONT_CTX * arg0, const GO_BIGNUM *arg1, GO_BN_CTX *arg2), (arg0, arg1, arg2))

#endif /* OPENSSL_FUNCS_H */
