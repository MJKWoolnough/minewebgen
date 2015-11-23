package main

import (
	"strconv"
	"time"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/gopherjs/overlay"
	"github.com/MJKWoolnough/gopherjs/style"
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xform"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/MJKWoolnough/minewebgen/internal/data"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"
	"honnef.co/go/js/dom"
)

func init() {
	style.Add(`.serverName {
	background-color : #ddd;
	cursor: pointer;
	cursor: hand;
}
`)
}

type Server struct {
	data.Server
	row    dom.Node
	name   *dom.HTMLTableCellElement
	status *dom.HTMLTableCellElement
	button *dom.HTMLButtonElement
}

func ServersTab() func(dom.Element) {
	forceUpdate := make(chan struct{})
	ns := xdom.Button()
	ns.AddEventListener("click", false, func(dom.Event) {
		d := xdom.Div()
		o := overlay.New(d)
		d.AppendChild(transferFile("Server", "Upload/Download", 0, o))
		o.OnClose(func() {
			go func() {
				forceUpdate <- struct{}{}
			}()
		})
		xjs.Body().AppendChild(o)
	})
	noneTd := xdom.Td()
	noneTd.ColSpan = 3
	none := xjs.AppendChildren(xdom.Tr(), xjs.SetInnerText(noneTd, "No Servers Found"))
	serverList := xjs.AppendChildren(xdom.Table(),
		xjs.AppendChildren(xdom.Thead(), xjs.AppendChildren(xdom.Tr(),
			xjs.SetInnerText(xdom.Th(), "Server Name"),
			xjs.SetInnerText(xdom.Th(), "Status"),
			xjs.SetInnerText(xdom.Th(), "Controls"),
		)),
		none,
	)
	nodes := xjs.AppendChildren(xdom.Div(),
		xjs.SetInnerText(xdom.H2(), "Servers"),
		xjs.SetInnerText(ns, "New Server"),
		serverList,
	)
	servers := make(map[int]*Server)

	return func(c dom.Element) {
		c.AppendChild(nodes)
		updateStop := make(chan struct{})
		registerUpdateStopper(c, updateStop)
		for {
			servs, err := RPC.ServerList()
			if err != nil {
				xjs.Alert("Error getting server list: %s", err)
				return
			}

			if none.ParentNode() != nil {
				serverList.RemoveChild(none)
			}

			for _, s := range servers {
				s.ID = -1
			}

			for _, s := range servs {
				os, ok := servers[s.ID]
				if ok {
					os.Server = s
				} else {
					name := xdom.Td()
					status := xdom.Td()
					startStop := xdom.Button()
					os = &Server{
						Server: s,
						row: xjs.AppendChildren(xdom.Tr(),
							name,
							status,
							xjs.AppendChildren(xdom.Td(), startStop),
						),
						name:   name,
						status: status,
						button: startStop,
					}
					servers[s.ID] = os
					serverList.AppendChild(os.row)
					name.Class().SetString("serverName")
					name.AddEventListener("click", false, func() func(dom.Event) {
						s := os
						return func(dom.Event) {
							go func() {
								d, err := RPC.ServerEULA(s.ID)
								if err != nil {
									d = ""
								}
								t := []tabs.Tab{
									{"General", serverGeneral(s.Server)},
									{"Properties", serverProperties(s.Server)},
									{"Console", serverConsole(s.Server)},
								}
								if d != "" {
									t = append(t, tabs.Tab{"EULA", serverEULA(s.Server, d)})
								}
								div := xdom.Div()
								o := overlay.New(div)
								t = append(t, tabs.Tab{"Misc.", misc("server", strconv.Itoa(s.Server.ID), o, RPC.RemoveServer)})
								div.AppendChild(tabs.New(t))
								o.OnClose(func() {
									go func() {
										forceUpdate <- struct{}{}
									}()
								})
								xjs.Body().AppendChild(o)
							}()

						}
					}())
					startStop.AddEventListener("click", false, func() func(dom.Event) {
						b := startStop
						s := os
						return func(dom.Event) {
							go func() {
								b.Disabled = true
								switch s.State {
								case data.StateStopped:
									err := RPC.StartServer(s.ID)
									if err != nil {
										xjs.Alert("Error starting server: %s", err)
										return
									}
								case data.StateRunning:
									err := RPC.StopServer(s.ID)
									if err != nil {
										xjs.Alert("Error stopping server: %s", err)
										return
									}
								default:
									return
								}
								go func() {
									forceUpdate <- struct{}{}
								}()
							}()
						}
					}())
				}
				if os.Map >= 0 {
					xjs.SetInnerText(os.status, os.State.String())
					switch os.State {
					case data.StateStopped:
						xjs.SetInnerText(os.button, "Start")
						os.button.Disabled = false
					case data.StateRunning:
						xjs.SetInnerText(os.button, "Stop")
						os.button.Disabled = false
					default:
						xjs.SetInnerText(os.button, "N/A")
						os.button.Disabled = true
					}
				} else {
					xjs.SetInnerText(os.status, "No Map")
					os.button.Disabled = true
					xjs.SetInnerText(os.button, "N/A")
				}
				xjs.SetInnerText(os.name, os.Name)
			}

			for id, s := range servers {
				if s.ID == -1 {
					delete(servers, id)
					serverList.RemoveChild(s.row)
				}
			}

			if len(servers) == 0 {
				serverList.AppendChild(none)
			}

			// Sleep until update
			if !updateSleep(forceUpdate, updateStop) {
				return
			}
		}
	}
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
		log := xform.TextArea("log", "")
		log.ReadOnly = true
		command := xform.InputText("command", "")
		command.Required = true
		send := xform.InputSubmit("Send")
		c.AppendChild(xjs.AppendChildren(xdom.Form(), xjs.AppendChildren(xdom.Fieldset(),
			xjs.SetInnerText(xdom.Legend(), "Console"),
			xform.Label("Log", ""), log, xdom.Br(),
			xform.Label("Command", "command"), command, send,
		)))
		if s.State == data.StateStopped {
			send.Disabled = true
			command.Disabled = true
		} else {
			send.AddEventListener("click", false, func(e dom.Event) {
				if command.Value == "" {
					return
				}
				e.PreventDefault()
				send.Disabled = true
				cmd := command.Value
				log.Value += "\n>" + cmd + "\n"
				log.Set("scrollTop", log.Get("scrollHeight"))
				command.Value = ""
				go func() {
					err := RPC.WriteCommand(s.ID, cmd)
					if err != nil {
						xjs.Alert("Error sending command: %s", err)
						return
					}
					send.Disabled = false
				}()
			})
		}
		go func() {
			conn, err := websocket.Dial("ws://" + js.Global.Get("location").Get("host").String() + "/console")
			if err != nil {
				xjs.Alert("Failed to connect to console: %s", err)
				return
			}
			defer conn.Close()
			w := byteio.StickyWriter{Writer: byteio.LittleEndianWriter{Writer: conn}}
			r := byteio.StickyReader{Reader: byteio.LittleEndianReader{Reader: conn}}
			updateStop := make(chan struct{})
			registerUpdateStopper(c, updateStop)
			done := false
			go func() {
				<-updateStop
				done = true
				conn.Close()
			}()
			w.WriteInt32(int32(s.ID))
			for {
				state := r.ReadUint8()
				switch state {
				case 0:
					if !done {
						err := ReadError(&r)
						if r.Err != nil {
							err = r.Err
						}
						log.Value += "\n\nError reading from console: " + err.Error()
						log.Set("scrollTop", log.Get("scrollHeight"))
					}
					return
				case 1:
					log.Value += ReadString(&r)
					log.Set("scrollTop", log.Get("scrollHeight"))
				}
			}
		}()
	}
}

func serverEULA(s data.Server, d string) func(dom.Element) {
	return func(c dom.Element) {
		t := xform.TextArea("eula", d)
		submit := xform.InputSubmit("Save")
		c.AppendChild(xjs.AppendChildren(xdom.Form(), xjs.AppendChildren(xdom.Fieldset(),
			xjs.SetInnerText(xdom.Legend(), "End User License Agreement"),
			xform.Label("EULA", "eula"), t, xdom.Br(),
			submit,
		)))
		submit.AddEventListener("click", false, func(e dom.Event) {
			e.PreventDefault()
			submit.Disabled = true
			go func() {
				err := RPC.SetServerEULA(s.ID, t.Value)
				if err != nil {
					xjs.Alert("Error setting server EULA: %s", err)
					return
				}
				d = t.Value
				span := xdom.Span()
				span.Style().Set("color", "#f00")
				c.AppendChild(xjs.SetInnerText(span, "Saved!"))
				time.Sleep(5 * time.Second)
				c.RemoveChild(span)
				submit.Disabled = false
			}()
		})
	}
}
