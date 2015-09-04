package main

import (
	"strconv"

	"github.com/MJKWoolnough/gopherjs/overlay"
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

func maps(c dom.Element) {
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

func createMap(o overlay.Overlay) func(dom.Element) {
	return func(c dom.Element) {
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
		for k, v := range []string{"Survival", "Creative", "Adventure", "Hardcore", "Spectator"} {
			o := xjs.CreateElement("option").(*dom.HTMLOptionElement)
			o.Value = strconv.Itoa(k)
			xjs.SetInnerText(o, v)
			gameMode.AppendChild(o)
		}

		seedLabel := xjs.CreateElement("label").(*dom.HTMLLabelElement)
		seedLabel.For = "seed"
		xjs.SetInnerText(nameLabel, "Level Seed")

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

		bonusLabel := xjs.CreateElement("label").(*dom.HTMLLabelElement)
		bonusLabel.For = "bonus"
		xjs.SetInnerText(bonusLabel, "Bonus Chest")

		bonus := xjs.CreateElement("input").(*dom.HTMLInputElement)
		bonus.Type = "checkbox"
		bonus.Checked = false
		xjs.SetInnerText(bonus, "Bonus Chest")

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
		c.AppendChild(bonusLabel)
		c.AppendChild(bonus)
		c.AppendChild(xjs.CreateElement("br"))

		c.AppendChild(tabs.MakeTabs([]tabs.Tab{
			{"Default", createMapMode(0, o, name, seed, structures, cheats, bonus)},
			{"Super Flat", createSuperFlatMap(o, name, seed, structures, cheats, bonus)},
			{"Large Biomes", createMapMode(1, o, name, seed, structures, cheats, bonus)},
			{"Amplified", createMapMode(2, o, name, seed, structures, cheats, bonus)},
			{"Customised", createCustomisedMap(o, name, seed, structures, cheats, bonus)},
		}))
	}
}

func createMapMode(mode int, o overlay.Overlay, name, seed, structures, cheats, bonus *dom.HTMLInputElement) func(dom.Element) {
	submit := xjs.CreateElement("input").(*dom.HTMLInputElement)
	submit.Type = "button"
	submit.Value = "Create Map"
	submit.AddEventListener("click", false, func(dom.Event) {
		o.Close()
	})
	return func(c dom.Element) {
		c.AppendChild(submit)
	}
}

func createSuperFlatMap(o overlay.Overlay, name, seed, structures, cheats, bonus *dom.HTMLInputElement) func(dom.Element) {
	d := xjs.CreateElement("div")
	return func(c dom.Element) {
		c.AppendChild(d)
	}
}

func createCustomisedMap(o overlay.Overlay, name, seed, structures, cheats, bonus *dom.HTMLInputElement) func(dom.Element) {
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
