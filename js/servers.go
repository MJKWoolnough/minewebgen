package main

import (
	"image/color"
	"io"
	"strconv"
	"strings"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/gopherjs/files"
	"github.com/MJKWoolnough/gopherjs/progress"
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"
	"honnef.co/go/js/dom"

	"github.com/MJKWoolnough/gopherjs/overlay"
)

func servers(c dom.Element) {
	xjs.RemoveChildren(c)
	serversDiv := xdom.Div()
	defer c.AppendChild(serversDiv)
	list, err := RPC.ServerList()
	if err != nil {
		xjs.SetInnerText(serversDiv, err.Error())
		return
	}
	newButton := xdom.Input()
	newButton.Value = "New Server"
	newButton.Type = "button"
	newButton.AddEventListener("click", false, newServer(c))
	c.AppendChild(newButton)
	table := xdom.Table()
	head := xdom.Tr()
	head.AppendChild(xjs.SetInnerText(xdom.Th(), "Name"))
	head.AppendChild(xjs.SetInnerText(xdom.Th(), "Status"))
	head.AppendChild(xjs.SetInnerText(xdom.Th(), "Controls"))
	table.AppendChild(head)
	for _, s := range list {
		tr := xdom.Tr()
		name := xdom.Td()
		xjs.SetInnerText(name, s.Name)
		name.AddEventListener("click", false, viewServer(c, s))
		tr.AppendChild(name)
		status := xdom.Tr()
		xjs.SetInnerText(status, "")
		tr.AppendChild(status)
		controls := xdom.Td()
		tr.AppendChild(controls)
		if s.Map >= 0 {
			b := xdom.Input()
			b.Type = "button"
			if s.IsRunning() {
				b.Value = "Stop"
				b.AddEventListener("click", false, stopServer(c, s))
			} else {
				b.Value = "Start"
				b.AddEventListener("click", false, startServer(c, s))
			}
			controls.AppendChild(b)
		}
		table.AppendChild(tr)
	}
	serversDiv.AppendChild(table)
	c.AppendChild(serversDiv)
}

func startServer(c dom.Element, s Server) func(dom.Event) {
	return func(dom.Event) {
		go func() {
			err := RPC.ServerStart(s.ID)
			if err != nil {
				xjs.Alert("%s", err)
			}
		}()
	}
}

func stopServer(c dom.Element, s Server) func(dom.Event) {
	return func(dom.Event) {
		go func() {
			err := RPC.ServerStop(s.ID)
			if err != nil {
				xjs.Alert("%s", err)
			}
		}()
	}
}

