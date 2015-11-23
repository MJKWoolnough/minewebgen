package main

import (
	"io"
	"os"
	"path"
	"strings"

	"github.com/MJKWoolnough/byteio"
)

func (t Transfer) generator(name string, _ *byteio.StickyReader, w *byteio.StickyWriter, f *os.File, _ int64) error {
	name = strings.Replace(name, "/", "", -1)
	err := t.c.Generators.LoadGenerator(name, f)
	if err != nil {
		return err
	}
	f.Seek(0, 0)
	g, err := os.Create(path.Join(t.c.Settings().DirGenerators, name+".gen"))
	if err != nil {
		return err
	}
	defer g.Close()
	_, err = io.Copy(g, f)
	return err
}
