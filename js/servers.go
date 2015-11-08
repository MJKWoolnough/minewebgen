package main

import (
	"image/color"
	"io"
	"sort"
	"strconv"
	"time"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/gopherjs/files"
	"github.com/MJKWoolnough/gopherjs/overlay"
	"github.com/MJKWoolnough/gopherjs/progress"
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xform"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/MJKWoolnough/minewebgen/internal/data"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"
	"honnef.co/go/js/dom"
)

func serversTab(c dom.Element) {
	xjs.RemoveChildren(c)
	c.AppendChild(xjs.SetInnerText(xdom.H2(), "Servers"))
	ns := xdom.Button()
	c.AppendChild(xjs.SetInnerText(ns, "New Server"))
	ns.AddEventListener("click", false, WrapEvent(newServer, c))
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
			d := xdom.Div()
			uo := overlay.New(d)
			uo.OnClose(func() {
				o.Close()
			})
			xjs.Body().AppendChild(uo)
			status := xdom.Div()
			d.AppendChild(xjs.SetInnerText(status, "Transferring..."))
			conn, err := websocket.Dial("ws://" + js.Global.Get("location").Get("host").String() + "/transfer")
			if err != nil {
				xjs.SetInnerText(status, err.Error())
				return
			}
			defer conn.Close()
			w := byteio.StickyWriter{Writer: byteio.LittleEndianWriter{Writer: conn}}
			r := byteio.StickyReader{Reader: byteio.LittleEndianReader{Reader: conn}}

			pb := progress.New(color.RGBA{255, 0, 0, 0}, color.RGBA{0, 0, 255, 0}, 400, 50)
			d.AppendChild(pb)

			if url.Checked {
				w.WriteUint8(0)
				WriteString(&w, urlI.Value)
				length := int(r.ReadInt32())
				total := 0
				for total < length {
					switch v := r.ReadUint8(); v {
					case 1:
						i := int(r.ReadInt32())
						total += i
						pb.Percent(100 * total / length)
					default:
						xjs.SetInnerText(status, ReadError(&r).Error())
						return
					}
				}
			} else {
				f := files.NewFileReader(files.NewFile(fileI.Files()[0]))
				l := f.Len()
				if l == 0 {
					xjs.SetInnerText(status, "Zero-length file")
					return
				}
				w.WriteUint8(1)
				w.WriteInt32(int32(l))
				io.Copy(&w, pb.Reader(f, l))
			}

			d.RemoveChild(pb)
			xjs.SetInnerText(status, "Checking File")

			WriteString(&w, sn.Value)

			for {
				switch v := r.ReadUint8(); v {
				case 0:
					xjs.SetInnerText(status, ReadError(&r).Error())
					return
				case 1:
					jars := make([]xform.Option, r.ReadInt16())
					for i := range jars {
						jars[i] = xform.Option{
							Value: strconv.Itoa(i),
							Label: ReadString(&r),
						}
					}
					j := xform.SelectBox("jars", jars...)
					sel := xjs.SetInnerText(xdom.Button(), "Select")
					jo := overlay.New(xjs.AppendChildren(xdom.Div(), xjs.AppendChildren(xdom.Fieldset(),
						xjs.SetInnerText(xdom.Legend(), "Please select the server jar"),
						xform.Label("Jar File", "jars"),
						j,
						xdom.Br(),
						sel,
					)))
					c := make(chan int16, 0)
					done := false
					jo.OnClose(func() {
						if !done {
							done = true
							c <- -1
						}
					})
					sel.AddEventListener("click", false, func(dom.Event) {
						if !done {
							done = true
							v, err := strconv.Atoi(j.Value)
							if err != nil {
								v = -1
							}
							c <- int16(v)
							jo.Close()
						}
					})
					xjs.Body().AppendChild(jo)
					w.WriteInt16(<-c)
					close(c)
				case 255:
					uo.Close()
					return
				}
			}

		}()
	})

	xjs.Body().AppendChild(o)
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
						n += "* - " + n
					} else {
						n = "! - " + n
					}
				} else {
					n = "   " + n
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
		go func() {
			sp, err := RPC.ServerProperties(s.ID)
			if err != nil {
				c.AppendChild(xjs.SetInnerText(xdom.Div(), "Failed to get server properties: "+err.Error()))
				return
			}
			props := make(PropertyList, 0, len(sp))
			for k, v := range sp {
				props = append(props, [2]string{k, v})
			}
			sort.Sort(props)
			propE := make([][2]*dom.HTMLSpanElement, len(props))
			df := xjs.DocumentFragment()

			toggleFunc := func(k, v *dom.HTMLSpanElement, toggle *dom.HTMLInputElement) func(dom.Event) {
				return func(dom.Event) {
					if toggle.Value == "-" {
						k.SetContentEditable("false")
						v.SetContentEditable("false")
						k.Style().SetProperty("background-color", "#888", "")
						v.Style().SetProperty("background-color", "#888", "")
						toggle.Value = "+"
					} else {
						k.SetContentEditable("true")
						v.SetContentEditable("true")
						k.Style().RemoveProperty("background-color")
						v.Style().RemoveProperty("background-color")
						toggle.Value = "-"
					}
				}
			}

			for i, prop := range props {
				k := xform.InputSizeable("", prop[0])
				v := xform.InputSizeable("", prop[1])
				toggle := xform.InputButton("-")
				toggle.AddEventListener("click", false, toggleFunc(k, v, toggle))
				propE[i][0] = k
				propE[i][1] = v
				xjs.AppendChildren(df,
					toggle,
					k,
					xjs.SetInnerText(xdom.Span(), "="),
					v,
					xdom.Br(),
				)
			}

			add := xform.InputButton("Add")
			submit := xform.InputButton("Save")
			fs := xjs.AppendChildren(xdom.Fieldset(), xjs.AppendChildren(
				df,
				xjs.SetInnerText(xdom.Legend(), "Server Properties"),
				add,
				submit,
			))

			add.AddEventListener("click", false, func(dom.Event) {
				k := xform.InputSizeable("", "")
				v := xform.InputSizeable("", "")
				toggle := xform.InputButton("-")
				toggle.AddEventListener("click", false, toggleFunc(k, v, toggle))
				propE = append(propE, [2]*dom.HTMLSpanElement{k, v})
				fs.InsertBefore(toggle, add)
				fs.InsertBefore(k, add)
				fs.InsertBefore(xjs.SetInnerText(xdom.Span(), "="), add)
				fs.InsertBefore(v, add)
				fs.InsertBefore(xdom.Br(), add)
			})

			submit.AddEventListener("click", false, func(dom.Event) {
				submit.Disabled = true
				props := make(map[string]string, len(propE))
				for _, spans := range propE {
					if spans[0].IsContentEditable() {
						props[spans[0].TextContent()] = spans[1].TextContent()
					}
				}
				go func() {
					err := RPC.SetServerProperties(s.ID, props)
					if err != nil {
						xjs.Alert("Error setting server properties: %s", err)
						return
					}
					span := xdom.Span()
					span.Style().Set("color", "#f00")
					fs.AppendChild(xjs.SetInnerText(span, "Saved!"))
					time.Sleep(5 * time.Second)
					fs.RemoveChild(span)
					submit.Disabled = false
				}()
			})

			xjs.AppendChildren(c, xjs.AppendChildren(xdom.Form(), fs))
		}()
	}
}

func serverConsole(s data.Server) func(dom.Element) {
	return func(c dom.Element) {
		c.AppendChild(xjs.SetInnerText(xdom.Div(), "Console"))
	}
}
