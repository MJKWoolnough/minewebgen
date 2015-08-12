package main

import (
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

func maps(c dom.Element) {
	mapsDiv := xjs.CreateElement("div")
	defer c.AppendChild(mapsDiv)
	list, err := ServerList()
	if err != nil {
		xjs.SetInnerText(mapsDiv, err.Error())
		return
	}
	for _, s := range list {
		sd := xjs.CreateElement("div")
		xjs.SetInnerText(sd, s.Name)
		mapsDiv.AppendChild(sd)
	}
}
