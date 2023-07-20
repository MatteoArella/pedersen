// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

//go:build openssldev
// +build openssldev

package big

// #cgo CFLAGS: -DGO_OPENSSL_DEV
// #cgo pkg-config: libcrypto
// #include "goopenssl.h"
import "C"
import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

var (
	initOnce sync.Once
	// errInit is set when first calling Init().
	errInit error
	// vMajor and vMinor hold the major/minor OpenSSL version.
	// It is only populated if Init has been called.
	vMajor, vMinor int
)

func errUnsuportedVersion() error {
	return errors.New("openssl: OpenSSL version: " + strconv.Itoa(vMajor) + "." + strconv.Itoa(vMinor) + " is not supported")
}

// Init loads and initializes OpenSSL.
// It must be called before any other OpenSSL call.
//
// Only the first call to Init is effective,
// subsequent calls will return the same error result as the one from the first call.
func Init() error {
	initOnce.Do(func() {
		vMajor = int(C.go_openssl_version_major())
		vMinor = int(C.go_openssl_version_minor())

		if vMajor == -1 || vMinor == -1 {
			errInit = errors.New("openssl: can't retrieve OpenSSL version")
			return
		}
		var supported bool
		if vMajor == 1 {
			supported = vMinor == 0 || vMinor == 1
		} else if vMajor == 3 {
			// OpenSSL team guarantees API and ABI compatibility within the same major version since OpenSSL 3.
			supported = true
		}
		if !supported {
			errInit = errUnsuportedVersion()
			return
		}

		C.go_openssl_OPENSSL_init()
		if vMajor == 1 && vMinor == 0 {
			if C.go_openssl_thread_setup() != 1 {
				errInit = newOpenSSLError("openssl: thread setup")
				return
			}
			C.go_openssl_ERR_load_crypto_strings()
		} else {
			flags := C.uint64_t(C.GO_OPENSSL_INIT_LOAD_CONFIG | C.GO_OPENSSL_INIT_LOAD_CRYPTO_STRINGS)
			if C.go_openssl_OPENSSL_init_crypto(flags, nil) != 1 {
				errInit = newOpenSSLError("openssl: init crypto")
				return
			}
		}
	})

	return errInit
}

func init() {
	err := Init()
	if err != nil {
		panic(err)
	}
}

// VersionText returns the version text of the OpenSSL currently loaded.
func VersionText() string {
	return C.GoString(C.go_openssl_OpenSSL_version(0))
}

func isLegacy1() bool {
	return vMajor == 1
}

func is11() bool {
	return (vMajor == 1 && vMinor == 1)
}

func isGeq11() bool {
	return is30() || is11()
}

func is30() bool {
	return vMajor == 3
}

func newOpenSSLError(msg string) error {
	var b strings.Builder
	var e C.ulong

	b.WriteString(msg)
	b.WriteString("\nopenssl error(s):\n")

	for {
		e = C.go_openssl_ERR_get_error()
		if e == 0 {
			break
		}
		var buf [256]byte
		C.go_openssl_ERR_error_string_n(e, (*C.char)(unsafe.Pointer(&buf[0])), 256)
		b.Write(buf[:])
		b.WriteByte('\n')
	}
	return errors.New(b.String())
}
