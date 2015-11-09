package main

import (
	"errors"
	"strconv"
	"time"

	"github.com/MJKWoolnough/gopherjs/overlay"
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xform"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/MJKWoolnough/minewebgen/internal/data"
	"honnef.co/go/js/dom"
)

func mapsTab(c dom.Element) {
	xjs.RemoveChildren(c)
	c.AppendChild(xjs.SetInnerText(xdom.H2(), "Maps"))
	nm := xdom.Button()
	c.AppendChild(xjs.SetInnerText(nm, "New Map"))
	nm.AddEventListener("click", false, func(dom.Event) {
		d := xdom.Div()
		o := overlay.New(d)
		o.OnClose(func() {
			go mapsTab(c)
		})
		xjs.AppendChildren(d,
			xjs.SetInnerText(xdom.H1(), "New Map"),
			tabs.New([]tabs.Tab{
				{"Create", createMap(o)},
				{"Upload/Download", func(c dom.Element) {
					c.AppendChild(transferFile("Map", "Upload/Download", 1, o))
				}},
				{"Generate", func(c dom.Element) {
					c.AppendChild(transferFile("Map", "Generate", 2, o))
				}},
			}),
		)
		xjs.Body().AppendChild(o)
	})
	m, err := RPC.MapList()
	if err != nil {
		c.AppendChild(xjs.SetInnerText(xdom.Div(), err.Error()))
		return
	}
	if len(m) == 0 {
		c.AppendChild(xjs.SetInnerText(xdom.Div(), "No Maps"))
		return
	}
	t := xjs.AppendChildren(xdom.Table(), xjs.AppendChildren(xdom.Thead(), xjs.AppendChildren(xdom.Tr(),
		xjs.SetInnerText(xdom.Th(), "Map Name"),
	)))

	for _, mp := range m {
		name := xjs.SetInnerText(xdom.Td(), mp.Name)
		name.AddEventListener("click", false, func() func(dom.Event) {
			m := mp
			return func(dom.Event) {
				o := overlay.New(xjs.AppendChildren(xdom.Div(), tabs.New([]tabs.Tab{
					{"General", mapGeneral(m)},
					{"Properties", mapProperties(m)},
				})))
				o.OnClose(func() {
					mapsTab(c)
				})
				xjs.Body().AppendChild(o)
			}
		}())
		t.AppendChild(xjs.AppendChildren(xdom.Tr(), name))
	}
	c.AppendChild(t)
}

var gameModes = [...]string{"Survival", "Creative", "Adventure", "Hardcore", "Spectator"}

