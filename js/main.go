package main

import (
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

func SetTitle(title string) {
	title += " Server"
	xjs.SetInnerText(dom.GetWindow().Document().(dom.HTMLDocument).Body().ChildNodes()[0], title)
	tDoc, ok := dom.GetWindow().Document().(dom.HTMLDocument)
	if ok {
		tDoc.SetTitle(title)
	}
}

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
			body.AppendChild(xdom.H1())
			SetTitle(title)
			body.AppendChild(tabs.New([]tabs.Tab{
				{"Servers", serversTab},
				{"Maps", mapsTab},
				{"Settings", settingsTab},
			}))
		}()
	})
}
