package main

import (
	"errors"
	"io"

	"github.com/MJKWoolnough/byteio"
	"honnef.co/go/js/dom"
)

func WrapEvent(f func(...dom.Element), c ...dom.Element) func(dom.Event) {
	return func(dom.Event) {
		go f(c...)
	}
}

func ReadError(r *byteio.StickyReader) error {
	s := ReadString(r)
	if r.Err != nil {
		return r.Err
	}
	return errors.New(s)
}

func WriteString(w *byteio.StickyWriter, s string) {
	w.WriteUint16(uint16(len(s)))
	w.Write([]byte(s))
}

func ReadString(r *byteio.StickyReader) string {
	length := r.ReadUint16()
	str := make([]byte, int(length))
	io.ReadFull(r, str)
	return string(str)
}
