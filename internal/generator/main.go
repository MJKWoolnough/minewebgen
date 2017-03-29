package main

import (
	"encoding/json"
	"image/color"
	"os"
	"path"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/minecraft"
	"github.com/MJKWoolnough/minewebgen/internal/data"
	"github.com/MJKWoolnough/ora"
)

func LoadGenerator(f *os.File) (*generator, error) {
	g := new(generator)
	err := json.NewDecoder(f).Decode(&g.generator)
	if err != nil {
		return nil, err
	}
	if len(g.generator.Terrain) == 0 {
		g.generator.Terrain = []data.ColourBlocks{{Name: "Empty"}}
	}
	if len(g.generator.Biomes) == 0 {
		g.generator.Biomes = []data.ColourBiome{{Name: "Plains", Biome: 1}}
	}
	if len(g.generator.Plants) == 0 {
		g.generator.Plants = []data.ColourBlocks{{Name: "Empty"}}
	}

	g.Terrain.Blocks = make([]data.Blocks, len(g.generator.Terrain)+1)
	g.Terrain.Palette = make(color.Palette, len(g.generator.Terrain))
	for i := range g.generator.Terrain {
		g.Terrain.Blocks[i] = g.generator.Terrain[i].Blocks
		g.Terrain.Palette[i] = g.generator.Terrain[i].Colour
	}
	g.Terrain.Blocks[len(g.Terrain.Blocks)-1].Base.ID = 9

	g.Biomes.Values = make([]minecraft.Biome, len(g.generator.Biomes))
	g.Biomes.Palette = make(color.Palette, len(g.generator.Biomes))
	for i := range g.generator.Biomes {
		g.Biomes.Values[i] = g.generator.Biomes[i].Biome
		g.Biomes.Palette[i] = g.generator.Biomes[i].Colour
	}

	g.Plants.Blocks = make([]data.Blocks, len(g.generator.Plants))
	g.Plants.Palette = make(color.Palette, len(g.generator.Plants))
	for i := range g.generator.Plants {
		g.Plants.Blocks[i] = g.generator.Plants[i].Blocks
		g.Plants.Palette[i] = g.generator.Plants[i].Colour
	}
	return g, nil
}

func main() {
	r := byteio.StickyReader{Reader: &byteio.LittleEndianReader{Reader: os.Stdin}}
	w := byteio.StickyWriter{Writer: &byteio.LittleEndianWriter{Writer: os.Stdout}}
	if err := generate(&r, &w, os.NewFile(3, "data.ora")); err != nil {
		w.WriteUint8(0)
		data.WriteString(&w, err.Error())
		os.Exit(1)
	}
}

func generate(r *byteio.StickyReader, w *byteio.StickyWriter, of *os.File) error {
	memoryLimit := r.ReadUint64()
	size := r.ReadInt64()
	gPath := data.ReadString(r)
	levelName := data.ReadString(r)
	mapPath := data.ReadString(r)
	if r.Err != nil {
		return r.Err
	}
	o, err := ora.Open(of, size)
	if err != nil {
		return err
	}
	f, err := os.Open(path.Join(gPath, "data.gen"))
	if err != nil {
		return err
	}
	g, err := LoadGenerator(f)
	if e := f.Close(); e != nil {
		return e
	}
	if err != nil {
		return err
	}

	b := o.Bounds()
	w.WriteUint8(2)
	w.WriteInt32(int32(b.Max.X) >> 4)
	w.WriteInt32(int32(b.Max.Y) >> 4)

	c := make(chan paint, 1024)
	m := make(chan string, 4)
	e := make(chan struct{}, 0)
	go func() {
		defer close(e)
		defer close(m)
		defer close(c)
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

	err = g.Generate(levelName, mapPath, o, c, m, memoryLimit)

	e <- struct{}{}
	<-e
	return err
}

type paint struct {
	color.Color
	X, Y int32
}