func newServer(c dom.Element) func(dom.Event) {
	return func(dom.Event) {
		f := xdom.Div()
		o := overlay.New(f)
		f.SetID("serverUpload")

		f.AppendChild(xjs.SetInnerText(xdom.H1(), "New Server"))

		nameLabel := xdom.Label()
		nameLabel.For = "name"
		xjs.SetInnerText(nameLabel, "Level Name")
		nameInput := xdom.Input()
		nameInput.Type = "text"
		nameInput.SetID("name")

		urlLabel := xdom.Label()
		urlLabel.For = "url"
		xjs.SetInnerText(urlLabel, "URL")
		urlInput := xdom.Input()
		urlInput.Type = "radio"
		urlInput.Name = "type"
		urlInput.SetID("url")
		urlInput.Checked = true

		uploadLabel := xdom.Label()
		uploadLabel.For = "upload"
		xjs.SetInnerText(uploadLabel, "Upload")
		uploadInput := xdom.Input()
		uploadInput.Type = "radio"
		uploadInput.Name = "type"
		uploadInput.SetID("upload")

		fileLabel := xdom.Label()
		fileLabel.For = "file"
		xjs.SetInnerText(fileLabel, "File")
		fileInput := xdom.Input()
		fileInput.Type = "text"
		fileInput.SetID("file")

		urlInput.AddEventListener("click", false, func(dom.Event) {
			fileInput.Type = "text"
		})

		uploadInput.AddEventListener("click", false, func(dom.Event) {
			fileInput.Type = "file"
		})

		submit := xdom.Input()
		submit.Value = "Submit"
		submit.Type = "button"

		submit.AddEventListener("click", false, func(e dom.Event) {
			name := nameInput.Value
			if len(name) == 0 {
				dom.GetWindow().Alert("Name cannot be empty")
				return
			}
			var file readLener
			uploadType := uint8(3)
			if fileInput.Type == "file" {
				uploadType = 4
				fs := fileInput.Files()
				if len(fs) != 1 {
					dom.GetWindow().Alert("File Error occurred")
					return
				}
				f := files.NewFile(fs[0])
				file = files.NewFileReader(f)
			} else {
				url := fileInput.Value
				if len(url) == 0 {
					dom.GetWindow().Alert("URL cannot be empty")
					return
				}
				file = strings.NewReader(url)
			}
			length := file.Len()
			status := xdom.Div()
			pb := progress.New(color.RGBA{255, 0, 0, 0}, color.RGBA{0, 0, 255, 0}, 400, 50)
			xjs.RemoveChildren(f)
			f.AppendChild(status)
			f.AppendChild(pb)

			go func() {
				conn, err := websocket.Dial("ws://" + js.Global.Get("location").Get("host").String() + "/upload")
				if err != nil {
					xjs.SetInnerText(status, err.Error())
					return
				}
				defer removeCloser(closeOnExit(conn))
				defer conn.Close()
				o.OnClose(func() { conn.Close() })

				w := &byteio.StickyWriter{Writer: &byteio.LittleEndianWriter{Writer: conn}}
				xjs.SetInnerText(status, "Uploading Data...")
				uploadFile(uploadType, pb.Reader(file, length), w)
				if w.Err != nil {
					xjs.SetInnerText(status, w.Err.Error())
					return
				}

				r := &byteio.StickyReader{Reader: &byteio.LittleEndianReader{conn}}

				if r.ReadUint8() == 0 {
					xjs.SetInnerText(status, readError(r).Error())
					return
				}

				f.RemoveChild(pb)
				xjs.SetInnerText(status, "Checking Zip...")

				w.WriteUint8(uint8(len(name)))
				w.Write([]byte(name))
				for {
					switch r.ReadUint8() {
					case 1:
						numJars := r.ReadInt16()
						jars := make([]string, numJars)
						for i := int16(0); i < numJars; i++ {
							jars[i] = readString(r)
						}
						if r.Err != nil {
							xjs.SetInnerText(status, r.Err.Error())
						}

						c := make(chan int16, 1)

						jarSelect := xdom.Div()
						jso := overlay.New(jarSelect)
						selected := false
						jso.OnClose(func() {
							if !selected {
								selected = true
								c <- -1
							}
						})

						jarSelect.AppendChild(xjs.SetInnerText(xdom.H1(), "Select Server JAR"))
						radios := make([]*dom.HTMLInputElement, numJars)

						for num, name := range jars {
							r := xdom.Input()
							r.Type = "radio"
							r.Name = "jarChoose"
							v := strconv.Itoa(num)
							r.Value = v
							r.SetID("jarChoose_" + v)
							if num == 0 {
								r.DefaultChecked = true
							}

							l := xdom.Label()
							xjs.SetInnerText(l, name)
							l.For = "jarChoose_" + v

							jarSelect.AppendChild(r)
							jarSelect.AppendChild(l)
							jarSelect.AppendChild(xdom.Br())
							radios[num] = r
						}

						choose := xdom.Input()
						choose.Type = "button"
						choose.Value = "Select"
						choose.AddEventListener("click", false, func(dom.Event) {
							if !selected {
								selected = true
								choice := -1
								for num, r := range radios {
									if r.Checked {
										choice = num
										break
									}
								}

								c <- int16(choice)
								jso.Close()
							}
						})
						jarSelect.AppendChild(choose)
						f.AppendChild(jso)
						w.WriteInt16(<-c)
						close(c)
						if w.Err != nil {
							xjs.SetInnerText(status, w.Err.Error())
						}
					case 255:
						o.Close()
						servers(c)
						return
					default:
						xjs.SetInnerText(status, readError(r).Error())
						return
					}
				}
			}()
		})

		f.AppendChild(nameLabel)
		f.AppendChild(nameInput)
		f.AppendChild(xdom.Br())

		f.AppendChild(urlLabel)
		f.AppendChild(urlInput)
		f.AppendChild(xdom.Br())

		f.AppendChild(uploadLabel)
		f.AppendChild(uploadInput)
		f.AppendChild(xdom.Br())

		f.AppendChild(fileLabel)
		f.AppendChild(fileInput)
		f.AppendChild(xdom.Br())

		f.AppendChild(submit)

		dom.GetWindow().Document().DocumentElement().AppendChild(o)
	}
}

