package main

import (
	"image"
	"image/color"
	"image/draw"
	"strconv"
	"strings"

	"github.com/MJKWoolnough/minecraft"
	"github.com/MJKWoolnough/minecraft/nbt"
	"github.com/MJKWoolnough/minewebgen/internal/data"
	"github.com/MJKWoolnough/ora"
)

func toGray(o *ora.ORA, name string) (*image.Gray, error) {
	var p *image.Gray
	if l := o.Layer(name); l != nil {
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
	if l := o.Layer(name); l != nil {
		p = image.NewPaletted(o.Bounds(), palette)
		i, err := l.Image()
		if err != nil {
			return nil, err
		}
		draw.Draw(p, image.Rect(0, 0, p.Bounds().Max.X, p.Bounds().Max.Y), i, image.Point{}, draw.Src)
	}
	return p, nil
}

type level struct {
	*minecraft.Level
	memoryLimit int64
}

func (l *level) checkMem() {

}

func (l *level) GetBiome(x, z int32) (minecraft.Biome, error) {
	l.checkMem()
	return l.Level.GetBiome(x, z)
}

func (l *level) GetBlock(x, y, z int32) (minecraft.Block, error) {
	l.checkMem()
	return l.Level.GetBlock(x, y, z)
}

func (l *level) GetHeight(x, z int32) (int32, error) {
	l.checkMem()
	return l.Level.GetHeight(x, z)
}

func (l *level) SetBiome(x, z int32, biome minecraft.Biome) error {
	l.checkMem()
	return l.Level.SetBiome(x, z, biome)
}

func (l *level) SetBlock(x, y, z int32, block minecraft.Block) error {
	l.checkMem()
	return l.Level.SetBlock(x, y, z, block)
}

type generator struct {
	generator data.GeneratorData
	Terrain   struct {
		Blocks  []data.Blocks
		Palette color.Palette
	}
	Biomes struct {
		Values  []minecraft.Biome
		Palette color.Palette
	}
	Plants struct {
		Blocks  []data.Blocks
		Palette color.Palette
	}
}

func (g *generator) Generate(name, mapPath string, o *ora.ORA, c chan paint, m chan string, memoryLimit int64) error {
	sTerrain, err := toPaletted(o, "terrain", g.Terrain.Palette)
	if err != nil {
		return err
	}
	if sTerrain == nil {
		return layerError{"terrain"}
	}

	sHeight, err := toGray(o, "height")
	if err != nil {
		return err
	}
	if sHeight == nil {
		return layerError{"height"}
	}

	sBiomes, err := toPaletted(o, "biomes", g.Biomes.Palette)
	if err != nil {
		return err
	}
	sWater, err := toGray(o, "water")
	if err != nil {
		return err
	}
	sPlants, err := toPaletted(o, "plants", g.Plants.Palette)
	if err != nil {
		return err
	}

	p, err := minecraft.NewFilePath(mapPath)
	if err != nil {
		return err
	}

	l, err := minecraft.NewLevel(p)
	if err != nil {
		return err
	}
	level := &level{l, memoryLimit}

	level.LevelName(name)

	m <- "Building Terrain"
	if err = g.buildTerrain(p, level, sTerrain, sBiomes, sPlants, sHeight, sWater, c); err != nil {
		return err
	}

	level.LevelName(name)
	level.Generator(minecraft.FlatGenerator)
	level.GeneratorOptions("0")
	level.GameMode(minecraft.Creative)

	for k, v := range g.generator.Options {
		v = strings.ToLower(v)
		switch strings.ToLower(k) {
		case "generate-structures":
			level.MapFeatures(v != "false")
		case "hardcore":
			level.Hardcore(v != "false")
		case "gamemode":
			gm, _ := strconv.Atoi(v)
			if gm >= 0 && gm <= 3 {
				level.GameMode(int32(gm))
			}
		case "difficulty":
			d, _ := strconv.Atoi(v)
			if d >= 0 && d <= 3 {
				level.Difficulty(int8(d))
			}
		case "daylight-cycle":
			level.DayLightCycle(v != "false")
		case "fire-tick":
			level.FireTick(v != "false")
		case "keep-inventory":
			level.KeepInventory(v != "false")
		}
	}

	level.AllowCommands(true)
	level.MobSpawning(false)
	level.MobGriefing(false)
	level.Spawn(10, 250, 10)

	m <- "Exporting"
	level.Save()
	level.Close()
	return nil
}

type layerError struct {
	name string
}

func (l layerError) Error() string {
	return "missing layer: " + l.name
}

type blocks struct {
	Base, Top minecraft.Block
	TopLevel  uint8
}

func modeTerrain(p *image.Paletted, l int) uint8 {
	b := p.Bounds()
	modeMap := make([]uint8, l)
	var most, mode uint8
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
	var total uint64
	for i := b.Min.X; i < b.Max.X; i++ {
		for j := b.Min.Y; j < b.Max.Y; j++ {
			total += uint64(g.GrayAt(i, j).Y)
		}
	}
	return uint8(total / uint64((b.Dx() * b.Dy())))
}

type chunkCache struct {
	mem    *minecraft.MemPath
	level  *minecraft.Level
	clear  nbt.Tag
	cache  map[uint16]nbt.Tag
	blocks []data.Blocks
}

func newCache(blocks []data.Blocks) *chunkCache {
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
		mem:    mem,
		level:  l,
		clear:  clearChunk,
		cache:  make(map[uint16]nbt.Tag),
		blocks: blocks,
	}
}

