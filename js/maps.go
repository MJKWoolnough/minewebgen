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

func mapsTab(c dom.Element) {
	xjs.RemoveChildren(c)
	c.AppendChild(xjs.SetInnerText(xdom.H2(), "Maps"))
	nm := xdom.Button()
	c.AppendChild(xjs.SetInnerText(nm, "New Map"))
	nm.AddEventListener("click", false, func(dom.Event) {
		d := xdom.Div()
		o := overlay.New(d)
		o.OnClose(func() {
			go mapsTab(c)
		})
		xjs.AppendChildren(d,
			xjs.SetInnerText(xdom.H1(), "New Map"),
			tabs.New([]tabs.Tab{
				{"Create", createMap(o)},
				{"Upload/Download", func(c dom.Element) {
					c.AppendChild(transferFile("Map", "Upload/Download", 1, o))
				}},
				{"Generate", func(c dom.Element) {
					c.AppendChild(transferFile("Map", "Generate", 2, o))
				}},
			}),
		)
		xjs.Body().AppendChild(o)
	})
	m, err := RPC.MapList()
	if err != nil {
		c.AppendChild(xjs.SetInnerText(xdom.Div(), err.Error()))
		return
	}
	if len(m) == 0 {
		c.AppendChild(xjs.SetInnerText(xdom.Div(), "No Maps"))
		return
	}
	t := xjs.AppendChildren(xdom.Table(), xjs.AppendChildren(xdom.Thead(), xjs.AppendChildren(xdom.Tr(),
		xjs.SetInnerText(xdom.Th(), "Map Name"),
	)))

	for _, mp := range m {
		name := xjs.SetInnerText(xdom.Td(), mp.Name)
		name.AddEventListener("click", false, func() func(dom.Event) {
			m := mp
			return func(dom.Event) {
				o := overlay.New(xjs.AppendChildren(xdom.Div(), tabs.New([]tabs.Tab{
					{"General", mapGeneral(m)},
					{"Properties", mapProperties(m)},
				})))
				o.OnClose(func() {
					mapsTab(c)
				})
				xjs.Body().AppendChild(o)
			}
		}())
		t.AppendChild(xjs.AppendChildren(xdom.Tr(), name))
	}
	c.AppendChild(t)
}

func createMap(o *overlay.Overlay) func(dom.Element) {
	return func(c dom.Element) {

	}
}

func mapGeneral(m data.Map) func(dom.Element) {
	return func(c dom.Element) {
		go func() {
			servers, err := RPC.ServerList()
			if err != nil {
				c.AppendChild(xjs.SetInnerText(xdom.Div(), "Error getting server list: "+err.Error()))
				return
			}
			name := xform.InputText("name", m.Name)
			name.Required = true
			opts := make([]xform.Option, 1, len(servers)+1)
			opts[0] = xform.Option{
				Label:    "-- None -- ",
				Value:    "-1",
				Selected: m.Server == -1,
			}
			for i, s := range servers {
				n := s.Name
				if s.Map == -1 {
					if s.ID == m.Server {
						n += "* - " + n
					} else {
						n += "! - " + n
					}
				} else {
					n = "    " + n
				}
				opts = append(opts, xform.Option{
					Label:    n,
					Value:    strconv.Itoa(i),
					Selected: s.ID == m.Server,
				})
			}
			sel := xform.SelectBox("server", opts...)
			submit := xform.InputSubmit("Set")
			submit.AddEventListener("click", false, func(e dom.Event) {
				if name.Value == "" {
					return
				}
				mID, err := strconv.Atoi(sel.Value)
				if err != nil || mID < -1 || mID >= len(servers) {
					return
				}
				submit.Disabled = true
				e.PreventDefault()
				if mID >= 0 {
					s := servers[mID]
					mID = s.ID
				}
				go func() {
					err = RPC.SetServerMap(mID, m.ID)
					if err != nil {
						xjs.Alert("Error setting map server: %s", err)
						return
					}
					m.Name = name.Value
					err = RPC.SetMap(m)
					if err != nil {
						xjs.Alert("Error setting map data: %s", err)
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
				xform.Label("Server Name", "server"),
				sel,
				xdom.Br(),
				submit,
			))
		}()
	}
}

func mapProperties(m data.Map) func(dom.Element) {
	return func(c dom.Element) {
		go editProperties(c, "Map", m.ID, RPC.MapProperties, RPC.SetMapProperties)
	}
}
