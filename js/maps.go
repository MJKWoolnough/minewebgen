package main

import (
	"errors"
	"strconv"

	"github.com/MJKWoolnough/gopherjs/overlay"
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

func maps(c dom.Element) {
	xjs.RemoveChildren(c)
	mapsDiv := xjs.CreateElement("div")
	defer c.AppendChild(mapsDiv)
	list, err := MapList()
	if err != nil {
		xjs.SetInnerText(mapsDiv, err.Error())
		return
	}

	newButton := xjs.CreateElement("input").(*dom.HTMLInputElement)
	newButton.Type = "button"
	newButton.Value = "New Map"
	newButton.AddEventListener("click", false, newMap(c))

	mapsDiv.AppendChild(newButton)

	for _, m := range list {
		sd := xjs.CreateElement("div")
		xjs.SetInnerText(sd, m.Name)
		sd.AddEventListener("click", false, viewMap(m))
		mapsDiv.AppendChild(sd)
	}
	c.AppendChild(mapsDiv)
}

func newMap(c dom.Element) func(dom.Event) {
	return func(dom.Event) {
		f := xjs.CreateElement("div")
		o := overlay.New(f)
		f.AppendChild(xjs.SetInnerText(xjs.CreateElement("h1"), "New Map"))
		f.AppendChild(tabs.MakeTabs([]tabs.Tab{
			{"Create", createMap(o)},
			{"Upload/Download", uploadMap(o)},
			{"Generate", generate},
		}))
		o.OnClose(func() {
			maps(c)
		})
		c.AppendChild(o)
	}
}

var gameModes = [...]string{"Survival", "Creative", "Adventure", "Hardcore", "Spectator"}

func createMap(o overlay.Overlay) func(dom.Element) {
	c := xjs.CreateElement("div")
	nameLabel := xjs.CreateElement("label").(*dom.HTMLLabelElement)
	nameLabel.For = "name"
	xjs.SetInnerText(nameLabel, "Level Name")

	name := xjs.CreateElement("input").(*dom.HTMLInputElement)
	name.Type = "text"
	name.SetID("name")

	gameModeLabel := xjs.CreateElement("label").(*dom.HTMLLabelElement)
	gameModeLabel.For = "gameMode"
	xjs.SetInnerText(gameModeLabel, "Game Mode")

	gameMode := xjs.CreateElement("select").(*dom.HTMLSelectElement)
	for k, v := range gameModes {
		o := xjs.CreateElement("option").(*dom.HTMLOptionElement)
		o.Value = strconv.Itoa(k)
		xjs.SetInnerText(o, v)
		gameMode.AppendChild(o)
	}

	seedLabel := xjs.CreateElement("label").(*dom.HTMLLabelElement)
	seedLabel.For = "seed"
	xjs.SetInnerText(seedLabel, "Level Seed")

	seed := xjs.CreateElement("input").(*dom.HTMLInputElement)
	seed.Type = "text"
	seed.SetID("seed")
	seed.Value = ""

	structuresLabel := xjs.CreateElement("label").(*dom.HTMLLabelElement)
	structuresLabel.For = "structures"
	xjs.SetInnerText(structuresLabel, "Generate Structures")

	structures := xjs.CreateElement("input").(*dom.HTMLInputElement)
	structures.Type = "checkbox"
	structures.Checked = true
	structures.SetID("structures")

	cheatsLabel := xjs.CreateElement("label").(*dom.HTMLLabelElement)
	cheatsLabel.For = "cheats"
	xjs.SetInnerText(cheatsLabel, "Allow Cheats")

	cheats := xjs.CreateElement("input").(*dom.HTMLInputElement)
	cheats.Type = "checkbox"
	cheats.Checked = false
	cheats.SetID("cheats")

	c.AppendChild(nameLabel)
	c.AppendChild(name)
	c.AppendChild(xjs.CreateElement("br"))
	c.AppendChild(gameModeLabel)
	c.AppendChild(gameMode)
	c.AppendChild(xjs.CreateElement("br"))
	c.AppendChild(seedLabel)
	c.AppendChild(seed)
	c.AppendChild(xjs.CreateElement("br"))
	c.AppendChild(structuresLabel)
	c.AppendChild(structures)
	c.AppendChild(xjs.CreateElement("br"))
	c.AppendChild(cheatsLabel)
	c.AppendChild(cheats)
	c.AppendChild(xjs.CreateElement("br"))
	c.AppendChild(xjs.CreateElement("br"))

	dataParser := func(mode int) func() (DefaultMap, error) {
		return func() (DefaultMap, error) {
			data := DefaultMap{
				Mode: mode,
			}
			var err error
			data.Name = name.Value
			si := gameMode.SelectedIndex
			if si < 0 || si >= len(gameModes) {
				return data, ErrInvalidGameMode
			}
			if seed.Value == "" {
				seed.Value = "0"
			}
			data.Seed, err = strconv.ParseInt(seed.Value, 10, 64)
			if err != nil {
				return data, err
			}
			data.Structures = structures.Checked
			data.Cheats = cheats.Checked
			return data, nil
		}
	}

	c.AppendChild(tabs.MakeTabs([]tabs.Tab{
		{"Default", createMapMode(0, o, dataParser(0))},
		{"Super Flat", createSuperFlatMap(o, dataParser(1))},
		{"Large Biomes", createMapMode(2, o, dataParser(2))},
		{"Amplified", createMapMode(3, o, dataParser(3))},
		{"Customised", createCustomisedMap(o, dataParser(4))},
	}))
	return func(d dom.Element) {
		d.AppendChild(c)
	}
}

var worldTypes = [...]string{
	"The standard minecraft map generation.",
	"A simple generator allowing customised levels of blocks.",
	"The standard minecraft map generation, but tweaked to allow for much larger biomes.",
	"The standard minecraft map generation, but tweaked to stretch the land upwards.",
	"A completely customiseable generator.",
}

func createMapMode(mode int, o overlay.Overlay, dataParser func() (DefaultMap, error)) func(dom.Element) {
	submit := xjs.CreateElement("input").(*dom.HTMLInputElement)
	submit.Type = "button"
	submit.Value = "Create Map"
	submit.AddEventListener("click", false, func(dom.Event) {
		data, err := dataParser()
		if err != nil {
			return
		}
		go func() {
			err = CreateDefaultMap(data)
			if err != nil {
				dom.GetWindow().Alert(err.Error())
			}
			o.Close()
		}()
	})
	return func(c dom.Element) {
		d := xjs.CreateElement("div")
		xjs.SetPreText(d, worldTypes[mode])
		c.AppendChild(d)
		c.AppendChild(xjs.CreateElement("br"))
		c.AppendChild(submit)
	}
}

func createSuperFlatMap(o overlay.Overlay, dataParser func() (DefaultMap, error)) func(dom.Element) {
	d := xjs.CreateElement("div")
	return func(c dom.Element) {
		c.AppendChild(d)
	}
}

func createCustomisedMap(o overlay.Overlay, dataParser func() (DefaultMap, error)) func(dom.Element) {
	d := xjs.CreateElement("div")
	return func(c dom.Element) {
		c.AppendChild(d)
	}
	// Sea Level - 0-255
	// Caves, Strongholds, Villages, Mineshafts, Temples, Ocean Monuments, Ravines
	// Dungeons + Count 1-100
	// Water Lakes + Rarity 1-100
	// Lava Lakes + Rarity 1-100
	// Lava Oceans
	// Biome - All/Choose
	// Biome Size 1-8
	// River Size 1-5
	// Ores -> Dirt/Gravel/Granite/Diorite/Andesite/Coal Ore/Iron Ore/Gold Ore/Redstone Ore/Diamond Ore/Lapis Lazuli Ore ->
	//           Spawn Size - 1-50
	//           Spawn Tries - 0-40
	//           Min-Height - 0-255
	//           Max-Height - 0-255
	// Advanced ->
	//           Main Noise Scale X - 1-5000
	//           Main Noise Scale Y - 1-5000
	//           Main Noise Scale Z - 1-5000
	//           Depth Noise Scale X - 1-2000
	//           Depth Noise Scale Y - 1-2000
	//           Depth Noise Scale Z - 1-2000
	//           Depth Base Size - 1-25
	//           Coordinate Scale - 1-6000
	//           Height Scale - 1-6000
	//           Height Stretch - 0.01-50
	//           Upper Limit Scale - 1-5000
	//           Lower Limit Scale - 1-5000
	//           Biome Depth Weight - 1-20
	//           Biome Depth Offset - 1-20
	//           Biome Scale Weight - 1-20
	//           Biome Scale Offset - 1-20

}

func uploadMap(o overlay.Overlay) func(dom.Element) {
	return func(c dom.Element) {
	}
}

func viewMap(m Map) func(dom.Event) {
	return func(dom.Event) {
		d := xjs.CreateElement("div")
		od := overlay.New(d)
		d.AppendChild(xjs.SetInnerText(xjs.CreateElement("h1"), "Map Details"))

		nameLabel := xjs.CreateElement("label").(*dom.HTMLLabelElement)
		nameLabel.For = "name"
		xjs.SetInnerText(nameLabel, "Name")
		name := xjs.CreateElement("input").(*dom.HTMLInputElement)
		xjs.SetInnerText(nameLabel, "Name")
		name.Value = m.Name
		name.Type = "text"

		d.AppendChild(nameLabel)
		d.AppendChild(name)

		dom.GetWindow().Document().DocumentElement().AppendChild(od)
	}
}

// Errors
var ErrInvalidGameMode = errors.New("invalid game mode")