func viewServer(c dom.Element, s Server) func(dom.Event) {
	return func(dom.Event) {
		d := xdom.Div()
		od := overlay.New(d)
		d.AppendChild(xjs.SetInnerText(xdom.H1(), "Server Details"))

		cTabs := []tabs.Tab{
			{"Details", serverDetails(c, od, s)},
		}
		if s.IsRunning() {
			cTabs = append(cTabs, tabs.Tab{"Console", serverConsole(s.ID)})
		}

		d.AppendChild(tabs.MakeTabs(cTabs))

		dom.GetWindow().Document().DocumentElement().AppendChild(od)
	}
}

func serverDetails(c dom.Element, od io.Closer, s Server) func(dom.Element) {
	return func(d dom.Element) {
		go func() {
			m, err := RPC.GetMap(s.Map)
			if err != nil {
				dom.GetWindow().Alert(err.Error())
				return
			}

			nameLabel := xdom.Label()
			xjs.SetInnerText(nameLabel, "Name")
			name := xdom.Input()
			name.Value = s.Name
			name.Type = "text"

			d.AppendChild(nameLabel)
			d.AppendChild(name)
			d.AppendChild(xdom.Br())

			argsLabel := xdom.Label()
			xjs.SetInnerText(argsLabel, "Arguments")

			d.AppendChild(argsLabel)

			argSpans := make([]*dom.HTMLSpanElement, len(s.Args))

			for num, arg := range s.Args {
				a := xdom.Span()
				a.SetAttribute("contenteditable", "true")
				a.SetClass("sizeableInput")
				a.SetTextContent(arg)
				argSpans[num] = a
				d.AppendChild(a)
			}

			remove := xdom.Input()
			remove.Type = "button"
			remove.Value = "-"
			remove.AddEventListener("click", false, func(dom.Event) {
				if len(argSpans) > 0 {
					d.RemoveChild(argSpans[len(argSpans)-1])
					argSpans = argSpans[:len(argSpans)-1]
				}
			})
			add := xdom.Input()
			add.Type = "button"
			add.Value = "+"
			add.AddEventListener("click", false, func(dom.Event) {
				a := xdom.Span()
				a.SetAttribute("contenteditable", "true")
				a.SetClass("sizeableInput")
				argSpans = append(argSpans, a)
				d.InsertBefore(a, remove)
			})

			d.AppendChild(remove)
			d.AppendChild(add)
			d.AppendChild(xdom.Br())
			d.AppendChild(xdom.Br())

			submit := xdom.Input()
			submit.Value = "Make Changes"
			submit.Type = "button"
			submit.AddEventListener("click", false, func(dom.Event) {
				go func() {
					args := make([]string, len(argSpans))
					for num, arg := range argSpans {
						args[num] = arg.TextContent()
					}
					n := name.Value
					err := RPC.SetServer(Server{
						ID:   s.ID,
						Name: n,
						Path: s.Path,
						Args: args,
					})
					if err == nil {
						od.Close()
						servers(c)
						return
					}
					xjs.RemoveChildren(d)
					errDiv := xdom.Div()
					xjs.SetPreText(errDiv, err.Error())
					d.AppendChild(errDiv)
				}()
			})

			d.AppendChild(submit)

			d.AppendChild(xdom.Br())
			d.AppendChild(xjs.SetInnerText(xdom.Label(), "Map"))
			if m.ID < 0 {
				d.AppendChild(xjs.SetInnerText(xdom.Div(), "[Unassigned]"))
			} else {
				d.AppendChild(xjs.SetInnerText(xdom.Div(), m.Name))
			}

		}()
	}
}

func serverConsole(sID int) func(dom.Element) {
	return func(c dom.Element) {
		c.AppendChild(xdom.Textarea())
		input := xdom.Input()
		input.Type = "text"
		input.AddEventListener("keypress", false, func(e dom.Event) {
			ev := e.(*dom.KeyboardEvent)
			if ev.Key == "Enter" {
				dom.GetWindow().Alert(input.Value)
				input.Value = ""
			}
		})
		c.AppendChild(input)
	}
}
