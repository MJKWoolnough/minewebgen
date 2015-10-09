package main

import (
	"errors"
	"strconv"

	"github.com/MJKWoolnough/gopherjs/overlay"
	"github.com/MJKWoolnough/gopherjs/style"
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

func init() {
	style.Add(`
.serverUnassigned {
	color : #f00;
}
	`)
}

func maps(c dom.Element) {
	xjs.RemoveChildren(c)
	mapsDiv := xdom.Div()
	newButton := xdom.Input()
	newButton.Type = "button"
	newButton.Value = "New Map"
	newButton.AddEventListener("click", false, newMap(c))

	mapsDiv.AppendChild(newButton)

	mapsTable := xdom.Table()
	mapsHeader := xdom.Tr()
	mapsHeader.AppendChild(xjs.SetInnerText(xdom.Th(), "Map Name"))
	mapsHeader.AppendChild(xjs.SetInnerText(xdom.Th(), "Server"))
	mapsTable.AppendChild(mapsHeader)
	go func() {
		list, err := RPC.MapList()
		if err != nil {
			xjs.SetInnerText(mapsDiv, err.Error())
			return
		}

		for _, m := range list {
			mr := xdom.Tr()
			mn := xdom.Td()
			xjs.SetInnerText(mn, m.Name)
			mn.AddEventListener("click", false, viewMap(m))
			ms := xdom.Td()
			s, err := RPC.GetServer(m.Server)
			if err != nil {
				xjs.SetInnerText(ms, "[Error]")
				ms.SetClass("serverUnassigned")
			} else if m.Server >= 0 {
				xjs.SetInnerText(ms, s.Name)
				ms.AddEventListener("click", false, assignServer(c, m, s))
			} else {
				xjs.SetInnerText(ms, "[Unassigned]")
				ms.SetClass("serverUnassigned")
				ms.AddEventListener("click", false, assignServer(c, m, Server{ID: -1}))
			}
			if m.Server >= 0 {
			}
			mr.AppendChild(mn)
			mr.AppendChild(ms)
			mapsTable.AppendChild(mr)
		}
	}()
	mapsDiv.AppendChild(mapsTable)
	c.AppendChild(mapsDiv)
}

func newMap(c dom.Element) func(dom.Event) {
	return func(dom.Event) {
		f := xdom.Div()
		o := overlay.New(f)
		f.AppendChild(xjs.SetInnerText(xdom.H1(), "New Map"))
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
	c := xdom.Div()
	nameLabel := xdom.Label()
	nameLabel.For = "name"
	xjs.SetInnerText(nameLabel, "Level Name")

	name := xdom.Input()
	name.Type = "text"
	name.SetID("name")

	gameModeLabel := xdom.Label()
	gameModeLabel.For = "gameMode"
	xjs.SetInnerText(gameModeLabel, "Game Mode")

	gameMode := xdom.Select()
	for k, v := range gameModes {
		o := xdom.Option()
		o.Value = strconv.Itoa(k)
		xjs.SetInnerText(o, v)
		gameMode.AppendChild(o)
	}

	seedLabel := xdom.Label()
	seedLabel.For = "seed"
	xjs.SetInnerText(seedLabel, "Level Seed")

	seed := xdom.Input()
	seed.Type = "text"
	seed.SetID("seed")
	seed.Value = ""

	structuresLabel := xdom.Label()
	structuresLabel.For = "structures"
	xjs.SetInnerText(structuresLabel, "Generate Structures")

	structures := xdom.Input()
	structures.Type = "checkbox"
	structures.Checked = true
	structures.SetID("structures")

	cheatsLabel := xdom.Label()
	cheatsLabel.For = "cheats"
	xjs.SetInnerText(cheatsLabel, "Allow Cheats")

	cheats := xdom.Input()
	cheats.Type = "checkbox"
	cheats.Checked = false
	cheats.SetID("cheats")

	c.AppendChild(nameLabel)
	c.AppendChild(name)
	c.AppendChild(xdom.Br())
	c.AppendChild(gameModeLabel)
	c.AppendChild(gameMode)
	c.AppendChild(xdom.Br())
	c.AppendChild(seedLabel)
	c.AppendChild(seed)
	c.AppendChild(xdom.Br())
	c.AppendChild(structuresLabel)
	c.AppendChild(structures)
	c.AppendChild(xdom.Br())
	c.AppendChild(cheatsLabel)
	c.AppendChild(cheats)
	c.AppendChild(xdom.Br())
	c.AppendChild(xdom.Br())

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
	submit := xdom.Input()
	submit.Type = "button"
	submit.Value = "Create Map"
	submit.AddEventListener("click", false, func(dom.Event) {
		data, err := dataParser()
		if err != nil {
			dom.GetWindow().Alert(err.Error())
			return
		}
		go func() {
			err = RPC.CreateDefaultMap(data)
			if err != nil {
				dom.GetWindow().Alert(err.Error())
			}
			o.Close()
		}()
	})
	return func(c dom.Element) {
		d := xdom.Div()
		xjs.SetPreText(d, worldTypes[mode])
		c.AppendChild(d)
		c.AppendChild(xdom.Br())
		c.AppendChild(submit)
	}
}

func createSuperFlatMap(o overlay.Overlay, dataParser func() (DefaultMap, error)) func(dom.Element) {
	d := xdom.Div()
	return func(c dom.Element) {
		c.AppendChild(d)
	}
}

func createCustomisedMap(o overlay.Overlay, dataParser func() (DefaultMap, error)) func(dom.Element) {
	d := xdom.Div()
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
		go func() {
			d := xdom.Div()
			od := overlay.New(d)
			d.AppendChild(xjs.SetInnerText(xdom.H1(), "Map Details"))

			nameLabel := xdom.Label()
			nameLabel.For = "name"
			xjs.SetInnerText(nameLabel, "Name")
			name := xdom.Input()
			xjs.SetInnerText(nameLabel, "Name")
			name.SetID("name")
			name.Value = m.Name
			name.Type = "text"

			submit := xdom.Input()
			submit.Type = "button"
			submit.Value = "Make Changes"
			submit.AddEventListener("click", false, func(dom.Event) {
				if name.Value == "" {
					dom.GetWindow().Alert("Name cannot be empty")
					return
				}
				m.Name = name.Value
				go func() {
					err := RPC.SetMap(m)
					if err != nil {
						dom.GetWindow().Alert(err.Error())
					}
				}()
			})

			d.AppendChild(nameLabel)
			d.AppendChild(name)
			d.AppendChild(xdom.Br())
			d.AppendChild(submit)

			dom.GetWindow().Document().DocumentElement().AppendChild(od)
		}()
	}
}

func assignServer(c dom.Element, m Map, s Server) func(dom.Event) {
	return func(dom.Event) {
		go func() {
			servers, err := RPC.ServerList()
			if err != nil {
				return
			}
			d := xdom.Div()
			od := overlay.New(d)
			d.AppendChild(xjs.SetInnerText(xdom.H1(), "Map Server Assignment"))

			od.OnClose(func() {
				maps(c)
			})

			serverLabel := xdom.Label()
			serverLabel.For = "server"
			xjs.SetInnerText(serverLabel, "Server")
			d.AppendChild(serverLabel)
			if m.Server < 0 {
				sel := xdom.Select()
				sel.SetID("server")
				sel.AppendChild(xjs.SetInnerText(xdom.Option(), "--"))
				for _, s := range servers {
					if s.Map != -1 {
						continue
					}
					o := xdom.Option()
					o.Value = strconv.Itoa(s.ID)
					xjs.SetInnerText(o, s.Name)
					if s.ID == m.Server {
						o.Selected = true
					}
					sel.AppendChild(o)
				}
				d.AppendChild(sel)
				if len(servers) > 0 {
					assign := xdom.Input()
					assign.Type = "button"
					assign.Value = "Set Server"
					assign.AddEventListener("click", false, func(dom.Event) {
						sID, err := strconv.Atoi(sel.Value)
						if err != nil {
							return
						}
						go func() {
							err = RPC.SetServerMap(m.ID, sID)
							if err != nil {
								dom.GetWindow().Alert(err.Error())
							} else {
								od.Close()
								for _, ts := range servers {
									if ts.ID == sID {
										s = ts
										break
									}
								}
								m.Server = sID
								assignServer(c, m, s)(nil)
							}
						}()
					})
					d.AppendChild(assign)
				}
			} else {
				d.AppendChild(xjs.SetInnerText(xdom.Div(), s.Name))
				if !s.IsRunning() {
					remove := xdom.Input()
					remove.Type = "button"
					remove.Value = "X"
					remove.AddEventListener("click", false, func(dom.Event) {
						go func() {
							err := RPC.RemoveServerMap(m.Server)
							if err != nil {
								dom.GetWindow().Alert(err.Error())
							} else {
								od.Close()
								m.Server = -1
								assignServer(c, m, Server{ID: -1})(nil)
							}
						}()
					})
					d.AppendChild(remove)
				}
			}

			dom.GetWindow().Document().DocumentElement().AppendChild(od)
		}()
	}
}

// Errors
var ErrInvalidGameMode = errors.New("invalid game mode")
