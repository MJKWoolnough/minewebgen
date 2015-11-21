package main

import (
	"github.com/MJKWoolnough/gopherjs/style"
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

func init() {
	style.Add(`fieldset {
	padding-left : 0;
	padding-right : 0;
	border : 1px solid #000;
}
fieldset legend {
	border : 1px solid #000;
	margin-left : auto;
	margin-right : auto;
}

textarea {
	width : 400px;
	height : 200px;
}
`)
}

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
			body.Underlying().Set("spellcheck", "false")
			xjs.RemoveChildren(body)
			body.AppendChild(xdom.H1())
			SetTitle(title)
			body.AppendChild(tabs.New([]tabs.Tab{
				{"Servers", ServersTab()},
				{"Maps", MapsTab()},
				{"Generators", GeneratorsTab},
				{"Settings", settingsTab},
			}))
		}()
	})
}
