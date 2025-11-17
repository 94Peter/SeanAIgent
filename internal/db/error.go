package db

import "errors"

var (
	ErrNotFound   = errors.New("data not found")
	ErrReadFailed = errors.New("read failed")
)
