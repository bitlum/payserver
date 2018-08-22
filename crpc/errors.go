package crpc

import (
	"fmt"
)

const (
	ErrAssetNotSupported   = iota + 1
	ErrNetworkNotSupported
	ErrInvalidArgument

	// ErrInternal...
	ErrInternal
)

type Error struct {
	code     int
	errMsg   string
	internal bool
}

func (e Error) Error() string {
	return e.errMsg
}

func newErrNetworkNotSupported(network, operation string) Error {
	return Error{
		code: ErrNetworkNotSupported,
		errMsg: fmt.Sprintf("%v: operation \"%v\" isn't supported for network %v",
			ErrNetworkNotSupported, operation, network),
	}
}

func newErrAssetNotSupported(asset, media string) Error {
	return Error{
		code: ErrAssetNotSupported,
		errMsg: fmt.Sprintf("%v: asset(%v) is not supported for media(%v)",
			ErrAssetNotSupported, asset, media),
	}
}

func newErrInternal(desc string) Error {
	return Error{
		code:   ErrInternal,
		errMsg: fmt.Sprintf("%v: internal error: %v", ErrInternal, desc),
	}
}

func newErrInvalidArgument(argName string) Error {
	return Error{
		code: ErrInvalidArgument,
		errMsg: fmt.Sprintf("%v: invalid argument '%v'", ErrInvalidArgument,
			argName),
	}
}
