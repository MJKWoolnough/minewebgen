package main

import (
	"archive/zip"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/MJKWoolnough/byteio"
)

func (t Transfer) server(r *byteio.StickyReader, w *byteio.StickyWriter, f *os.File) error {
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(f, stat.Size())
	if err != nil {
		return err
	}
	name := readString(r)
	jars := make([]*zip.File, 0, 16)
	for _, file := range zr.File {
		if strings.HasSuffix(file.Name, ".jar") {
			jars = append(jars, file)
		}
	}
	s := t.c.NewServer()
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
			p := r.ReadInt16()
			if int(p) >= len(jars) || p < 0 {
				err = errors.New("error selecting server jar")
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
		t.c.RemoveServer(s.ID)
		return err
	}
	serverProperties := DefaultServerSettings()
	ps, err := os.OpenFile(path.Join(d, "properties.server"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		t.c.RemoveServer(s.ID)
		return err
	}
	defer ps.Close()
	err = serverProperties.WriteTo(ps)
	if err != nil {
		t.c.RemoveServer(s.ID)
		return err
	}
	go t.c.Save()
	return nil
}
