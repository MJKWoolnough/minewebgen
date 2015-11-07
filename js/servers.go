package main

import (
	"github.com/MJKWoolnough/gopherjs/overlay"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xform"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

func serversTab(c dom.Element) {
	c.AppendChild(xjs.SetInnerText(xdom.H2(), "Servers"))
	ns := xdom.Button()
	c.AppendChild(xjs.SetInnerText(ns, "New Server"))
	c.AddEventListener("click", false, WrapEvent(newServer, c))
}

func newServer(c ...dom.Element) {
	d := xdom.Div()
	o := overlay.New(d)
	o.OnClose(func() {
		xjs.RemoveChildren(c[0])
		go serversTab(c[0])
	})
	d.AppendChild(xjs.SetInnerText(xdom.H1(), "New Server"))

	d.AppendChild(xform.Label("Server Name", "serverName"))
	sn := xform.InputText("serverName", "")
	d.AppendChild(sn)
	d.AppendChild(xdom.Br())

	d.AppendChild(xform.Label("URL", "url"))
	url := xform.InputRadio("url", "switch", true)
	d.AppendChild(url)
	d.AppendChild(xdom.Br())

	d.AppendChild(xform.Label("Upload", "upload"))
	upload := xform.InputRadio("upload", "switch", false)
	d.AppendChild(upload)
	d.AppendChild(xdom.Br())

	d.AppendChild(xform.Label("File", ""))
	fileI := xform.InputUpload("")
	urlI := xform.InputText("", "")

	d.AppendChild(fileI)
	d.AppendChild(urlI)

	typeFunc := func(dom.Event) {
		if url.Checked {
			urlI.Style().RemoveProperty("display")
			fileI.Style().SetProperty("display", "none", "")
		} else {
			fileI.Style().RemoveProperty("display")
			urlI.Style().SetProperty("display", "none", "")
		}
	}

	typeFunc(nil)

	url.AddEventListener("change", false, typeFunc)
	upload.AddEventListener("change", false, typeFunc)

	d.AppendChild(xdom.Br())

	s := xdom.Button()
	d.AppendChild(xjs.SetInnerText(s, "Create"))
	s.AddEventListener("click", false, func(dom.Event) {
		s.Disabled = true
	})

	dom.GetWindow().Document().(dom.HTMLDocument).Body().AppendChild(o)
}
