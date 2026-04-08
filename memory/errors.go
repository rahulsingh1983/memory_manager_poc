package memory

import "errors"

var (
	ErrOutOfMemory  = errors.New("out of memory")
	ErrInvalidHandle = errors.New("invalid handle")
	ErrDoubleFree   = errors.New("double free")
	ErrInvalidSize  = errors.New("invalid size")
	ErrOutOfBounds  = errors.New("out of bounds")
)