func (c *chunkCache) getFromCache(x, z int32, terrain uint8, height int32) nbt.Tag {
	cacheID := uint16(terrain)<<8 | uint16(height)
	chunk, ok := c.cache[cacheID]
	if !ok {
		b := c.blocks[terrain].Base
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

func (g *generator) buildTerrain(mpath minecraft.Path, level *level, terrain, biomes, plants *image.Paletted, height, water *image.Gray, c chan paint) error {
	b := terrain.Bounds()
	proceed := make(chan uint8, 10)
	errChan := make(chan error, 1)
	go func() {
		defer close(proceed)
		cc := newCache(g.Terrain.Blocks)
		for j := 0; j < b.Max.Y; j += 16 {
			chunkZ := int32(j >> 4)
			for i := 0; i < b.Max.X; i += 16 {
				chunkX := int32(i >> 4)
				h := int32(meanHeight(height.SubImage(image.Rect(i, j, i+16, j+16)).(*image.Gray)))
				wh := int32(meanHeight(water.SubImage(image.Rect(i, j, i+16, j+16)).(*image.Gray)))
				var t uint8
				if wh >= h<<1 { // more water than land...
					c <- paint{
						color.RGBA{0, 0, 255, 255},
						chunkX, chunkZ,
					}
					t = uint8(len(g.Terrain.Blocks) - 1)
					h = wh
				} else {
					t = modeTerrain(terrain.SubImage(image.Rect(i, j, i+16, j+16)).(*image.Paletted), len(g.Terrain.Palette))
					c <- paint{
						g.Terrain.Palette[t],
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
						level.SetBiome(x, z, g.Biomes.Values[biomes.ColorIndexAt(int(x), int(z))])
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
					if plants != nil {
						p := g.Plants.Blocks[plants.ColorIndexAt(int(x), int(z))]
						py := int32(1)
						for ; py <= int32(p.Level); py++ {
							level.SetBlock(x, y+py, z, p.Base)
						}
						level.SetBlock(x, y+py, z, p.Top)
					}
					for ; y > h; y-- {
						level.SetBlock(x, y, z, minecraft.Block{ID: 9})
					}
					t := terrain.ColorIndexAt(int(x), int(z))
					tb := g.Terrain.Blocks[t]
					for ; y > h-int32(tb.Level); y-- {
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
