package main

import (
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
	for _, m := range list {
		sd := xjs.CreateElement("div")
		xjs.SetInnerText(sd, m.Name)
		mapsDiv.AppendChild(sd)
	}
}
