package main

import (
	"archive/zip"
	"errors"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

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
		err = moveFile(f.Name(), path.Join(d, "server.jar"))
	} else {
		if len(jars) > 1 {
			w.WriteUint8(1)
			w.WriteInt16(int16(len(jars)))
			for _, jar := range jars {
				writeString(w, jar.Name)
			}
			p := r.ReadInt16()
			if int(p) >= len(jars) || p < 0 {
				err = ErrNoServer
			} else {
				jars[0] = jars[p]
			}
		}
		if err == nil {
			err = unzip(zr, d)
			if err == nil {
				err = os.Rename(path.Join(d, jars[0].Name), path.Join(d, "server.jar"))
			}
		}
	}
	if err != nil {
		os.RemoveAll(d)
		return err
	}
	serverProperties := DefaultServerSettings()
	f, err := os.Create(path.Join(d, "properties.server"))
	if err != nil {
		os.RemoveAll(d)
		return err
	}
	_, err = serverProperties.WriteTo(f)
	if err != nil {
		os.RemoveAll(d)
		return err
	}
	config.createServer(string(name), d)
	return nil
}

var serverDirLock sync.Mutex

func setupServerDir() (string, error) {
	serverDirLock.Lock()
	defer serverDirLock.Unlock()
	num := 0
	for {
		dir := path.Join(config.ServersDir, strconv.Itoa(num))
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0777)
			if err != nil {
				return "", err
			}
			return dir, nil
		}
		num++
	}
}

// Errors
var (
	ErrNoName   = errors.New("no name received")
	ErrNoServer = errors.New("no server found")
)
