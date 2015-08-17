package main

import (
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

func servers(c dom.Element) {
	serversDiv := xjs.CreateElement("div")
	defer c.AppendChild(serversDiv)
	list, err := MapList()
	if err != nil {
		xjs.SetInnerText(serversDiv, err.Error())
		return
	}
	for _, s := range list {
		sd := xjs.CreateElement("div")
		xjs.SetInnerText(sd, s.Name)
		serversDiv.AppendChild(sd)
	}
}
