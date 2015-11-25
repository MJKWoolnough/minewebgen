package main

import (
	"errors"
	"image/color"
	"os"
	"path"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/minecraft"
	"github.com/MJKWoolnough/minewebgen/internal/data"
	"github.com/MJKWoolnough/ora"
)

func (t Transfer) generate(name string, r *byteio.StickyReader, w *byteio.StickyWriter, f *os.File, size int64) error {
	o, err := ora.Open(f, size)
	if err != nil {
		return err
	}
	mp := t.c.NewMap()
	if mp == nil {
		return errors.New("failed to create map")
	}

	done := false
	defer func() {
		if !done {
			t.c.RemoveMap(mp.ID)
		}
		go t.c.Save()
	}()

	mp.Lock()
	mp.Name = name
	mapPath := mp.Path
	mp.Server = -2
	mp.Unlock()

	b := o.Bounds()
	w.WriteUint8(2)
	w.WriteInt32(int32(b.Max.X) >> 4)
	w.WriteInt32(int32(b.Max.Y) >> 4)

	gNames := t.c.Generators.Names()
	var gID int16
	if len(gNames) == 0 {
		return errors.New("no generators installed")
	} else if len(gNames) == 1 {
		gID = 0
	} else {
		w.WriteUint8(1)
		w.WriteInt16(int16(len(gNames)))
		for _, gName := range gNames {
			data.WriteString(w, gName)
		}
		if w.Err != nil {
			return w.Err
		}
		gID = r.ReadInt16()
		if gID < 0 || int(gID) >= len(gNames) {
			return errors.New("unknown generator")
		}
	}

	g := t.c.Generators.Get(gNames[gID])
	if g == nil {
		return errors.New("generator removed")
	}

	c := make(chan paint, 1024)
	m := make(chan string, 4)
	e := make(chan struct{}, 0)
	defer close(e)
	go func() {
		defer close(c)
		defer close(m)
		for {
			select {
			case message := <-m:
				w.WriteUint8(3)
				data.WriteString(w, message)
			case p := <-c:
				w.WriteUint8(4)
				w.WriteInt32(p.X)
				w.WriteInt32(p.Y)
				r, g, b, a := p.RGBA()
				w.WriteUint8(uint8(r >> 8))
				w.WriteUint8(uint8(g >> 8))
				w.WriteUint8(uint8(b >> 8))
				w.WriteUint8(uint8(a >> 8))
			case <-e:
				return
			}
		}
	}()

	if err = g.Generate(name, mapPath, o, c, m); err != nil {
		return err
	}

	ms := DefaultMapSettings()
	ms["level-type"] = minecraft.FlatGenerator
	ms["generator-settings"] = "0"
	ms["motd"] = name
	for k, v := range g.generator.Options {
		ms[k] = v
	}

	pf, err := os.Create(path.Join(mapPath, "properties.map"))
	if err != nil {
		return err
	}

	if err = ms.WriteTo(pf); err != nil {
		return err
	}
	pf.Close()

	done = true
	mp.Lock()
	mp.Server = -1
	mp.Unlock()

	return nil
}

type paint struct {
	color.Color
	X, Y int32
}
