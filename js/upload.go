package main

import (
	"io"

	"github.com/MJKWoolnough/byteio"
)

type readLener interface {
	io.Reader
	Len() int
}

func uploadFile(id uint8, file readLener, w byteio.StickyWriter) {
	w.WriteUint8(id)
	w.WriteInt64(int64(file.Len()))
	if w.Err == nil {
		_, err := io.Copy(w.Writer, file)
		w.Err = err
	}
}
