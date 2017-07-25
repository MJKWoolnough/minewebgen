package main

import (
	"archive/zip"
	"io"
	"os"
	"path"
	"syscall"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/minewebgen/internal/data"
)

func writeError(w *byteio.StickyLittleEndianWriter, err error) {
	w.WriteUint8(0)
	data.WriteString(w, err.Error())
}

func moveFile(from, to string) error {
	err := os.Rename(from, to)
	if e, ok := err.(*os.LinkError); !ok || e.Err != syscall.EXDEV {
		return err
	}
	fromf, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromf.Close()
	tof, err := os.Create(to)
	if err != nil {
		return err
	}
	defer tof.Close()
	_, err = io.Copy(tof, fromf)
	return err
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
