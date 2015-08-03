package main

import (
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xjs"

	"honnef.co/go/js/dom"
)

func main() {
	dom.GetWindow().AddEventListener("load", false, func(e dom.Event) {
		body := dom.GetWindow().Document().(dom.HTMLDocument).Body()
		xjs.RemoveChildren(body)
		body.AppendChild(xjs.SetInnerText(xjs.CreateElement("h1"), "Ferrumwood Server"))
		body.AppendChild(tabs.MakeTabs([]tabs.Tab{
			{"Maps", maps},
			{"Upload", upload},
		}))
	})
}
