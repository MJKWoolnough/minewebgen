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
	sn := xform.InputText("serverName", s.ServerName)
	sn.Required = true
	la := xform.InputText("listenAddr", s.ListenAddr)
	la.Required = true
	sp := xform.InputText("serversPath", s.DirServers)
	sp.Required = true
	mp := xform.InputText("mapsPath", s.DirMaps)
	mp.Required = true
	gp := xform.InputText("gensPath", s.DirGenerators)
	gp.Required = true
	ge := xform.InputText("genExe", s.GeneratorExecutable)
	ge.Required = true
	mm := xform.InputNumber("maxMem", 100, 10000, float64(s.GeneratorMaxMem/1024/1024))
	mm.Required = true
	sb := xform.InputSubmit("Save")
	sb.AddEventListener("click", false, func(e dom.Event) {
		if sn.Value == "" || la.Value == "" || sp.Value == "" || mp.Value == "" || gp.Value == "" || ge.Value == "" || !mm.CheckValidity() {
			return
		}
		e.PreventDefault()
		sb.Disabled = true
		go func() {
			s.ServerName = sn.Value
			s.ListenAddr = la.Value
			s.DirServers = sp.Value
			s.DirMaps = mp.Value
			s.DirGenerators = gp.Value
			s.GeneratorExecutable = ge.Value
			s.GeneratorMaxMem = int64(mm.ValueAsNumber * 1024 * 1024)
			if err := RPC.SetSettings(s); err != nil {
				xjs.Alert("Error saving settings: %s", err)
				return
			}
			SetTitle(sn.Value)
			sb.Disabled = false
		}()
	})
	xjs.AppendChildren(c, xjs.AppendChildren(xdom.Form(), xjs.AppendChildren(xdom.Fieldset(),
		xjs.SetInnerText(xdom.Legend(), "Change Settings"),
		xform.Label("Server Name", "serverName"), sn, xdom.Br(),
		xform.Label("Listen Address", "listenAddr"), la, xdom.Br(),
		xform.Label("Servers Path", "serversPath"), sp, xdom.Br(),
		xform.Label("Maps Path", "mapsPath"), mp, xdom.Br(),
		xform.Label("Generators Path", "gensPath"), gp, xdom.Br(),
		xform.Label("Generator Executable", "genExe"), ge, xdom.Br(),
		xform.Label("Generator Memory (MB)", "mamMem"), mm, xdom.Br(),
		sb,
	)))
}
