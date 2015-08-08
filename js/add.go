package main

import (
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

func add(c dom.Element) {
	u := xjs.CreateElement("input")
	u.SetAttribute("value", "Upload")
	u.SetAttribute("type", "button")
	u.AddEventListener("click", false, func(dom.Event) {
		xjs.RemoveChildren(c)
		upload(c)
	})
	d := xjs.CreateElement("input")
	d.SetAttribute("value", "Download")
	d.SetAttribute("type", "button")
	d.AddEventListener("click", false, func(dom.Event) {
		xjs.RemoveChildren(c)
		download(c)
	})
	g := xjs.CreateElement("input")
	g.SetAttribute("value", "Generate")
	g.SetAttribute("type", "button")
	g.AddEventListener("click", false, func(dom.Event) {
		xjs.RemoveChildren(c)
		generate(c)
	})
	c.AppendChild(u)
	c.AppendChild(d)
	c.AppendChild(g)
}
