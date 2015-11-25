package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/minewebgen/internal/data"
)

func (t Transfer) generator(name string, r *byteio.StickyReader, w *byteio.StickyWriter, f *os.File, size int64) error {
	g := t.c.Generators.New(t.c.Settings().DirGenerators)
	if g == nil {
		return errors.New("error creating generator")
	}

	done := false
	defer func() {
		if !done {
			t.c.RemoveGenerator(g.ID)
		}
		go t.c.Save()
	}()

	zr, err := zip.NewReader(f, size)
	if err != nil {
		f.Seek(0, 0)
		e := json.NewDecoder(f).Decode(new(data.GeneratorData))
		if e != nil {
			return err
		}
		err = moveFile(f.Name(), path.Join(g.Path, "data.gen"))
		if err != nil {
			return err
		}

		done = true
		return nil
	}

	gens := make([]*zip.File, 0, 16)
	for _, file := range zr.File {
		if file.Name == "data.gen" {
			gens = []*zip.File{file}
			break
		}
		if strings.HasSuffix(file.Name, ".gen") || strings.HasSuffix(file.Name, ".json") {
			gens = append(gens, file)
		}
	}

	if len(gens) == 0 {
		return errors.New("cannot find generator data in zip")
	}
	if len(gens) > 1 {
		w.WriteUint8(1)
		w.WriteInt16(int16(len(gens)))
		for _, gen := range gens {
			data.WriteString(w, gen.Name)
		}
		if w.Err != nil {
			return w.Err
		}
		p := r.ReadInt16()
		if r.Err != nil {
			return r.Err
		}
		if int(p) >= len(gens) || p < 0 {
			return errors.New("error selecting generator data")
		}
		gens[0] = gens[p]
	}

	gd, err := gens[0].Open()
	if err != nil {
		return err
	}

	err = json.NewDecoder(gd).Decode(new(data.GeneratorData))
	if err != nil {
		return err
	}

	err = unzip(zr, g.Path)
	if err != nil {
		return err
	}

	err = os.Rename(path.Join(g.Path, gens[0].Name), path.Join(g.Path, "data.gen"))
	if err != nil {
		return err
	}

	done = true
	return nil
}
