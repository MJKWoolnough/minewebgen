package main

import (
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

func main() {
	dom.GetWindow().AddEventListener("load", false, func(dom.Event) {
		go func() {
			if err := rpcInit(); err != nil {
				xjs.Alert("Failed to connect to RPC server: %s", err)
				return
			}
			title, err := RPC.ServerName()
			if err != nil {
				xjs.Alert("Error retrieving server name: %s", err)
				return
			}
			body := dom.GetWindow().Document().(dom.HTMLDocument).Body()
			xjs.RemoveChildren(body)
			body.AppendChild(xjs.SetInnerText(xdom.H1(), title+" Server"))
			tDoc, ok := dom.GetWindow().Document().(dom.HTMLDocument)
			if ok {
				tDoc.SetTitle(title + " Server")
			}
			body.AppendChild(tabs.MakeTabs([]tabs.Tab{
				{"Servers", serversTab},
				{"Maps", mapsTab},
				{"Settings", settingsTab},
			}))
		}()
	})
}
