package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"

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
	for _, f := range zr.File {
		name := path.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(name, f.Mode())
			if err != nil {
				return err
			}
			continue
		}
		inf, err := f.Open()
		if err != nil {
			return err
		}
		of, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			inf.Close()
			return err
		}
		_, err = io.Copy(of, inf)
		inf.Close()
		of.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
