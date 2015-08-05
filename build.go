package main

import (
	"image/color"

	"github.com/MJKWoolnough/minecraft"
	"github.com/MJKWoolnough/ora"
)

type terrain struct {
	Base, Top minecraft.Block
	TopLevel  uint8
}

var (
	terrainColours = color.Palette{
		color.RGBA{255, 255, 0, 255},   // Yellow - Sand
		color.RGBA{0, 255, 0, 255},     // Green - Grass
		color.RGBA{87, 59, 12, 255},    // Brown - Dirt
		color.RGBA{255, 128, 0, 255},   // Orange - Farm
		color.RGBA{127, 127, 127, 255}, // Grey - Stone
		color.RGBA{255, 255, 255, 255}, // White - Snow
	}
	terrainBlocks = []terrain{
		{minecaft.Block{ID: 24, Data: 2}, minecaft.Block{ID: 12}, 5}, // Sandstone - Sand
		{minecraft.Block{ID: 3}, minecaft.Block{2}, 1},               // Dirt - Grass
		{minecaft.Block{ID: 3}, minecaft.Block{ID: 3}, 0},            // Dirt - Dirt
		{minecaft.Block{ID: 3}, minecaft.Block{ID: 60, Data: 7}, 1},  // Dirt - Farmland
		{minecaft.Block{ID: 1}, minecaft.Block{ID: 1}, 0},            // Stone - Stone
		{minecaft.Block{ID: 1}, minecaft.Block{ID: 50}, 3},           // Stone - Snow
	}
)

func buildMap(o *ora.ORA, c chan paint) {
	defer close(c)
}
