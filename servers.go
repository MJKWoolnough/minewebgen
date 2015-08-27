package main

import (
	"archive/zip"
	"errors"
	"os"
	"path"
	"strconv"
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
	w.WriteUint8(1)
	l := r.ReadUint8()
	name := make([]byte, l)
	r.Read(name)
	jars := make([]*zip.File, 0, 16)
	for _, file := range zr.File {
		if strings.HasSuffix(file.Name, ".jar") {
			jars = append(jars, file)
		}
	}
	d, err := setupServerDir()
	if len(jars) == 0 {
		err = os.Rename(f.Name(), path.Join(d, "server.jar"))
	} else {
		if len(jars) > 1 {
			w.WriteUint8(1)
			w.WriteInt16(int16(len(jars)))
			for _, jar := range jars {
				writeString(jar.Name())
			}
			p := r.ReadUint16()
			if int(p) >= len(jars) {
				err = ErrNoServer
			}
			jars[0] = jars[p]
		}
		if err == nil {
			err = unzip(zr, d)
			if err == nil {
				err = os.Rename(path.Join(d, jars[0]), path.Join(d, "server.jar"))
			}
		}
	}
	if err != nil {
		os.RemoveAll(d)
		return err
	}
	config.createServer(string(name), d)
	return nil
}

func setupServerDir() (string, error) {
	num := 0
	for {
		dir := path.Join(config.ServersDir, strconv.Itoa(num))
		err := os.MkdirAll(dir, 0777)
		if err == nil {
			break
		}
		if !os.IsExist(err) {
			return err
		}
		num++
	}
	return nil
}

// Errors
var (
	ErrNoName   = errors.New("no name received")
	ErrNoServer = errors.New("no server found")
)
