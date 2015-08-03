package main

import (
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

var mapsDiv = xjs.CreateElement("div")

func maps(c dom.Element) {
	if !mapsDiv.HasChildNodes() {
		xjs.SetInnerText(mapsDiv, "MAPS")
	}
	c.AppendChild(mapsDiv)
}
