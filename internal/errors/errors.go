// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package errors

import (
	"errors"
	"fmt"
)

// Error represents an error that could be wrapping another error.
type Error struct {
	orig error
	msg  string
}

// WrapErrorf returns a wrapped error.
func WrapErrorf(orig error, format string, a ...interface{}) error {
	return &Error{
		orig: orig,
		msg:  fmt.Sprintf(format, a...),
	}
}

// NewErrorf instantiates a new error.
func NewErrorf(format string, a ...interface{}) error {
	return WrapErrorf(nil, format, a...)
}

// Error returns the message, when wrapping errors the wrapped error is returned.
func (e *Error) Error() string {
	if e.orig != nil {
		return fmt.Sprintf("%s: %v", e.msg, e.orig)
	}

	return e.msg
}

// Unwrap returns the wrapped error, if any.
func (e *Error) Unwrap() error {
	return e.orig
}

func UnwrapAll(err error) error {
	for {
		newErr := errors.Unwrap(err)
		if newErr == nil {
			return err
		}

		err = newErr
	}
}
