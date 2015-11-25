package data

import (
	"image/color"

	"github.com/MJKWoolnough/minecraft"
)

type Blocks struct {
	Base, Top minecraft.Block
	Level     int
}

type ColourBlocks struct {
	Colour color.RGBA
	Blocks Blocks
	Name   string
}

type ColourBiome struct {
	Colour color.RGBA
	Biome  minecraft.Biome
	Name   string
}

type GeneratorData struct {
	Terrain []ColourBlocks
	Biomes  []ColourBiome
	Plants  []ColourBlocks
	Options map[string]string
}
