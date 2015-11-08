package main

import (
	"github.com/MJKWoolnough/gopherjs/overlay"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xform"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

func serversTab(c dom.Element) {
	xjs.RemoveChildren(c)
	c.AppendChild(xjs.SetInnerText(xdom.H2(), "Servers"))
	ns := xdom.Button()
	c.AppendChild(xjs.SetInnerText(ns, "New Server"))
	ns.AddEventListener("click", false, WrapEvent(newServer, c))
}

func newServer(c ...dom.Element) {
	sn := xform.InputText("serverName", "")
	url := xform.InputRadio("url", "switch", true)
	upload := xform.InputRadio("upload", "switch", false)
	fileI := xform.InputUpload("")
	urlI := xform.InputText("", "")
	s := xform.InputSubmit("Create")

	sn.Required = true

	typeFunc := func(dom.Event) {
		if url.Checked {
			urlI.Style().RemoveProperty("display")
			fileI.Style().SetProperty("display", "none", "")
			urlI.Required = true
			fileI.Required = false
			fileI.SetID("")
			urlI.SetID("file")
		} else {
			fileI.Style().RemoveProperty("display")
			urlI.Style().SetProperty("display", "none", "")
			fileI.Required = true
			urlI.Required = false
			urlI.SetID("")
			fileI.SetID("file")
		}
	}

	typeFunc(nil)

	url.AddEventListener("change", false, typeFunc)
	upload.AddEventListener("change", false, typeFunc)

	o := overlay.New(xjs.AppendChildren(xdom.Form(), xjs.AppendChildren(xdom.Fieldset(),
		xjs.SetInnerText(xdom.Legend(), "New Server"),
		xform.Label("Server Name", "serverName"),
		sn,
		xdom.Br(),

		xform.Label("URL", "url"),
		url,
		xdom.Br(),

		xform.Label("Upload", "upload"),
		upload,
		xdom.Br(),

		xform.Label("File", "file"),

		fileI,
		urlI,

		xdom.Br(),
		s,
	)))
	o.OnClose(func() {
		go serversTab(c[0])
	})

	s.AddEventListener("click", false, func(e dom.Event) {
		if sn.Value == "" {
			return
		}
		if url.Checked {
			if urlI.Value == "" {
				return
			}
		} else if len(fileI.Files()) != 1 {
			return

		}
		s.Disabled = true
		sn.Disabled = true
		url.Disabled = true
		upload.Disabled = true
		fileI.Disabled = true
		urlI.Disabled = true
		e.PreventDefault()
		go func() {

		}()
	})

	xjs.Body().AppendChild(o)
}
