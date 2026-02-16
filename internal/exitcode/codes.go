package exitcode

import (
	"errors"
	"fmt"
	"net"

	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

const (
	Success    = 0
	Validation = 10
	Config     = 11
	Auth       = 12
	Network    = 13
	API        = 14
	Internal   = 15
)

type codedError struct {
	code int
	msg  string
	err  error
}

func (e codedError) Error() string {
	if e.err == nil {
		return e.msg
	}
	if e.msg == "" {
		return e.err.Error()
	}
	return fmt.Sprintf("%s: %v", e.msg, e.err)
}

func (e codedError) Unwrap() error {
	return e.err
}

func (e codedError) ExitCode() int {
	return e.code
}

func New(code int, msg string) error {
	return codedError{code: code, msg: msg}
}

func Wrap(code int, msg string, err error) error {
	if err == nil {
		return nil
	}
	return codedError{code: code, msg: msg, err: err}
}

func Code(err error) int {
	if err == nil {
		return Success
	}

	var c interface {
		ExitCode() int
	}
	if errors.As(err, &c) {
		return c.ExitCode()
	}

	var kErr kiteconnect.Error
	if errors.As(err, &kErr) {
		switch kErr.ErrorType {
		case kiteconnect.TokenError, kiteconnect.PermissionError, kiteconnect.TwoFAError:
			return Auth
		case kiteconnect.NetworkError, kiteconnect.DataError:
			return Network
		case kiteconnect.InputError:
			return Validation
		case kiteconnect.OrderError, kiteconnect.UserError:
			return API
		default:
			return API
		}
	}

	var nErr net.Error
	if errors.As(err, &nErr) {
		return Network
	}

	return Internal
}
