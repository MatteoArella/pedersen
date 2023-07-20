// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package big_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/matteoarella/pedersen/big"
)

func TestMain(m *testing.M) {
	if err := big.Init(); err != nil {
		// An error here could mean that this Linux distro does not have a supported OpenSSL version
		// or that there is a bug in the Init code.
		panic(err)
	}

	fmt.Println("OpenSSL version:", big.VersionText())
	os.Exit(m.Run())
}
