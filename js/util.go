package main

import (
	"errors"
	"io"

	"github.com/MJKWoolnough/byteio"
)

type readLener interface {
	io.Reader
	Len() int
}

func uploadFile(id uint8, file readLener, w *byteio.StickyWriter) {
	w.WriteUint8(id)
	w.WriteInt64(int64(file.Len()))
	if w.Err == nil {
		_, err := io.Copy(w.Writer, file)
		w.Err = err
	}
}

func readError(r *byteio.StickyReader) error {
	length := r.ReadUint16()
	if r.Err != nil {
		return r.Err
	}
	errStr := make([]byte, int(length))
	_, err := io.ReadFull(r, errStr)
	if err != nil {
		return err
	}
	return errors.New(string(errStr))
}

func readString(r *byteio.StickyReader) string {
	length := r.ReadUint16()
	str := make([]byte, int(length))
	io.ReadFull(r, str)
	return string(str)
}

var ErrUnknown = errors.New("Something bad occurred")
