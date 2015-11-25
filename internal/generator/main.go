package main

import (
	"encoding/json"
	"image/color"
	"os"

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

func errCheck(err error) {
	if err != nil {

	}
}

func main() {
	r := byteio.StickyReader{Reader: &byteio.LittleEndianReader{os.Stdin}}
	w := byteio.StickyWriter{Writer: &byteio.LittleEndianWriter{os.Stdout}}
	if err := generate(&r, &w, os.NewFile(3, "data.gen")); err != nil {
		w.WriteUint8(0)
		data.WriteString(&w, err.Error())
	}
}

func generate(r *byteio.StickyReader, w *byteio.StickyWriter, of *os.File) error {
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
	f, err := os.Open(gPath)
	if err != nil {
		return err
	}
	g, err := LoadGenerator(f)
	if e := (f.Close()); e != nil {
		return e
	}
	if err != nil {
		return err
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

	if err = g.Generate(levelName, mapPath, o, c, m); err != nil {
		return err
	}
	return nil
}

type paint struct {
	color.Color
	X, Y int32
}
