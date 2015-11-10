package main

import (
	"archive/zip"
	"os"
	"path"

	"github.com/MJKWoolnough/byteio"
)

func (t Transfer) maps(name string, r *byteio.StickyReader, w *byteio.StickyWriter, f *os.File, size int64) error {
	zr, err := zip.NewReader(f, size)
	if err != nil {
		return err
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
