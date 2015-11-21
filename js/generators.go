package main

import (
	"github.com/MJKWoolnough/gopherjs/overlay"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

func GeneratorsTab(c dom.Element) {
	go func() {
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
						xjs.Alert("%s", g)
					}
				}(g))
				table.AppendChild(xjs.AppendChildren(xdom.Tr(), xjs.SetInnerText(td, g)))
			}
		}
		c.AppendChild(table)
	}()
}
