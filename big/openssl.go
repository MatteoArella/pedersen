// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

//go:build !openssldev
// +build !openssldev

package big

// #include "goopenssl.h"
// #include <dlfcn.h>
// #cgo LDFLAGS: -ldl
import "C"
import (
	"errors"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
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

// knownVersions is a list of supported and well-known libcrypto.so suffixes in decreasing version order.
//
// FreeBSD library version numbering does not directly align to the version of OpenSSL.
// Its preferred search order is 11 -> 111.
//
// Some distributions use 1.0.0 and others (such as Debian) 1.0.2 to refer to the same OpenSSL 1.0.2 version.
//
// Fedora derived distros use different naming for the version 1.0.x.
var knownVersions = [...]string{"3", "1.1", "11", "111", "1.0.2", "1.0.0", "10"}

func errUnsuportedVersion() error {
	return errors.New("openssl: OpenSSL version: " + strconv.Itoa(vMajor) + "." + strconv.Itoa(vMinor) + " is not supported")
}

// Init loads and initializes OpenSSL.
// It must be called before any other OpenSSL call.
//
// Only the first call to Init is effective,
// subsequent calls will return the same error result as the one from the first call.
//
// If GO_OPENSSL_VERSION_OVERRIDE environment variable is not empty, its value will be appended to the OpenSSL shared library name
// as a version suffix when calling dlopen. For example, "GO_OPENSSL_VERSION_OVERRIDE=1.1.1k-fips"
// makes Init look for the shared library libcrypto.so.1.1.1k-fips.
// If GO_OPENSSL_VERSION_OVERRIDE environment variable is empty, Init will try to load the OpenSSL shared library
// using a list if supported and well-known version suffixes, going from higher to lower versions.
func Init() error {
	initOnce.Do(func() {
		version, _ := syscall.Getenv("GO_OPENSSL_VERSION_OVERRIDE")
		handle, err := loadLibrary(version)
		if err != nil {
			errInit = err
			return
		}

		vMajor = int(C.go_openssl_version_major(handle))
		vMinor = int(C.go_openssl_version_minor(handle))
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

		C.go_openssl_load_functions(handle, C.int(vMajor), C.int(vMinor))
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

func init() { //nolint:gochecknoinits
	err := Init()
	if err != nil {
		panic(err)
	}
}

func dlopen(version string) unsafe.Pointer {
	ext := ".so"
	if runtime.GOOS == "darwin" {
		ext = ".dylib"
	}

	cv := C.CString("libcrypto" + ext + "." + version)
	defer C.free(unsafe.Pointer(cv))
	return C.dlopen(cv, C.RTLD_LAZY|C.RTLD_LOCAL)
}

func loadLibrary(version string) (unsafe.Pointer, error) {
	if version != "" {
		// If version is specified try to load it or error out.
		handle := dlopen(version)
		if handle == nil {
			errstr := C.GoString(C.dlerror())
			return nil, errors.New("openssl: can't load libcrypto.so." + version + ": " + errstr)
		}
		return handle, nil
	}
	for _, v := range knownVersions {
		handle := dlopen(v)
		if handle != nil {
			return handle, nil
		}
	}
	return nil, errors.New("openssl: can't load libcrypto.so using any known version suffix")
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
