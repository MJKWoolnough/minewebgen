package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/MJKWoolnough/minecraft"
	"github.com/MJKWoolnough/minecraft/nbt"
	"github.com/MJKWoolnough/ora"
)

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
		{minecraft.Block{ID: 1}, minecraft.Block{ID: 50}, 3},           // Stone - Snow
	}
)

func modeTerrain(p *image.Paletted) uint8 {
	b := p.Bounds()
	var modeMap [7]uint16
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

	for i := int32(-16); i < 16; i++ {
		for j := int32(0); j < 255; j++ {
			for k := int32(-16); j < 16; j++ {
				l.SetBlock(i, j, k, bedrock)
			}
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
		for j := int32(0); j < height; j++ {
			for i := int32(0); i < 16; i++ {
				for k := int32(0); k < 16; k++ {
					c.level.SetBlock(i, j, k, b)
				}
			}
		}
		c.level.Save()
		c.level.Close()
		chunk, _ = c.mem.GetChunk(0, 0)
		c.mem.SetChunk(c.clear)
		c.cache[cacheID] = chunk
	}
	ld := chunk.Data().(nbt.Compound).Get("Level").Data().(nbt.Compound)
	ld.Set(nbt.NewTag("xPos", nbt.Int(x)))
	ld.Set(nbt.NewTag("zPos", nbt.Int(z)))
	return chunk
}

func buildMap(o *ora.ORA, c chan paint, m chan string, e chan error) {
	defer close(e)
	terrain := o.Layer("terrain")
	height := o.Layer("height")
	sTerrain := image.NewPaletted(o.Bounds(), terrainColours)
	terrainI, err := terrain.Image()
	if err != nil {
		e <- err
		return
	}
	draw.Draw(sTerrain, image.Rect(terrain.X, terrain.Y, sTerrain.Bounds().Max.X, sTerrain.Bounds().Max.Y), terrainI, image.Point{}, draw.Src)
	terrainI = nil
	sHeight := image.NewGray(o.Bounds())
	heightI, err := height.Image()
	if err != nil {
		e <- err
		return
	}
	draw.Draw(sHeight, image.Rect(height.X, height.Y, sTerrain.Bounds().Max.X, sTerrain.Bounds().Max.Y), heightI, image.Point{}, draw.Src)
	heightI = nil
	p, err := minecraft.NewFilePath("./test/")
	if err != nil {
		e <- err
		return
	}

	m <- "Building Terrain"
	buildTerrain(p, sTerrain, sHeight, c)
	m <- "Building Height Map"
	level, err := minecraft.NewLevel(p)
	if err != nil {
		e <- err
		return
	}
	level.Save()
}

func buildTerrain(mpath minecraft.Path, terrain *image.Paletted, height *image.Gray, c chan paint) {
	cc := newCache()
	b := terrain.Bounds()
	for j := 0; j < b.Max.Y; j += 16 {
		chunkZ := int32(j >> 4)
		for i := 0; i < b.Max.X; i += 16 {
			chunkX := int32(i >> 4)
			p := terrain.SubImage(image.Rect(i, j, i+16, j+16)).(*image.Paletted)
			g := height.SubImage(image.Rect(i, j, i+16, j+16)).(*image.Gray)
			t := modeTerrain(p)
			h := int32(meanHeight(g))
			mpath.SetChunk(cc.getFromCache(chunkX, chunkZ, t, h))
			c <- paint{
				terrainColours[t],
				chunkX, chunkZ,
				nil,
			}
		}
	}
}

func buildHeight(level *minecraft.Level, terrain *image.Paletted, height *image.Gray, c chan paint) {

}
