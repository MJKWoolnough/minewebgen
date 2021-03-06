package data

import (
	"io"

	"vimagination.zapto.org/byteio"
)

func WriteString(w *byteio.StickyLittleEndianWriter, s string) {
	w.WriteUint16(uint16(len(s)))
	w.Write([]byte(s))
}

func ReadString(r *byteio.StickyLittleEndianReader) string {
	length := r.ReadUint16()
	str := make([]byte, int(length))
	io.ReadFull(r, str)
	return string(str)
}
