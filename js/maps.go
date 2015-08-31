package main

import (
	"github.com/MJKWoolnough/gopherjs/overlay"
	"github.com/MJKWoolnough/gopherjs/tabs"
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

	newButton := xjs.CreateElement("input").(*dom.HTMLInputElement)
	newButton.Type = "button"
	newButton.Value = "New Map"
	newButton.AddEventListener("click", false, newMap(c))

	mapsDiv.AppendChild(newButton)

	for _, m := range list {
		sd := xjs.CreateElement("div")
		xjs.SetInnerText(sd, m.Name)
		mapsDiv.AppendChild(sd)
	}
	c.AppendChild(mapsDiv)
}

func newMap(c dom.Element) func(dom.Event) {
	return func(dom.Event) {
		f := xjs.CreateElement("div")
		o := overlay.New(f)
		f.AppendChild(tabs.MakeTabs([]tabs.Tab{
			{"Create", createMap},
			{"Upload/Download", uploadMap},
			{"Generate", generate},
		}))
		o.OnClose(func() {
			maps(c)
		})
		c.AppendChild(o)
	}
}

func createMap(c dom.Element) {

}

func uploadMap(c dom.Element) {

}
