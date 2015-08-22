package main

import (
	"archive/zip"
	"fmt"

	"github.com/MJKWoolnough/byteio"
)

func writeError(w *byteio.StickyWriter, err error) {
	w.WriteUint8(0)
	errStr := []byte(err.Error())
	w.WriteUint16(uint16(len(errStr)))
	w.Write(errStr)
	fmt.Println("error:", err)
}

func unzip(zr *zip.Reader, dest string) error {
	return nil
}
