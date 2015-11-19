package main

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/minecraft"
	"github.com/MJKWoolnough/minecraft/nbt"
	"github.com/MJKWoolnough/ora"
)

func toGray(o *ora.ORA, name string) (*image.Gray, error) {
	var p *image.Gray
	if l := o.Layer("water"); l != nil {
		p = image.NewGray(o.Bounds())
		i, err := l.Image()
		if err != nil {
			return nil, err
		}
		draw.Draw(p, image.Rect(0, 0, p.Bounds().Max.X, p.Bounds().Max.Y), i, image.Point{}, draw.Src)
	}
	return p, nil
}

func toPaletted(o *ora.ORA, name string, palette color.Palette) (*image.Paletted, error) {
	var p *image.Paletted
	if l := o.Layer("biomes"); l != nil {
		p = image.NewPaletted(o.Bounds(), palette)
		i, err := l.Image()
		if err != nil {
			return nil, err
		}
		draw.Draw(p, image.Rect(0, 0, p.Bounds().Max.X, p.Bounds().Max.Y), i, image.Point{}, draw.Src)
	}
	return p
}

func (t Transfer) generate(name string, _ *byteio.StickyReader, w *byteio.StickyWriter, f *os.File, size int64) error {
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

	ms := DefaultMapSettings()
	ms["level-type"] = minecraft.FlatGenerator
	ms["generator-settings"] = "0"
	ms["motd"] = name

	pf, err := os.Create(path.Join(mapPath, "properties.map"))
	if err != nil {
		return err
	}

	if err = ms.WriteTo(pf); err != nil {
		return err
	}
	pf.Close()

	b := o.Bounds()
	w.WriteUint8(2)
	w.WriteInt32(int32(b.Max.X) >> 4)
	w.WriteInt32(int32(b.Max.Y) >> 4)
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
				writeString(w, message)
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

	sTerrain := toPaletted(o, "terrain", terrainColours)
	if sTerrain == nil {
		return layerError{"terrain"}
	}

	sHeight := toGray(o, "height")
	if sHeight == nil {
		return layerError{"height"}
	}

	sBiomes := toPaletted(o, "biomes", biomePalette)
	sWater := toGray(o, "water")

	p, err := minecraft.NewFilePath(mapPath)
	if err != nil {
		return err
	}

	level, err := minecraft.NewLevel(p)
	if err != nil {
		return err
	}

	level.LevelName(name)

	m <- "Building Terrain"
	if err := buildTerrain(p, level, sTerrain, sBiomes, sHeight, sWater, c); err != nil {
		return err
	}

	level.LevelName(name)
	level.MobSpawning(false)
	level.KeepInventory(true)
	level.FireTick(false)
	level.DayLightCycle(false)
	level.MobGriefing(false)
	level.Spawn(10, 250, 10)
	level.Generator(minecraft.FlatGenerator)
	level.GeneratorOptions("0")
	level.GameMode(minecraft.Creative)
	level.AllowCommands(true)
	level.MapFeatures(false)

	m <- "Exporting"
	level.Save()
	level.Close()
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

type layerError struct {
	name string
}

func (l layerError) Error() string {
	return "missing layer: " + l.name
}

type terrain struct {
	Base, Top minecraft.Block
	TopLevel  uint8
}

var (
	terrainColours = color.Palette{
		color.RGBA{},
		color.RGBA{255, 255, 0, 255},   // Yellow - Sand
		color.RGBA{0, 255, 0, 255},     // Green - Grass
		color.RGBA{87, 59, 12, 255},    // Brown - Dirt
		color.RGBA{255, 128, 0, 255},   // Orange - Farm
		color.RGBA{128, 128, 128, 255}, // Grey - Stone
		color.RGBA{255, 255, 255, 255}, // White - Snow
	}
	terrainBlocks = []terrain{
		{},
		{minecraft.Block{ID: 24, Data: 2}, minecraft.Block{ID: 12}, 5}, // Sandstone - Sand
		{minecraft.Block{ID: 3}, minecraft.Block{ID: 2}, 1},            // Dirt - Grass
		{minecraft.Block{ID: 3}, minecraft.Block{ID: 3}, 0},            // Dirt - Dirt
		{minecraft.Block{ID: 3}, minecraft.Block{ID: 60, Data: 7}, 1},  // Dirt - Farmland
		{minecraft.Block{ID: 1}, minecraft.Block{ID: 1}, 0},            // Stone - Stone
		{minecraft.Block{ID: 1}, minecraft.Block{ID: 80}, 3},           // Stone - Snow
		{minecraft.Block{ID: 9}, minecraft.Block{ID: 9}, 0},
	}
	biomePalette = color.Palette{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{60, 60, 255, 255},
		color.RGBA{20, 100, 20, 255},
		color.RGBA{255, 255, 0, 255},
		color.RGBA{0, 255, 255, 255},
	}
	biomeList = []minecraft.Biome{
		minecraft.Plains,
		minecraft.Ocean,
		minecraft.Forest,
		minecraft.Desert,
		minecraft.River,
	}
)

func modeTerrain(p *image.Paletted) uint8 {
	b := p.Bounds()
	modeMap := make([]uint16, len(terrainColours))
	most := uint16(0)
	mode := uint8(0)
	for i := b.Min.X; i < b.Max.X; i++ {
		for j := b.Min.Y; j < b.Max.Y; j++ {
			pos := p.ColorIndexAt(i, j)
			modeMap[pos]++
			if m := modeMap[pos]; m > most {
				most = m
				mode = pos
			}
		}
	}
	return mode
}

func meanHeight(g *image.Gray) uint8 {
	b := g.Bounds()
	var total int64
	for i := b.Min.X; i < b.Max.X; i++ {
		for j := b.Min.Y; j < b.Max.Y; j++ {
			total += int64(g.GrayAt(i, j).Y)
		}
	}
	return uint8(total / int64((b.Dx() * b.Dy())))
}

type chunkCache struct {
	mem   *minecraft.MemPath
	level *minecraft.Level
	clear nbt.Tag
	cache map[uint16]nbt.Tag
}

func newCache() *chunkCache {
	mem := minecraft.NewMemPath()
	l, _ := minecraft.NewLevel(mem)

	bedrock := minecraft.Block{ID: 7}

	l.SetBlock(0, 0, 0, minecraft.Block{})
	l.Save()
	l.Close()
	clearChunk, _ := mem.GetChunk(0, 0)

	for j := int32(0); j < 255; j++ {
		l.SetBlock(-1, j, -1, bedrock)
		l.SetBlock(-1, j, 16, bedrock)
		l.SetBlock(16, j, -1, bedrock)
		l.SetBlock(16, j, 16, bedrock)
		for i := int32(0); i < 16; i++ {
			l.SetBlock(i, j, -1, bedrock)
			l.SetBlock(i, j, 16, bedrock)
			l.SetBlock(-1, j, i, bedrock)
			l.SetBlock(16, j, i, bedrock)
		}
	}
	l.Save()
	l.Close()
	mem.SetChunk(clearChunk)
	return &chunkCache{
		mem,
		l,
		clearChunk,
		make(map[uint16]nbt.Tag),
	}
}

func (c *chunkCache) getFromCache(x, z int32, terrain uint8, height int32) nbt.Tag {
	cacheID := uint16(terrain)<<8 | uint16(height)
	chunk, ok := c.cache[cacheID]
	if !ok {
		b := terrainBlocks[terrain].Base
		closest := c.clear
		var (
			closestLevel int32
			cl           int32
			h            int32
		)
		for {
			cl++
			h = height - cl
			if h == 0 {
				break
			}
			if chunk, ok := c.cache[uint16(terrain)<<8|uint16(h)]; ok {
				closestLevel = h
				closest = chunk
				break
			}
			h = height + cl
			if h > 255 {
				continue
			}
			if chunk, ok := c.cache[uint16(terrain)<<8|uint16(h)]; ok {
				closestLevel = h
				closest = chunk
				break
			}
		}
		ld := closest.Data().(nbt.Compound).Get("Level").Data().(nbt.Compound)
		ld.Set(nbt.NewTag("xPos", nbt.Int(0)))
		ld.Set(nbt.NewTag("zPos", nbt.Int(0)))
		c.mem.SetChunk(closest)
		if closestLevel < height {
			for j := height - 1; j >= closestLevel; j-- {
				for i := int32(0); i < 16; i++ {
					for k := int32(0); k < 16; k++ {
						c.level.SetBlock(i, j, k, b)
					}
				}
			}
		} else {
			for j := closestLevel; j > height; j-- {
				for i := int32(0); i < 16; i++ {
					for k := int32(0); k < 16; k++ {
						c.level.SetBlock(i, j, k, minecraft.Block{})
					}
				}
			}
		}
		c.level.Save()
		c.level.Close()
		chunk, _ = c.mem.GetChunk(0, 0)
		c.cache[cacheID] = chunk
	}
	ld := chunk.Data().(nbt.Compound).Get("Level").Data().(nbt.Compound)
	ld.Set(nbt.NewTag("xPos", nbt.Int(x)))
	ld.Set(nbt.NewTag("zPos", nbt.Int(z)))
	return chunk
}

func buildTerrain(mpath minecraft.Path, level *minecraft.Level, terrain, biomes *image.Paletted, height, water *image.Gray, c chan paint) error {
	b := terrain.Bounds()
	proceed := make(chan uint8, 10)
	errChan := make(chan error, 1)
	go func() {
		defer close(proceed)
		cc := newCache()
		for j := 0; j < b.Max.Y; j += 16 {
			chunkZ := int32(j >> 4)
			for i := 0; i < b.Max.X; i += 16 {
				chunkX := int32(i >> 4)
				p := terrain.SubImage(image.Rect(i, j, i+16, j+16)).(*image.Paletted)
				g := height.SubImage(image.Rect(i, j, i+16, j+16)).(*image.Gray)
				w := water.SubImage(image.Rect(i, j, i+16, j+16)).(*image.Gray)
				h := int32(meanHeight(g))
				wh := int32(meanHeight(w))
				var t uint8
				if wh >= h<<1 { // more water than land...
					c <- paint{
						color.RGBA{0, 0, 255, 255},
						chunkX, chunkZ,
					}
					t = uint8(len(terrainBlocks) - 1)
					h = wh
				} else {
					t = modeTerrain(p)
					c <- paint{
						terrainColours[t],
						chunkX, chunkZ,
					}
				}
				if err := mpath.SetChunk(cc.getFromCache(chunkX, chunkZ, t, h)); err != nil {
					errChan <- err
					return
				}
				proceed <- t
			}
		}
	}()
	ts := make([]uint8, 0, 1024)
	for i := 0; i < (b.Max.X>>4)+2; i++ {
		ts = append(ts, <-proceed) // get far enough ahead so all chunks are surrounded before shaping, to get correct lighting
	}
	select {
	case err := <-errChan:
		return err
	default:
	}
	for j := int32(0); j < int32(b.Max.Y); j += 16 {
		chunkZ := j >> 4
		for i := int32(0); i < int32(b.Max.X); i += 16 {
			chunkX := i >> 4
			var totalHeight int32
			ot := ts[0]
			ts = ts[1:]
			oy, _ := level.GetHeight(i, j)
			for x := i; x < i+16; x++ {
				for z := j; z < j+16; z++ {
					if biomes != nil {
						level.SetBiome(x, z, biomeList[biomes.ColorIndexAt(int(x), int(z))])
					}
					h := int32(height.GrayAt(int(x), int(z)).Y)
					totalHeight += h
					wl := int32(water.GrayAt(int(x), int(z)).Y)
					y := oy
					if h > y {
						y = h
					}
					if wl > y {
						y = wl
					}
					for ; y > h && y > wl; y-- {
						level.SetBlock(x, y, z, minecraft.Block{})
					}
					for ; y > h; y-- {
						level.SetBlock(x, y, z, minecraft.Block{ID: 9})
					}
					t := terrain.ColorIndexAt(int(x), int(z))
					tb := terrainBlocks[t]
					for ; y > h-int32(tb.TopLevel); y-- {
						level.SetBlock(x, y, z, tb.Top)
					}
					if t != ot {
						h = 0
					} else {
						h = oy
					}
					for ; y >= h; y-- {
						level.SetBlock(x, y, z, tb.Base)
					}
				}
			}
			c <- paint{
				color.Alpha{uint8(totalHeight >> 8)},
				chunkX, chunkZ,
			}
			select {
			case p, ok := <-proceed:
				if ok {
					ts = append(ts, p)
				}
			case err := <-errChan:
				return err
			}
		}
	}
	return nil
}
