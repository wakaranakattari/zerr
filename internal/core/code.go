package core

import (
	"errors"

	"github.com/wakaranakattari/zerr/internal/multierror"
)

// Code is a machine-readable classification for errors. It implements
// error so codes can be used as sentinel targets with errors.Is:
//
//	var ErrNotFound = zerr.Code("not_found")
//
//	err := zerr.WithCode(ErrFailure, zerr.Code("load_failed"), "load")
//	errors.Is(err, ErrNotFound) // false
//
// A Code is comparable and treats only exact string equality as a
// match. Codes are rendered in %+v output and delivered to structured
// logs through LogValue.
type Code string

// Error implements the error interface so a Code can be a sentinel.
func (c Code) Error() string { return string(c) }

// Is reports whether err is classified with the given code anywhere in
// its chain. It is a convenience over errors.Is(err, code).
func Is(err error, code Code) bool { return errors.Is(err, code) }

// CodeOf returns the first code found while walking err's chain from
// the outermost node inward, or "" when the chain carries no code.
func CodeOf(err error) Code {
	switch e := err.(type) {
	case *Error:
		if e.code != "" {
			return e.code
		}
		if e.cause != nil {
			return CodeOf(e.cause)
		}
	case *multierror.Error:
		for _, c := range e.Unwrap() {
			if code := CodeOf(c); code != "" {
				return code
			}
		}
	}
	return ""
}
