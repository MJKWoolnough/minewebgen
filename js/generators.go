package main

import (
	"fmt"

	"github.com/MJKWoolnough/gopherjs/overlay"
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/MJKWoolnough/minewebgen/internal/data"
	"honnef.co/go/js/dom"
)

func GeneratorsTab(c dom.Element) {
	go func() {
		xjs.RemoveChildren(c)
		gs, err := RPC.Generators()
		if err != nil {
			xjs.Alert("Error getting generator list: %s", err)
			return
		}
		ng := xdom.Button()
		xjs.SetInnerText(ng, "New Generator")
		ng.AddEventListener("click", false, func(dom.Event) {
			d := xdom.Div()
			o := overlay.New(d)
			o.OnClose(func() {
				GeneratorsTab(c)
			})
			d.AppendChild(transferFile("Map", "Upload/Download", 3, o))
		})
		table := xjs.AppendChildren(xdom.Table(), xjs.AppendChildren(xdom.Tr(), xjs.SetInnerText(xdom.Th(), "Generator")))
		if len(gs) == 0 {
			table.AppendChild(xjs.AppendChildren(xdom.Tr(), xjs.SetInnerText(xdom.Td(), "No Generators")))
		} else {
			for _, g := range gs {
				td := xdom.Td()
				td.AddEventListener("click", false, func(g string) func(dom.Event) {
					return func(dom.Event) {
						d := xdom.Div()
						o := overlay.New(d)
						o.OnClose(func() {
							GeneratorsTab(c)
						})
						d.AppendChild(tabs.New([]tabs.Tab{
							{"Profile", generatorProfile(g)},
							{"Misc", misc("generator", g, o, RPC.RemoveGenerator)},
						}))
						xjs.Body().AppendChild(o)
					}
				}(g))
				table.AppendChild(xjs.AppendChildren(xdom.Tr(), xjs.SetInnerText(td, g)))
			}
		}
		c.AppendChild(table)
	}()
}

func generatorProfile(name string) func(dom.Element) {
	var (
		loaded bool
		g      data.Generator
	)
	return func(c dom.Element) {
		if !loaded {
			var err error
			g, err = RPC.Generator(name)
			if err != nil {
				xjs.Alert("Error while getting generator settings: %s", err)
				return
			}
			loaded = true
		}
		tTable := xjs.AppendChildren(xdom.Table(), xjs.AppendChildren(xdom.Thead(), xjs.AppendChildren(xdom.Tr(),
			xjs.SetInnerText(xdom.Th(), "Colour"),
			xjs.SetInnerText(xdom.Th(), "Colour Code"),
			xjs.SetInnerText(xdom.Th(), "Name"),
		)))
		for _, t := range g.Terrain {
			colour := xdom.Td()
			cc := fmt.Sprintf("rgb(%d, %d, %d)", t.Colour.R, t.Colour.G, t.Colour.B)
			colour.Style().SetProperty("background-color", cc, "")
			colour.Style().SetProperty("border", "1px solid #000", "")
			tTable.AppendChild(xjs.AppendChildren(xdom.Tr(),
				colour,
				xjs.SetInnerText(xdom.Td(), cc),
				xjs.SetInnerText(xdom.Td(), t.Name),
			))
		}
		bTable := xjs.AppendChildren(xdom.Table(), xjs.AppendChildren(xdom.Thead(), xjs.AppendChildren(xdom.Tr(),
			xjs.SetInnerText(xdom.Th(), "Colour"),
			xjs.SetInnerText(xdom.Th(), "Colour Code"),
			xjs.SetInnerText(xdom.Th(), "Name"),
		)))
		for _, b := range g.Biomes {
			colour := xdom.Td()
			cc := fmt.Sprintf("rgb(%d, %d, %d)", b.Colour.R, b.Colour.G, b.Colour.B)
			colour.Style().SetProperty("background-color", cc, "")
			colour.Style().SetProperty("border", "1px solid #000", "")
			bTable.AppendChild(xjs.AppendChildren(xdom.Tr(),
				colour,
				xjs.SetInnerText(xdom.Td(), cc),
				xjs.SetInnerText(xdom.Td(), b.Name),
			))
		}
		pTable := xjs.AppendChildren(xdom.Table(), xjs.AppendChildren(xdom.Thead(), xjs.AppendChildren(xdom.Tr(),
			xjs.SetInnerText(xdom.Th(), "Colour"),
			xjs.SetInnerText(xdom.Th(), "Colour Code"),
			xjs.SetInnerText(xdom.Th(), "Name"),
		)))
		for _, p := range g.Plants {
			colour := xdom.Td()
			cc := fmt.Sprintf("rgb(%d, %d, %d)", p.Colour.R, p.Colour.G, p.Colour.B)
			colour.Style().SetProperty("background-color", cc, "")
			colour.Style().SetProperty("border", "1px solid #000", "")
			pTable.AppendChild(xjs.AppendChildren(xdom.Tr(),
				colour,
				xjs.SetInnerText(xdom.Td(), cc),
				xjs.SetInnerText(xdom.Td(), p.Name),
			))
		}
		xjs.AppendChildren(c,
			xjs.SetInnerText(xdom.H2(), "Terrain"),
			tTable,
			xjs.SetInnerText(xdom.H2(), "Biomes"),
			bTable,
			xjs.SetInnerText(xdom.H2(), "Plants"),
			pTable,
		)
	}
}
