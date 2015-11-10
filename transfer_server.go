package main

import (
	"archive/zip"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/MJKWoolnough/byteio"
)

func (t Transfer) server(name string, r *byteio.StickyReader, w *byteio.StickyWriter, f *os.File, size int64) error {
	zr, err := zip.NewReader(f, size)
	if err != nil {
		return err
	}
	jars := make([]*zip.File, 0, 16)
	for _, file := range zr.File {
		if strings.HasSuffix(file.Name, ".jar") {
			jars = append(jars, file)
		}
	}
	s := t.c.NewServer()
	done := false
	defer func() {
		if !done {
			t.c.RemoveServer(s.ID)
		}
		go t.c.Save()
	}()
	if s == nil {
		return errors.New("error creating server")
	}
	s.Lock()
	s.Name = name
	d := s.Path
	s.Unlock()
	if len(jars) == 0 {
		err = moveFile(f.Name(), path.Join(d, "server.jar"))
	} else {
		if len(jars) > 1 {
			w.WriteUint8(1)
			w.WriteInt16(int16(len(jars)))
			for _, jar := range jars {
				writeString(w, jar.Name)
			}
			if w.Err != nil {
				return w.Err
			}
			p := r.ReadInt16()
			if r.Err != nil {
				return r.Err
			}
			if int(p) >= len(jars) || p < 0 {
				return errors.New("error selecting server jar")
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
		return err
	}
	serverProperties := DefaultServerSettings()
	ps, err := os.OpenFile(path.Join(d, "properties.server"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	defer ps.Close()
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	} else {
		err = serverProperties.WriteTo(ps)
		if err != nil {
			return err
		}
	}
	done = true
	return nil
}
