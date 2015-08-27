package main

import (
	"errors"
	"io"

	"github.com/MJKWoolnough/byteio"
)

func readError(r byteio.StickyReader) error {
	length := r.ReadUint16()
	if r.Err != nil {
		return r.Err
	}
	errStr := make([]byte, int(length))
	_, err := io.ReadFull(r.Reader, errStr)
	if err != nil {
		return err
	}
	return errors.New(string(errStr))
}

var ErrUnknown = errors.New("Something bad occurred")
