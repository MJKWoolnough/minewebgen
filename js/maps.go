package main

import (
	"errors"
	"strconv"
	"time"

	"github.com/MJKWoolnough/gopherjs/overlay"
	"github.com/MJKWoolnough/gopherjs/style"
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xform"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/MJKWoolnough/minewebgen/internal/data"
	"honnef.co/go/js/dom"
)

type Map struct {
	data.Map
	row    dom.Node
	name   *dom.HTMLTableCellElement
	status *dom.HTMLTableCellElement
}

func MapsTab() func(dom.Element) {
	forceUpdate := make(chan struct{})
	nm := xdom.Button()
	nm.AddEventListener("click", false, func(dom.Event) {
		d := xdom.Div()
		o := overlay.New(d)
		o.OnClose(func() {
			go func() {
				forceUpdate <- struct{}{}
			}()
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
	noneTd := xdom.Td()
	noneTd.ColSpan = 2
	none := xjs.AppendChildren(xdom.Tr(), xjs.SetInnerText(noneTd, "No Maps Found"))
	mapList := xjs.AppendChildren(xdom.Table(),
		xjs.AppendChildren(xdom.Thead(), xjs.AppendChildren(xdom.Tr(),
			xjs.SetInnerText(xdom.Th(), "Map Name"),
			xjs.SetInnerText(xdom.Th(), "Status"),
		)),
		none,
	)
	nodes := xjs.AppendChildren(xdom.Div(),
		xjs.SetInnerText(xdom.H2(), "Maps"),
		xjs.SetInnerText(nm, "New Map"),
		mapList,
	)

	maps := make(map[int]*Map)
	return func(c dom.Element) {
		c.AppendChild(nodes)
		updateStop := make(chan struct{})
		registerUpdateStopper(c, updateStop)
		for {
			mps, err := RPC.MapList()
			if err != nil {
				xjs.Alert("Error getting map list: %s", err)
				return
			}

			if none.ParentNode() != nil {
				mapList.RemoveChild(none)
			}

			for _, m := range maps {
				m.ID = -1
			}

			for _, m := range mps {
				om, ok := maps[m.ID]
				if ok {
					om.Map = m
				} else {
					name := xdom.Td()
					status := xdom.Td()
					om = &Map{
						Map: m,
						row: xjs.AppendChildren(xdom.Tr(),
							name,
							status,
						),
						name:   name,
						status: status,
					}
					maps[m.ID] = om
					mapList.AppendChild(om.row)
					name.AddEventListener("click", false, func() func(dom.Event) {
						m := om
						return func(dom.Event) {
							o := overlay.New(xjs.AppendChildren(xdom.Div(), tabs.New([]tabs.Tab{
								{"General", mapGeneral(m.Map)},
								{"Properties", mapProperties(m.Map)},
								{"Misc.", mapMisc(m.Map)},
							})))
							o.OnClose(func() {
								go func() {
									forceUpdate <- struct{}{}
								}()
							})
							xjs.Body().AppendChild(o)
						}
					}())
				}
				switch om.Server {
				case -2:
					xjs.SetInnerText(om.status, "Busy")
					om.status.Style().SetProperty("color", "#f00", "")
				case -1:
					xjs.SetInnerText(om.status, "Unassigned")
					om.status.Style().SetProperty("color", "#00f", "")
				default:
					serv, err := RPC.Server(om.Server)
					if err == nil {
						xjs.SetInnerText(om.status, "Assigned")
					} else {
						xjs.SetInnerText(om.status, "Assigned - "+serv.Name)
					}
					om.status.Style().SetProperty("color", "#000", "")
				}
				xjs.SetInnerText(om.name, om.Name)
			}

			for id, m := range maps {
				if m.ID == -1 {
					delete(maps, id)
					mapList.RemoveChild(m.row)
				}
			}

			if len(maps) == 0 {
				mapList.AppendChild(none)
			}

			// Sleep until update
			if !updateSleep(forceUpdate, updateStop) {
				return
			}
		}
	}
}

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
					{"Misc.", mapMisc(m)},
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
				xjs.Alert("Error creating map: %s", err)
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
	gs.Required = true
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
				xjs.Alert("Error creating map: %s", err)
				return
			}
			o.Close()

		}()
	})
	return func(c dom.Element) {
		c.AppendChild(d)
	}
}

func rangeWatch(r *dom.HTMLInputElement) dom.Element {
	s := xdom.Span()
	xjs.SetInnerText(s, r.Value)
	r.AddEventListener("input", false, func(dom.Event) {
		xjs.SetInnerText(s, r.Value)
	})
	return s
}

