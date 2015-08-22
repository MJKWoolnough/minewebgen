package main

import (
	"archive/zip"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/MJKWoolnough/byteio"
)

func setupServer(f *os.File, r *byteio.StickyReader, w *byteio.StickyWriter) error {
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(f, stat.Size())
	if err != nil {
		return err
	}
	l := r.ReadUint8()
	name := make([]byte, l)
	r.Read(name)
	jars := make([]*zip.File, 0, 16)
	for _, file := range zr.File {
		if strings.HasSuffix(file.Name, ".jar") {
			jars = append(jars, file)
		}
	}
	d, err := setupServerDir(string(name))
	if len(jars) == 0 {
		err = os.Rename(f.Name(), path.Join(d, "server.jar"))
		if err != nil {
			os.RemoveAll(d)
			return err
		}
	} else {
		if len(jars) > 1 {
			w.WriteInt8(2)
			w.WriteInt16(int16(len(jars)))
			p := r.ReadUint16()
			if int(p) >= len(jars) {
				return ErrNoServer
			}
			jars[0] = jars[p]
		}
		err = unzip(zr, d)
	}
	w.WriteUint8(1)
	config.createServer(string(name), d)
	return nil
}

func setupServerDir(name string) (string, error) {
	return "", nil
}

// Errors
var (
	ErrNoName   = errors.New("no name received")
	ErrNoServer = errors.New("no server found")
)
