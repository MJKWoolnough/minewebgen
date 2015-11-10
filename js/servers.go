package main

import (
	"strconv"
	"time"

	"github.com/MJKWoolnough/gopherjs/overlay"
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xform"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/MJKWoolnough/minewebgen/internal/data"
	"honnef.co/go/js/dom"
)

func serversTab(c dom.Element) {
	xjs.RemoveChildren(c)
	c.AppendChild(xjs.SetInnerText(xdom.H2(), "Servers"))
	ns := xdom.Button()
	c.AppendChild(xjs.SetInnerText(ns, "New Server"))
	ns.AddEventListener("click", false, func(dom.Event) {
		d := xdom.Div()
		o := overlay.New(d)
		d.AppendChild(transferFile("Server", "Upload/Download", 0, o))
		o.OnClose(func() {
			go serversTab(c)
		})
		xjs.Body().AppendChild(o)
	})
	s, err := RPC.ServerList()
	if err != nil {
		c.AppendChild(xjs.SetInnerText(xdom.Div(), err.Error()))
		return
	}
	if len(s) == 0 {
		c.AppendChild(xjs.SetInnerText(xdom.Div(), "No Servers"))
		return
	}
	t := xjs.AppendChildren(xdom.Table(), xjs.AppendChildren(xdom.Thead(), xjs.AppendChildren(xdom.Tr(),
		xjs.SetInnerText(xdom.Th(), "Server Name"),
	)))

	for _, serv := range s {
		name := xjs.SetInnerText(xdom.Td(), serv.Name)
		name.AddEventListener("click", false, func() func(dom.Event) {
			s := serv
			return func(dom.Event) {
				o := overlay.New(xjs.AppendChildren(xdom.Div(), tabs.New([]tabs.Tab{
					{"General", serverGeneral(s)},
					{"Properties", serverProperties(s)},
					{"Console", serverConsole(s)},
				})))
				o.OnClose(func() {
					go serversTab(c)
				})
				xjs.Body().AppendChild(o)
			}
		}())
		t.AppendChild(xjs.AppendChildren(xdom.Tr(), name))
	}
	c.AppendChild(t)
}

func serverGeneral(s data.Server) func(dom.Element) {
	return func(c dom.Element) {
		go func() {
			maps, err := RPC.MapList()
			if err != nil {
				c.AppendChild(xjs.SetInnerText(xdom.Div(), "Error getting map list: "+err.Error()))
				return
			}
			name := xform.InputText("name", s.Name)
			name.Required = true
			opts := make([]xform.Option, 1, len(maps)+1)
			opts[0] = xform.Option{
				Label:    "-- None -- ",
				Value:    "-1",
				Selected: s.Map == -1,
			}
			for i, m := range maps {
				n := m.Name
				if m.Server != -1 {
					if m.ID == s.Map {
						n = "* - " + n
					} else {
						n = "! - " + n
					}
				} else {
					n = "    " + n
				}
				opts = append(opts, xform.Option{
					Label:    n,
					Value:    strconv.Itoa(i),
					Selected: m.ID == s.Map,
				})
			}
			args := xform.InputSizeableList(s.Args...)
			sel := xform.SelectBox("map", opts...)
			submit := xform.InputSubmit("Set")
			submit.AddEventListener("click", false, func(e dom.Event) {
				if s.State != data.StateStopped {
					xjs.Alert("Cannot modify these settings while the server is running")
					return
				}
				if name.Value == "" {
					return
				}
				sID, err := strconv.Atoi(sel.Value)
				if err != nil || sID < -1 || sID >= len(maps) {
					return
				}
				submit.Disabled = true
				e.PreventDefault()
				if sID >= 0 {
					m := maps[sID]
					sID = m.ID
				}
				go func() {
					err = RPC.SetServerMap(s.ID, sID)
					if err != nil {
						xjs.Alert("Error setting server map: %s", err)
						return
					}
					s.Name = name.Value
					s.Args = args.Values()
					err = RPC.SetServer(s)
					if err != nil {
						xjs.Alert("Error setting server data: %s", err)
						return
					}
					span := xdom.Span()
					span.Style().Set("color", "#f00")
					c.AppendChild(xjs.SetInnerText(span, "Saved!"))
					time.Sleep(5 * time.Second)
					c.RemoveChild(span)
					submit.Disabled = false
				}()
			})
			xjs.AppendChildren(c, xjs.AppendChildren(xdom.Form(),
				xform.Label("Server Name", "name"),
				name,
				xdom.Br(),
				xform.Label("Arguments", "args"),
				args,
				xdom.Br(),
				xform.Label("Map Name", "map"),
				sel,
				xdom.Br(),
				submit,
			))
		}()
	}
}

type PropertyList [][2]string

func (p PropertyList) Len() int {
	return len(p)
}

func (p PropertyList) Less(i, j int) bool {
	return p[i][0] < p[j][0]
}

func (p PropertyList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func serverProperties(s data.Server) func(dom.Element) {
	return func(c dom.Element) {
		go editProperties(c, "Server", s.ID, RPC.ServerProperties, RPC.SetServerProperties)
	}
}

func serverConsole(s data.Server) func(dom.Element) {
	return func(c dom.Element) {
		c.AppendChild(xjs.SetInnerText(xdom.Div(), "Console"))
	}
}