func enabler(c, r *dom.HTMLInputElement) {
	c.AddEventListener("click", false, func(dom.Event) {
		r.Disabled = !c.Checked
	})
}

var biomes = []xform.Option{
	{"All", "-1", false},
	{"Ocean", "0", false},
	{"Plains", "1", false},
	{"Desert", "2", false},
	{"Extreme Hills", "3", false},
	{"Forest", "4", false},
	{"Taiga", "5", false},
	{"Swampland", "6", false},
	{"River", "7", false},
	{"Frozen Ocean", "10", false},
	{"Frozen River", "11", false},
	{"Ice Plains", "12", false},
	{"Ice Mountains", "13", false},
	{"Mushroom Island", "14", false},
	{"Mushroom Island Shore", "15", false},
	{"Beach", "16", false},
	{"Desert Hills", "17", false},
	{"Forest Hills", "18", false},
	{"Taiga Hills", "19", false},
	{"Extreme Hills Edge", "20", false},
	{"Jungle", "21", false},
	{"Jungle Hills", "22", false},
	{"Jungle Edge", "23", false},
	{"Deep Ocean", "24", false},
	{"Stone Beach", "25", false},
	{"Cold Beach", "26", false},
	{"Birch Forest", "27", false},
	{"Birch Forest Hills", "28", false},
	{"Roofed Forest", "29", false},
	{"Cold Taiga", "30", false},
	{"Cold Taiga Hills", "31", false},
	{"Mega Taiga", "32", false},
	{"Mega Taiga Hills", "33", false},
	{"Extreme Hills+", "34", false},
	{"Savanna", "35", false},
	{"Savanna Plateau", "36", false},
	{"Mesa", "37", false},
	{"Mesa Plateau F", "38", false},
	{"Mesa Plateau", "39", false},
}

func init() {
	style.Add(`.brClear br {
	clear : left;
}
`)
}

