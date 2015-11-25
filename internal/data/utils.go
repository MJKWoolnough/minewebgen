package data

import (
	"io"

	"github.com/MJKWoolnough/byteio"
)

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
