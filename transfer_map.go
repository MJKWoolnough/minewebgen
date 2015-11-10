package main

import (
	"archive/zip"
	"os"
	"path"

	"github.com/MJKWoolnough/byteio"
)

func (t Transfer) maps(r *byteio.StickyReader, w *byteio.StickyWriter, f *os.File) error {
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(f, stat.Size())
	if err != nil {
		return err
	}
	name := readString(r)
	if r.Err != nil {
		return r.Err
	}
	m := t.c.NewMap()
	done := false
	go func() {
		if !done {
			t.c.RemoveMap(m.ID)
		}
		go t.c.Save()
	}()
	m.Lock()
	m.Name = name
	d := m.Path
	m.Unlock()
	err = unzip(zr, d)
	if err != nil {
		return err
	}
	mapProperties := DefaultMapSettings()
	pm, err := os.OpenFile(path.Join(d, "properties.map"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	defer pm.Close()
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	} else {
		err = mapProperties.WriteTo(pm)
		if err != nil {
			return err
		}
	}
	done = true
	return nil
}
