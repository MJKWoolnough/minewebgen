package main

import (
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xform"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"honnef.co/go/js/dom"
)

func settingsTab(c dom.Element) {
	s, err := RPC.Settings()
	if err != nil {
		xjs.Alert("Error reading settings: %s", err)
		return
	}
	c.AppendChild(xjs.SetInnerText(xdom.H2(), "Change Settings"))
	c.AppendChild(xform.Label("Server Name", "serverName"))
	sn := xform.InputText("serverName", s.ServerName)
	c.AppendChild(sn)
	c.AppendChild(xdom.Br())
	c.AppendChild(xform.Label("Listen Address", "listenAddr"))
	la := xform.InputText("listenAddr", s.ListenAddr)
	c.AppendChild(la)
	c.AppendChild(xdom.Br())
	c.AppendChild(xform.Label("Servers Path", "serversPath"))
	sp := xform.InputText("serversPath", s.DirServers)
	c.AppendChild(sp)
	c.AppendChild(xdom.Br())
	c.AppendChild(xform.Label("Maps Path", "mapsPath"))
	mp := xform.InputText("mapsPath", s.DirMaps)
	c.AppendChild(mp)
	c.AppendChild(xdom.Br())

	sb := xdom.Button()
	xjs.SetInnerText(sb, "Save")
	sb.AddEventListener("click", false, func(dom.Event) {
		go func() {
			s.ServerName = sn.Value
			s.ListenAddr = la.Value
			s.DirServers = sp.Value
			s.DirMaps = mp.Value
			if err := RPC.SetSettings(s); err != nil {
				xjs.Alert("Error saving settings: %s", err)
			} else {
				SetTitle(sn.Value)
			}
		}()
	})
	c.AppendChild(sb)
}
