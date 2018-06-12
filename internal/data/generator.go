package data // import "vimagination.zapto.org/minewebgen/internal/data"

import (
	"image/color"

	"vimagination.zapto.org/minecraft"
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