func createCustomisedMap(o *overlay.Overlay, dataParser func() (data.DefaultMap, error)) func(dom.Element) {
	d := xdom.Div()
	d.Class().SetString("brClear")
	seaLevel := xform.InputRange("sea", 0, 255, 1, 63)
	caves := xform.InputCheckbox("caves", true)
	strongholds := xform.InputCheckbox("strongholds", true)
	villages := xform.InputCheckbox("villages", true)
	mineshafts := xform.InputCheckbox("mineshafts", true)
	temples := xform.InputCheckbox("templaes", true)
	oceanMonuments := xform.InputCheckbox("oceanMonuments", true)
	ravines := xform.InputCheckbox("ravines", true)
	dungeons := xform.InputCheckbox("dungeons", true)
	dungeonCount := xform.InputRange("dungeonCount", 1, 100, 1, 7)
	waterLakes := xform.InputCheckbox("waterLakes", true)
	waterLakeRarity := xform.InputRange("waterLakeRarity", 1, 100, 1, 4)
	lavaLakes := xform.InputCheckbox("lavaLakes", true)
	lavaLakeRarity := xform.InputRange("lavaLakeRarity", 1, 100, 1, 80)
	biome := xform.SelectBox("biomes", biomes...)
	biomeSize := xform.InputRange("biomeSize", 1, 8, 1, 4)
	riverSize := xform.InputRange("riverSize", 1, 5, 1, 4)

	dirtSpawnSize := xform.InputRange("dirtSpawnSize", 1, 50, 1, 33)
	dirtSpawnTries := xform.InputRange("dirtSpawnTries", 0, 40, 1, 10)
	dirtMinHeight := xform.InputRange("dirtMinHeight", 0, 255, 1, 0)
	dirtMaxHeight := xform.InputRange("dirtMaxHeight", 0, 255, 1, 256)

	gravelSpawnSize := xform.InputRange("gravelSpawnSize", 1, 50, 1, 33)
	gravelSpawnTries := xform.InputRange("gravelSpawnTries", 0, 40, 1, 8)
	gravelMinHeight := xform.InputRange("gravelMinHeight", 0, 255, 1, 0)
	gravelMaxHeight := xform.InputRange("gravelMaxHeight", 0, 255, 1, 256)

	graniteSpawnSize := xform.InputRange("graniteSpawnSize", 1, 50, 1, 33)
	graniteSpawnTries := xform.InputRange("graniteSpawnTries", 0, 40, 1, 10)
	graniteMinHeight := xform.InputRange("graniteMinHeight", 0, 255, 1, 0)
	graniteMaxHeight := xform.InputRange("graniteMaxHeight", 0, 255, 1, 80)

	dioriteSpawnSize := xform.InputRange("dioriteSpawnSize", 1, 50, 1, 33)
	dioriteSpawnTries := xform.InputRange("dioriteSpawnTries", 0, 40, 1, 10)
	dioriteMinHeight := xform.InputRange("dioriteMinHeight", 0, 255, 1, 0)
	dioriteMaxHeight := xform.InputRange("dioriteMaxHeight", 0, 255, 1, 80)

	andesiteSpawnSize := xform.InputRange("andesiteSpawnSize", 1, 50, 1, 33)
	andesiteSpawnTries := xform.InputRange("andesiteSpawnTries", 0, 40, 1, 10)
	andesiteMinHeight := xform.InputRange("andesiteMinHeight", 0, 255, 1, 0)
	andesiteMaxHeight := xform.InputRange("andesiteMaxHeight", 0, 255, 1, 80)

	coalOreSpawnSize := xform.InputRange("coalOreSpawnSize", 1, 50, 1, 17)
	coalOreSpawnTries := xform.InputRange("coalOreSpawnTries", 0, 40, 1, 20)
	coalOreMinHeight := xform.InputRange("coalOreMinHeight", 0, 255, 1, 0)
	coalOreMaxHeight := xform.InputRange("coalOreMaxHeight", 0, 255, 1, 128)

	ironOreSpawnSize := xform.InputRange("ironOreSpawnSize", 1, 50, 1, 9)
	ironOreSpawnTries := xform.InputRange("ironOreSpawnTries", 0, 40, 1, 20)
	ironOreMinHeight := xform.InputRange("ironOreMinHeight", 0, 255, 1, 0)
	ironOreMaxHeight := xform.InputRange("ironOreMaxHeight", 0, 255, 1, 64)

	goldOreSpawnSize := xform.InputRange("goldOreSpawnSize", 1, 50, 1, 9)
	goldOreSpawnTries := xform.InputRange("goldOreSpawnTries", 0, 40, 1, 2)
	goldOreMinHeight := xform.InputRange("goldOreMinHeight", 0, 255, 1, 0)
	goldOreMaxHeight := xform.InputRange("goldOreMaxHeight", 0, 255, 1, 32)

	redstoneOreSpawnSize := xform.InputRange("redstoneOreSpawnSize", 1, 50, 1, 8)
	redstoneOreSpawnTries := xform.InputRange("redstoneOreSpawnTries", 0, 40, 1, 8)
	redstoneOreMinHeight := xform.InputRange("redstoneOreMinHeight", 0, 255, 1, 0)
	redstoneOreMaxHeight := xform.InputRange("redstoneOreMaxHeight", 0, 255, 1, 16)

	diamondOreSpawnSize := xform.InputRange("diamondOreSpawnSize", 1, 50, 1, 8)
	diamondOreSpawnTries := xform.InputRange("diamondOreSpawnTries", 0, 40, 1, 1)
	diamondOreMinHeight := xform.InputRange("diamondOreMinHeight", 0, 255, 1, 0)
	diamondOreMaxHeight := xform.InputRange("diamondOreMaxHeight", 0, 255, 1, 16)

	lapisLazuliOreSpawnSize := xform.InputRange("lapisLazuliOreSpawnSize", 1, 50, 1, 7)
	lapisLazuliOreSpawnTries := xform.InputRange("lapisLazuliOreSpawnTries", 0, 40, 1, 1)
	lapisLazuliOreCentreHeight := xform.InputRange("lapisLazuliOreCentreHeight", 0, 255, 1, 16)
	lapisLazuliOreSpreadHeight := xform.InputRange("lapisLazuliOreSpreadHeight", 0, 255, 1, 16)

	mainNoiseScaleX := xform.InputRange("mainNoiseScaleX", 1, 5000, -1, 80)
	mainNoiseScaleY := xform.InputRange("mainNoiseScaleY", 1, 5000, -1, 160)
	mainNoiseScaleZ := xform.InputRange("mainNoiseScaleZ", 1, 5000, -1, 80)
	depthNoiseScaleX := xform.InputRange("depthNoiseScaleX", 1, 2000, -1, 200)
	depthNoiseScaleZ := xform.InputRange("depthNoiseScaleZ", 1, 2000, -1, 200)
	depthNoiseExponent := xform.InputRange("depthNoiseExponent", 0.01, 20, -1, 0.5)
	depthBaseSize := xform.InputRange("depthBaseSize", 1, 25, -1, 8.5)
	coordinateScale := xform.InputRange("coordinateScale", 1, 6000, -1, 684.412)
	heightScale := xform.InputRange("heightScale", 1, 6000, -1, 684.412)
	heightStretch := xform.InputRange("heightStretch", 0.01, 50, -1, 12)
	upperLimitScale := xform.InputRange("upperLimitScale", 1, 5000, -1, 512)
	lowerLimitScale := xform.InputRange("lowerLimitScale", 1, 5000, -1, 512)
	biomeDepthWeight := xform.InputRange("biomeDepthWeight", 1, 20, -1, 1)
	biomeDepthOffset := xform.InputRange("biomeDepthOffset", 0, 20, -1, 0)
	biomeScaleWeight := xform.InputRange("biomeScaleWeight", 1, 20, -1, 1)
	biomeScaleOffset := xform.InputRange("biomeScaleOffset", 0, 20, -1, 0)

	submit := xform.InputSubmit("Create Map")
	submit.AddEventListener("click", false, func(e dom.Event) {
		d, err := dataParser()
		if err != nil {
			xjs.Alert("Error parsing values: %s", err)
			return
		}
		e.PreventDefault()
		cd := data.CustomMap{
			DefaultMap: d,
		}
		cd.GeneratorSettings.SeaLevel = uint8(seaLevel.ValueAsNumber)
		cd.GeneratorSettings.Caves = caves.Checked
		cd.GeneratorSettings.Strongholds = strongholds.Checked
		cd.GeneratorSettings.Villages = villages.Checked
		cd.GeneratorSettings.Mineshafts = mineshafts.Checked
		cd.GeneratorSettings.Temples = temples.Checked
		cd.GeneratorSettings.OceanMonuments = oceanMonuments.Checked
		cd.GeneratorSettings.Ravines = ravines.Checked
		cd.GeneratorSettings.Dungeons = dungeons.Checked
		cd.GeneratorSettings.DungeonChance = uint8(dungeonCount.ValueAsNumber)
		cd.GeneratorSettings.WaterLake = waterLakes.Checked
		cd.GeneratorSettings.WaterLakeChance = uint8(waterLakeRarity.ValueAsNumber)
		cd.GeneratorSettings.LaveLake = lavaLakes.Checked
		cd.GeneratorSettings.LavaLakeChance = uint8(lavaLakeRarity.ValueAsNumber)
		b, err := strconv.Atoi(biome.Value)
		if err != nil {
			b = -1
		}
		cd.GeneratorSettings.Biome = int16(b)

		cd.GeneratorSettings.BiomeSize = uint8(biomeSize.ValueAsNumber)
		cd.GeneratorSettings.RiverSize = uint8(riverSize.ValueAsNumber)

		cd.GeneratorSettings.DirtSize = uint8(dirtSpawnSize.ValueAsNumber)
		cd.GeneratorSettings.DirtTries = uint8(dirtSpawnTries.ValueAsNumber)
		cd.GeneratorSettings.DirtMinHeight = uint8(dirtMinHeight.ValueAsNumber)
		cd.GeneratorSettings.DirtMaxHeight = uint8(dirtMaxHeight.ValueAsNumber)

		cd.GeneratorSettings.GravelSize = uint8(gravelSpawnSize.ValueAsNumber)
		cd.GeneratorSettings.GravelTries = uint8(gravelSpawnTries.ValueAsNumber)
		cd.GeneratorSettings.GravelMinHeight = uint8(gravelMinHeight.ValueAsNumber)
		cd.GeneratorSettings.GravelMaxHeight = uint8(gravelMaxHeight.ValueAsNumber)

		cd.GeneratorSettings.GraniteSize = uint8(graniteSpawnSize.ValueAsNumber)
		cd.GeneratorSettings.GraniteTries = uint8(graniteSpawnTries.ValueAsNumber)
		cd.GeneratorSettings.GraniteMinHeight = uint8(graniteMinHeight.ValueAsNumber)
		cd.GeneratorSettings.GraniteMaxHeight = uint8(graniteMaxHeight.ValueAsNumber)

		cd.GeneratorSettings.DiortiteSize = uint8(dioriteSpawnSize.ValueAsNumber)
		cd.GeneratorSettings.DiortiteTries = uint8(dioriteSpawnTries.ValueAsNumber)
		cd.GeneratorSettings.DiortiteMinHeight = uint8(dioriteMinHeight.ValueAsNumber)
		cd.GeneratorSettings.DiortiteMaxHeight = uint8(dioriteMaxHeight.ValueAsNumber)

		cd.GeneratorSettings.AndesiteSize = uint8(andesiteSpawnSize.ValueAsNumber)
		cd.GeneratorSettings.AndesiteTries = uint8(andesiteSpawnTries.ValueAsNumber)
		cd.GeneratorSettings.AndesiteMinHeight = uint8(andesiteMinHeight.ValueAsNumber)
		cd.GeneratorSettings.AndesiteMaxHeight = uint8(andesiteMaxHeight.ValueAsNumber)

		cd.GeneratorSettings.CoalSize = uint8(coalOreSpawnSize.ValueAsNumber)
		cd.GeneratorSettings.CoalTries = uint8(coalOreSpawnTries.ValueAsNumber)
		cd.GeneratorSettings.CoalMinHeight = uint8(coalOreMinHeight.ValueAsNumber)
		cd.GeneratorSettings.CoalMaxHeight = uint8(coalOreMaxHeight.ValueAsNumber)

		cd.GeneratorSettings.IronSize = uint8(ironOreSpawnSize.ValueAsNumber)
		cd.GeneratorSettings.IronTries = uint8(ironOreSpawnTries.ValueAsNumber)
		cd.GeneratorSettings.IronMinHeight = uint8(ironOreMinHeight.ValueAsNumber)
		cd.GeneratorSettings.IronMaxHeight = uint8(ironOreMaxHeight.ValueAsNumber)

		cd.GeneratorSettings.GoldSize = uint8(goldOreSpawnSize.ValueAsNumber)
		cd.GeneratorSettings.GoldTries = uint8(goldOreSpawnTries.ValueAsNumber)
		cd.GeneratorSettings.GoldMinHeight = uint8(goldOreMinHeight.ValueAsNumber)
		cd.GeneratorSettings.GoldMaxHeight = uint8(goldOreMaxHeight.ValueAsNumber)

		cd.GeneratorSettings.RedstoneSize = uint8(redstoneOreSpawnSize.ValueAsNumber)
		cd.GeneratorSettings.RedstoneTries = uint8(redstoneOreSpawnTries.ValueAsNumber)
		cd.GeneratorSettings.RedstoneMinHeight = uint8(redstoneOreMinHeight.ValueAsNumber)
		cd.GeneratorSettings.RedstoneMaxHeight = uint8(redstoneOreMaxHeight.ValueAsNumber)

		cd.GeneratorSettings.DiamondSize = uint8(diamondOreSpawnSize.ValueAsNumber)
		cd.GeneratorSettings.DiamondTries = uint8(diamondOreSpawnTries.ValueAsNumber)
		cd.GeneratorSettings.DiamondMinHeight = uint8(diamondOreMinHeight.ValueAsNumber)
		cd.GeneratorSettings.DiamondMaxHeight = uint8(diamondOreMaxHeight.ValueAsNumber)

		cd.GeneratorSettings.LapisSize = uint8(lapisLazuliOreSpawnSize.ValueAsNumber)
		cd.GeneratorSettings.LapisTries = uint8(lapisLazuliOreSpawnTries.ValueAsNumber)
		cd.GeneratorSettings.LapisCenterHeight = uint8(lapisLazuliOreCentreHeight.ValueAsNumber)
		cd.GeneratorSettings.LapisSpread = uint8(lapisLazuliOreSpreadHeight.ValueAsNumber)

		cd.GeneratorSettings.MainNoiseScaleX = mainNoiseScaleX.ValueAsNumber
		cd.GeneratorSettings.MainNoiseScaleY = mainNoiseScaleY.ValueAsNumber
		cd.GeneratorSettings.MainNoiseScaleZ = mainNoiseScaleZ.ValueAsNumber
		cd.GeneratorSettings.DepthNoiseScaleX = depthNoiseScaleX.ValueAsNumber
		cd.GeneratorSettings.DepthNoiseScaleZ = depthNoiseScaleZ.ValueAsNumber
		cd.GeneratorSettings.DepthNoiseScaleExponent = depthNoiseExponent.ValueAsNumber
		cd.GeneratorSettings.BaseSize = depthBaseSize.ValueAsNumber
		cd.GeneratorSettings.CoordinateScale = coordinateScale.ValueAsNumber
		cd.GeneratorSettings.HeightScale = heightScale.ValueAsNumber
		cd.GeneratorSettings.HeightStretch = heightStretch.ValueAsNumber
		cd.GeneratorSettings.UpperLimitScale = upperLimitScale.ValueAsNumber
		cd.GeneratorSettings.LowerLimitScale = lowerLimitScale.ValueAsNumber
		cd.GeneratorSettings.BiomeDepthWeight = biomeDepthWeight.ValueAsNumber
		cd.GeneratorSettings.BiomeDepthOffset = biomeDepthOffset.ValueAsNumber
		cd.GeneratorSettings.BiomeScaleWeight = biomeScaleWeight.ValueAsNumber
		cd.GeneratorSettings.BiomeScaleOffset = biomeScaleOffset.ValueAsNumber
		go func() {
			err = RPC.CreateCustomMap(cd)
			if err != nil {
				xjs.Alert("Error creating map: %s", err)
				return
			}
			o.Close()
		}()
	})

	enabler(dungeons, dungeonCount)
	enabler(waterLakes, waterLakeRarity)
	enabler(lavaLakes, lavaLakeRarity)

	xjs.AppendChildren(d,
		xform.Label("Sea Level", "sea"), seaLevel, rangeWatch(seaLevel), xdom.Br(),
		xform.Label("Caves", "caves"), caves, xdom.Br(),
		xform.Label("Strongholds", "strongholds"), strongholds, xdom.Br(),
		xform.Label("Villages", "villages"), villages, xdom.Br(),
		xform.Label("Mineshafts", "mineshafts"), mineshafts, xdom.Br(),
		xform.Label("Temples", "temples"), temples, xdom.Br(),
		xform.Label("Ocean Monuments", "oceanMonuments"), oceanMonuments, xdom.Br(),
		xform.Label("Ravines", "ravines"), ravines, xdom.Br(),
		xform.Label("Dungeons", "dungeons"), dungeons, xdom.Br(),
		xform.Label("Dungeon Count", "dungeonCount"), dungeonCount, rangeWatch(dungeonCount), xdom.Br(),
		xform.Label("Water Lakes", "waterLakes"), waterLakes, xdom.Br(),
		xform.Label("Water Lake Rarity", "waterLakeRarity"), waterLakeRarity, rangeWatch(waterLakeRarity), xdom.Br(),
		xform.Label("Lava Lakes", "lavaLakes"), lavaLakes, xdom.Br(),
		xform.Label("Lava Lake Rarity", "lavaLakeRarity"), lavaLakeRarity, rangeWatch(lavaLakeRarity), xdom.Br(),
		xform.Label("Biomes", "biomes"), biome, xdom.Br(),
		xform.Label("Biome Size", "biomeSize"), biomeSize, xdom.Br(),
		xform.Label("River Size", "riverSize"), riverSize, xdom.Br(),
		xdom.Br(),
		xjs.AppendChildren(xdom.Table(),
			xjs.AppendChildren(xdom.Tr(),
				xjs.AppendChildren(xdom.Td(),
					xform.Label("Dirt Spawn Size", "dirtSpawnSize"), dirtSpawnSize, rangeWatch(dirtSpawnSize), xdom.Br(),
					xform.Label("Dirt Spawn Tries", "dirtSpawnTries"), dirtSpawnTries, rangeWatch(dirtSpawnTries), xdom.Br(),
					xform.Label("Dirt Spawn Min Height", "dirtMinHeight"), dirtMinHeight, rangeWatch(dirtMinHeight), xdom.Br(),
					xform.Label("Dirt Spawn Max Height", "dirtMaxHeight"), dirtMaxHeight, rangeWatch(dirtMaxHeight), xdom.Br(),
				),
				xjs.AppendChildren(xdom.Td(),
					xform.Label("Gravel Spawn Size", "gravelSpawnSize"), gravelSpawnSize, rangeWatch(gravelSpawnSize), xdom.Br(),
					xform.Label("Gravel Spawn Tries", "gravelSpawnTries"), gravelSpawnTries, rangeWatch(gravelSpawnTries), xdom.Br(),
					xform.Label("Gravel Spawn Min Height", "gravelMinHeight"), gravelMinHeight, rangeWatch(gravelMinHeight), xdom.Br(),
					xform.Label("Gravel Spawn Max Height", "gravelMaxHeight"), gravelMaxHeight, rangeWatch(gravelMaxHeight), xdom.Br(),
				),
			),
			xjs.AppendChildren(xdom.Tr(),
				xjs.AppendChildren(xdom.Td(),
					xform.Label("Granite Spawn Size", "graniteSpawnSize"), graniteSpawnSize, rangeWatch(graniteSpawnSize), xdom.Br(),
					xform.Label("Granite Spawn Tries", "graniteSpawnTries"), graniteSpawnTries, rangeWatch(graniteSpawnTries), xdom.Br(),
					xform.Label("Granite Spawn Min Height", "graniteMinHeight"), graniteMinHeight, rangeWatch(graniteMinHeight), xdom.Br(),
					xform.Label("Granite Spawn Max Height", "graniteMaxHeight"), graniteMaxHeight, rangeWatch(graniteMaxHeight), xdom.Br(),
				),
				xjs.AppendChildren(xdom.Td(),
					xform.Label("Diorite Spawn Size", "dioriteSpawnSize"), dioriteSpawnSize, rangeWatch(dioriteSpawnSize), xdom.Br(),
					xform.Label("Diorite Spawn Tries", "dioriteSpawnTries"), dioriteSpawnTries, rangeWatch(dioriteSpawnTries), xdom.Br(),
					xform.Label("Diorite Spawn Min Height", "dioriteMinHeight"), dioriteMinHeight, rangeWatch(dioriteMinHeight), xdom.Br(),
					xform.Label("Diorite Spawn Max Height", "dioriteMaxHeight"), dioriteMaxHeight, rangeWatch(dioriteMaxHeight), xdom.Br(),
				),
			),
			xjs.AppendChildren(xdom.Tr(),
				xjs.AppendChildren(xdom.Td(),
					xform.Label("Andesite Spawn Size", "andesiteSpawnSize"), andesiteSpawnSize, rangeWatch(andesiteSpawnSize), xdom.Br(),
					xform.Label("Andesite Spawn Tries", "andesiteSpawnTries"), andesiteSpawnTries, rangeWatch(andesiteSpawnTries), xdom.Br(),
					xform.Label("Andesite Spawn Min Height", "andesiteMinHeight"), andesiteMinHeight, rangeWatch(andesiteMinHeight), xdom.Br(),
					xform.Label("Andesite Spawn Max Height", "andesiteMaxHeight"), andesiteMaxHeight, rangeWatch(andesiteMaxHeight), xdom.Br(),
				),
				xjs.AppendChildren(xdom.Td(),
					xform.Label("Coal Ore Spawn Size", "coalOreSpawnSize"), coalOreSpawnSize, rangeWatch(coalOreSpawnSize), xdom.Br(),
					xform.Label("Coal Ore Spawn Tries", "coalOreSpawnTries"), coalOreSpawnTries, rangeWatch(coalOreSpawnTries), xdom.Br(),
					xform.Label("Coal Ore Spawn Min Height", "coalOreMinHeight"), coalOreMinHeight, rangeWatch(coalOreMinHeight), xdom.Br(),
					xform.Label("Coal Ore Spawn Max Height", "coalOreMaxHeight"), coalOreMaxHeight, rangeWatch(coalOreMaxHeight), xdom.Br(),
				),
			),
			xjs.AppendChildren(xdom.Tr(),
				xjs.AppendChildren(xdom.Td(),
					xform.Label("Iron Ore Spawn Size", "ironOreSpawnSize"), ironOreSpawnSize, rangeWatch(ironOreSpawnSize), xdom.Br(),
					xform.Label("Iron Ore Spawn Tries", "ironOreSpawnTries"), ironOreSpawnTries, rangeWatch(ironOreSpawnTries), xdom.Br(),
					xform.Label("Iron Ore Spawn Min Height", "ironOreMinHeight"), ironOreMinHeight, rangeWatch(ironOreMinHeight), xdom.Br(),
					xform.Label("Iron Ore Spawn Max Height", "ironOreMaxHeight"), ironOreMaxHeight, rangeWatch(ironOreMaxHeight), xdom.Br(),
				),
				xjs.AppendChildren(xdom.Td(),
					xform.Label("Gold Ore Spawn Size", "goldOreSpawnSize"), goldOreSpawnSize, rangeWatch(goldOreSpawnSize), xdom.Br(),
					xform.Label("Gold Ore Spawn Tries", "goldOreSpawnTries"), goldOreSpawnTries, rangeWatch(goldOreSpawnTries), xdom.Br(),
					xform.Label("Gold Ore Spawn Min Height", "goldOreMinHeight"), goldOreMinHeight, rangeWatch(goldOreMinHeight), xdom.Br(),
					xform.Label("Gold Ore Spawn Max Height", "goldOreMaxHeight"), goldOreMaxHeight, rangeWatch(goldOreMaxHeight), xdom.Br(),
				),
			),
			xjs.AppendChildren(xdom.Tr(),
				xjs.AppendChildren(xdom.Td(),
					xform.Label("Redstone Ore Spawn Size", "redstoneOreSpawnSize"), redstoneOreSpawnSize, rangeWatch(redstoneOreSpawnSize), xdom.Br(),
					xform.Label("Redstone Ore Spawn Tries", "redstoneOreSpawnTries"), redstoneOreSpawnTries, rangeWatch(redstoneOreSpawnTries), xdom.Br(),
					xform.Label("Redstone Ore Spawn Min Height", "redstoneOreMinHeight"), redstoneOreMinHeight, rangeWatch(redstoneOreMinHeight), xdom.Br(),
					xform.Label("Redstone Ore Spawn Max Height", "redstoneOreMaxHeight"), redstoneOreMaxHeight, rangeWatch(redstoneOreMaxHeight), xdom.Br(),
				),
				xjs.AppendChildren(xdom.Td(),
					xform.Label("Diamond Ore Spawn Size", "diamondOreSpawnSize"), diamondOreSpawnSize, rangeWatch(diamondOreSpawnSize), xdom.Br(),
					xform.Label("Diamond Ore Spawn Tries", "diamondOreSpawnTries"), diamondOreSpawnTries, rangeWatch(diamondOreSpawnTries), xdom.Br(),
					xform.Label("Diamond Ore Spawn Min Height", "diamondOreMinHeight"), diamondOreMinHeight, rangeWatch(diamondOreMinHeight), xdom.Br(),
					xform.Label("Diamond Ore Spawn Max Height", "diamondOreMaxHeight"), diamondOreMaxHeight, rangeWatch(diamondOreMaxHeight), xdom.Br(),
				),
			),
			xjs.AppendChildren(xdom.Tr(),
				xjs.AppendChildren(xdom.Td(),
					xform.Label("Lapis Lazuli Ore Spawn Size", "lapisLazuliOreSpawnSize"), lapisLazuliOreSpawnSize, rangeWatch(lapisLazuliOreSpawnSize), xdom.Br(),
					xform.Label("Lapis Lazuli Ore Spawn Tries", "lapisLazuliOreSpawnTries"), lapisLazuliOreSpawnTries, rangeWatch(lapisLazuliOreSpawnTries), xdom.Br(),
					xform.Label("Lapis Lazuli Ore Centre Height", "lapisLazuliOreCentreHeight"), lapisLazuliOreCentreHeight, rangeWatch(lapisLazuliOreCentreHeight), xdom.Br(),
					xform.Label("Lapis Lazuli Ore Spread Height", "lapisLazuliOreSpreadHeight"), lapisLazuliOreSpreadHeight, rangeWatch(lapisLazuliOreSpreadHeight), xdom.Br(),
				),
			),
		),
		xform.Label("Main Noise Scale X", "mainNoiseScaleX"), mainNoiseScaleX, rangeWatch(mainNoiseScaleX), xdom.Br(),
		xform.Label("Main Noise Scale Y", "mainNoiseScaleY"), mainNoiseScaleY, rangeWatch(mainNoiseScaleY), xdom.Br(),
		xform.Label("Main Noise Scale Z", "mainNoiseScaleZ"), mainNoiseScaleZ, rangeWatch(mainNoiseScaleZ), xdom.Br(),
		xform.Label("Depth Noise Scale X", "depthNoiseScaleX"), depthNoiseScaleX, rangeWatch(depthNoiseScaleX), xdom.Br(),
		xform.Label("Depth Noise Scale Z", "depthNoiseScaleZ"), depthNoiseScaleZ, rangeWatch(depthNoiseScaleZ), xdom.Br(),
		xform.Label("Depth Noise Exponent", "depthNoiseExponent"), depthNoiseExponent, rangeWatch(depthNoiseExponent), xdom.Br(),
		xform.Label("Depth Base Size", "depthBaseSize"), depthBaseSize, rangeWatch(depthBaseSize), xdom.Br(),
		xform.Label("Coordinate Scale", "coordinateScale"), coordinateScale, rangeWatch(coordinateScale), xdom.Br(),
		xform.Label("Height Scale", "heightScale"), heightScale, rangeWatch(heightScale), xdom.Br(),
		xform.Label("Height Stretch", "heightStretch"), heightStretch, rangeWatch(heightStretch), xdom.Br(),
		xform.Label("Upper Limit Scale", "upperLimitScale"), upperLimitScale, rangeWatch(upperLimitScale), xdom.Br(),
		xform.Label("Lower Limit Scale", "lowerLimitScale"), lowerLimitScale, rangeWatch(lowerLimitScale), xdom.Br(),
		xform.Label("Biome Depth Weight", "biomeDepthWeight"), biomeDepthWeight, rangeWatch(biomeDepthWeight), xdom.Br(),
		xform.Label("Biome Depth Offset", "biomeDepthOffset"), biomeDepthOffset, rangeWatch(biomeDepthOffset), xdom.Br(),
		xform.Label("Biome Scale Weight", "biomeDepthWeight"), biomeScaleWeight, rangeWatch(biomeScaleWeight), xdom.Br(),
		xform.Label("Biome Scale Offset", "biomeDepthOffset"), biomeScaleOffset, rangeWatch(biomeScaleOffset), xdom.Br(),
		submit,
	)

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
			var cs data.Server
			for i, s := range servers {
				n := s.Name
				if s.Map != -1 {
					if s.ID == m.Server {
						cs = s
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
				if cs.State != data.StateStopped {
					xjs.Alert("Cannot modify these settings while connected server is running.")
					return
				}
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

func mapMisc(m data.Map) func(dom.Element) {
	return func(c dom.Element) {
		// Download Map
		// Delete Map
	}
}
