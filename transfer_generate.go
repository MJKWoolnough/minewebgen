package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/minecraft"
	"github.com/MJKWoolnough/minewebgen/internal/data"
)

func (t Transfer) generate(name string, r *byteio.StickyReader, w *byteio.StickyWriter, f *os.File, size int64) error {
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

	t.c.Generators.mu.RLock()
	gs := make([]data.Generator, len(t.c.Generators.List))
	for n, g := range t.c.Generators.List {
		gs[n] = *g
	}
	t.c.Generators.mu.RUnlock()
	var g data.Generator
	if len(gs) == 0 {
		return errors.New("no generators installed")
	} else if len(gs) == 1 {
		g = gs[0]
	} else {
		w.WriteUint8(1)
		w.WriteInt16(int16(len(gs)))
		for _, tg := range gs {
			data.WriteString(w, tg.Name)
		}
		if w.Err != nil {
			return w.Err
		}
		gID := r.ReadInt16()
		if gID < 0 || int(gID) >= len(gs) {
			return errors.New("unknown generator")
		}
		g = gs[gID]
	}

	ms := DefaultMapSettings()
	ms["level-type"] = minecraft.FlatGenerator
	ms["generator-settings"] = "0"
	ms["motd"] = name

	j, err := os.Open(path.Join(g.Path, "data.gen"))
	if err != nil {
		return err
	}
	var gj data.GeneratorData
	err = json.NewDecoder(j).Decode(&gj)
	j.Close()
	if err != nil {
		return err
	}

	for k, v := range gj.Options {
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

	cmd := exec.Command(t.c.Settings().GeneratorExecutable)
	cmd.ExtraFiles = append(cmd.ExtraFiles, f)
	cmd.Dir, err = os.Getwd()
	if err != nil {
		return err
	}
	cmd.Stdout = w
	pw, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	pww := byteio.StickyWriter{Writer: &byteio.LittleEndianWriter{pw}}
	pww.WriteInt64(t.c.Settings().GeneratorMaxMem)
	pww.WriteInt64(size)
	data.WriteString(&pww, g.Path)
	data.WriteString(&pww, name)
	data.WriteString(&pww, mapPath)

	if pww.Err != nil {
		return pww.Err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	done = true
	mp.Lock()
	mp.Server = -1
	mp.Unlock()

	return nil
}