func createMap(o *overlay.Overlay) func(dom.Element) {
	return func(c dom.Element) {
		name := xform.InputText("name", "")
		name.Required = true
		gmOpts := make([]xform.Option, len(gameModes))
		for i, m := range gameModes {
			gmOpts[i] = xform.Option{
				Label: m,
				Value: strconv.Itoa(i),
			}
		}
		gameMode := xform.SelectBox("gamemode", gmOpts...)
		seed := xform.InputText("seed", "")
		structures := xform.InputCheckbox("structures", true)
		cheats := xform.InputCheckbox("cheats", false)
		fs := xdom.Fieldset()
		fs.AppendChild(xjs.SetInnerText(xdom.Legend(), "Create Map"))
		c.AppendChild(xjs.AppendChildren(xdom.Form(), fs))
		dataParser := func(mode int) func() (data.DefaultMap, error) {
			return func() (data.DefaultMap, error) {
				data := data.DefaultMap{
					Mode: mode,
				}
				var err error
				data.Name = name.Value
				si := gameMode.SelectedIndex
				if si < 0 || si >= len(gameModes) {
					return data, errors.New("invalid gamemode")
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
		xjs.AppendChildren(fs,
			xform.Label("Name", "name"),
			name,
			xdom.Br(),
			xform.Label("Game Mode", "gamemode"),
			gameMode,
			xdom.Br(),
			xform.Label("Level Seed", "seed"),
			seed,
			xdom.Br(),
			xform.Label("Structures", "structures"),
			structures,
			xdom.Br(),
			xform.Label("Cheats", "cheats"),
			cheats,
			xdom.Br(),
			tabs.New([]tabs.Tab{
				{"Default", createMapMode(0, o, dataParser(0))},
				{"Super Flat", createSuperFlatMap(o, dataParser(1))},
				{"Large Biomes", createMapMode(2, o, dataParser(2))},
				{"Amplified", createMapMode(3, o, dataParser(3))},
				{"Customised", createCustomisedMap(o, dataParser(4))},
			}),
		)
	}
}

var worldTypes = [...]string{
	"The standard minecraft map generation.",
	"A simple generator allowing customised levels of blocks.",
	"The standard minecraft map generation, but tweaked to allow for much larger biomes.",
	"The standard minecraft map generation, but tweaked to stretch the land upwards.",
	"A completely customiseable generator.",
}

func createMapMode(mode int, o *overlay.Overlay, dataParser func() (data.DefaultMap, error)) func(dom.Element) {
	submit := xform.InputSubmit("Create Map")
	submit.AddEventListener("click", false, func(e dom.Event) {
		data, err := dataParser()
		if err != nil {
			xjs.Alert("Error parsing values: %s", err)
			return
		}
		e.PreventDefault()
		go func() {
			err = RPC.CreateDefaultMap(data)
			if err != nil {
				xjs.Alert("Error parsing values: %s", err)
				return
			}
			o.Close()

		}()
	})
	return func(c dom.Element) {
		xjs.AppendChildren(c,
			xjs.SetPreText(xdom.Div(), worldTypes[mode]),
			xdom.Br(),
			submit,
		)
	}
}

func createSuperFlatMap(o *overlay.Overlay, dataParser func() (data.DefaultMap, error)) func(dom.Element) {
	// create better UI here
	d := xdom.Div()
	gs := xform.InputText("settings", "0")
	submit := xform.InputSubmit("Create Map")
	xjs.AppendChildren(d,
		xjs.SetPreText(xdom.Div(), worldTypes[1]),
		xform.Label("Generator Settings", "settings"),
		gs,
		xdom.Br(),
		submit,
	)
	submit.AddEventListener("click", false, func(e dom.Event) {
		d, err := dataParser()
		if err != nil {
			xjs.Alert("Error parsing values: %s", err)
			return
		}
		e.PreventDefault()
		go func() {
			err = RPC.CreateSuperflatMap(data.SuperFlatMap{
				DefaultMap:        d,
				GeneratorSettings: gs.Value,
			})
			if err != nil {
				xjs.Alert("Error parsing values: %s", err)
				return
			}
			o.Close()

		}()
	})
	return func(c dom.Element) {
		c.AppendChild(d)
	}
}

func createCustomisedMap(o *overlay.Overlay, dataParser func() (data.DefaultMap, error)) func(dom.Element) {
	d := xdom.Div()
	return func(c dom.Element) {
		c.AppendChild(d)
	}
}

func mapGeneral(m data.Map) func(dom.Element) {
	return func(c dom.Element) {
		go func() {
			servers, err := RPC.ServerList()
			if err != nil {
				c.AppendChild(xjs.SetInnerText(xdom.Div(), "Error getting server list: "+err.Error()))
				return
			}
			name := xform.InputText("name", m.Name)
			name.Required = true
			opts := make([]xform.Option, 1, len(servers)+1)
			opts[0] = xform.Option{
				Label:    "-- None --",
				Value:    "-1",
				Selected: m.Server == -1,
			}
			for i, s := range servers {
				n := s.Name
				if s.Map == -1 {
					if s.ID == m.Server {
						n += "* - " + n
					} else {
						n += "! - " + n
					}
				} else {
					n = "    " + n
				}
				opts = append(opts, xform.Option{
					Label:    n,
					Value:    strconv.Itoa(i),
					Selected: s.ID == m.Server,
				})
			}
			sel := xform.SelectBox("server", opts...)
			submit := xform.InputSubmit("Set")
			submit.AddEventListener("click", false, func(e dom.Event) {
				if name.Value == "" {
					return
				}
				mID, err := strconv.Atoi(sel.Value)
				if err != nil || mID < -1 || mID >= len(servers) {
					return
				}
				submit.Disabled = true
				e.PreventDefault()
				if mID >= 0 {
					s := servers[mID]
					mID = s.ID
				}
				go func() {
					err = RPC.SetServerMap(mID, m.ID)
					if err != nil {
						xjs.Alert("Error setting map server: %s", err)
						return
					}
					m.Name = name.Value
					err = RPC.SetMap(m)
					if err != nil {
						xjs.Alert("Error setting map data: %s", err)
						return
					}
					span := xdom.Span()
					span.Style().Set("color", "#f00")
					c.AppendChild(xjs.SetInnerText(span, "Saved!"))
					time.Sleep(5 * time.Second)
					c.RemoveChild(span)
					submit.Disabled = false
				}()
			})
			xjs.AppendChildren(c, xjs.AppendChildren(xdom.Form(),
				xform.Label("Server Name", "name"),
				name,
				xdom.Br(),
				xform.Label("Server Name", "server"),
				sel,
				xdom.Br(),
				submit,
			))
		}()
	}
}

func mapProperties(m data.Map) func(dom.Element) {
	return func(c dom.Element) {
		go editProperties(c, "Map", m.ID, RPC.MapProperties, RPC.SetMapProperties)
	}
}
