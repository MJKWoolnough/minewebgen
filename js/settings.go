package main

import (
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

func settings(c dom.Element) {
	c.AppendChild(xjs.SetInnerText(xdom.H1(), "Settings"))
}
