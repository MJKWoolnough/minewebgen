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
		color.RGBA{127, 127, 127, 255}, // Grey - Stone
		color.RGBA{87, 59, 12, 255},    // Brown - Dirt
		color.RGBA{255, 128, 0, 255},   // Orange - Farm
	}
	terrainBlocks = []terrain{
		{minecaft.Block{ID: 24, Data: 2}, minecaft.Block{ID: 12}, 5},
		{minecraft.Block{ID: 3}, minecaft.Block{2}, 1},
		{minecaft.Block{ID: 1}, minecaft.Block{}, 0},
		{minecaft.Block{ID: 3}, minecaft.Block{}, 0},
		{minecaft.Block{ID: 3}, minecaft.Block{ID: 60, Data: 7}, 1},
	}
)

func buildMap(o *ora.ORA, c chan paint) {
	defer close(c)
}
