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
	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"
	"honnef.co/go/js/dom"

	"github.com/MJKWoolnough/gopherjs/overlay"
)

func servers(c dom.Element) {
	xjs.RemoveChildren(c)
	serversDiv := xjs.CreateElement("div")
	defer c.AppendChild(serversDiv)
	list, err := RPC.ServerList()
	if err != nil {
		xjs.SetInnerText(serversDiv, err.Error())
		return
	}
	newButton := xjs.CreateElement("input")
	newButton.SetAttribute("value", "New Server")
	newButton.SetAttribute("type", "button")
	newButton.AddEventListener("click", false, newServer(c))
	c.AppendChild(newButton)
	table := xjs.CreateElement("table")
	head := xjs.CreateElement("tr")
	head.AppendChild(xjs.SetInnerText(xjs.CreateElement("th"), "Name"))
	head.AppendChild(xjs.SetInnerText(xjs.CreateElement("th"), "Status"))
	head.AppendChild(xjs.SetInnerText(xjs.CreateElement("th"), "Controls"))
	table.AppendChild(head)
	for _, s := range list {
		tr := xjs.CreateElement("tr")
		name := xjs.CreateElement("td")
		xjs.SetInnerText(name, s.Name)
		name.AddEventListener("click", false, viewServer(c, s))
		tr.AppendChild(name)
		status := xjs.CreateElement("td")
		xjs.SetInnerText(status, "")
		tr.AppendChild(status)
		controls := xjs.CreateElement("td")
		tr.AppendChild(controls)
		if s.Map >= 0 {
			b := xjs.CreateElement("input").(*dom.HTMLInputElement)
			b.Type = "button"
			if s.IsRunning() {
				b.Value = "Stop"
				b.AddEventListener("click", false, startServer(c, s))
			} else {
				b.Value = "Start"
				b.AddEventListener("click", false, stopServer(c, s))
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
		RPC.ServerStart(s.ID)
	}
}

func stopServer(c dom.Element, s Server) func(dom.Event) {
	return func(dom.Event) {
		RPC.ServerStop(s.ID)
	}
}

func newServer(c dom.Element) func(dom.Event) {
	return func(dom.Event) {
		f := xjs.CreateElement("div")
		o := overlay.New(f)
		f.SetAttribute("id", "serverUpload")

		f.AppendChild(xjs.SetInnerText(xjs.CreateElement("h1"), "New Server"))

		nameLabel := xjs.CreateElement("label")
		nameLabel.SetAttribute("for", "name")
		xjs.SetInnerText(nameLabel, "Level Name")
		nameInput := xjs.CreateElement("input").(*dom.HTMLInputElement)
		nameInput.SetAttribute("type", "text")
		nameInput.SetID("name")

		urlLabel := xjs.CreateElement("label")
		urlLabel.SetAttribute("for", "url")
		xjs.SetInnerText(urlLabel, "URL")
		urlInput := xjs.CreateElement("input")
		urlInput.SetAttribute("type", "radio")
		urlInput.SetAttribute("name", "type")
		urlInput.SetID("url")
		urlInput.SetAttribute("checked", "true")

		uploadLabel := xjs.CreateElement("label")
		uploadLabel.SetAttribute("for", "upload")
		xjs.SetInnerText(uploadLabel, "Upload")
		uploadInput := xjs.CreateElement("input")
		uploadInput.SetAttribute("type", "radio")
		uploadInput.SetAttribute("name", "type")
		uploadInput.SetID("upload")

		fileLabel := xjs.CreateElement("label")
		fileLabel.SetAttribute("for", "file")
		xjs.SetInnerText(fileLabel, "File")
		fileInput := xjs.CreateElement("input").(*dom.HTMLInputElement)
		fileInput.SetAttribute("type", "text")
		fileInput.SetID("file")

		urlInput.AddEventListener("click", false, func(dom.Event) {
			fileInput.SetAttribute("type", "text")
		})

		uploadInput.AddEventListener("click", false, func(dom.Event) {
			fileInput.SetAttribute("type", "file")
		})

		submit := xjs.CreateElement("input")
		submit.SetAttribute("value", "Submit")
		submit.SetAttribute("type", "button")

		submit.AddEventListener("click", false, func(e dom.Event) {
			name := nameInput.Value
			if len(name) == 0 {
				dom.GetWindow().Alert("Name cannot be empty")
				return
			}
			var file readLener
			uploadType := uint8(3)
			if fileInput.GetAttribute("type") == "file" {
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
			status := xjs.CreateElement("div")
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

						jarSelect := xjs.CreateElement("div")
						jso := overlay.New(jarSelect)
						selected := false
						jso.OnClose(func() {
							if !selected {
								selected = true
								c <- -1
							}
						})

						jarSelect.AppendChild(xjs.SetInnerText(xjs.CreateElement("h1"), "Select Server JAR"))
						radios := make([]*dom.HTMLInputElement, numJars)

						for num, name := range jars {
							r := xjs.CreateElement("input").(*dom.HTMLInputElement)
							r.SetAttribute("type", "radio")
							r.SetAttribute("name", "jarChoose")
							v := strconv.Itoa(num)
							r.SetAttribute("value", v)
							r.SetID("jarChoose_" + v)
							if num == 0 {
								r.DefaultChecked = true
							}

							l := xjs.CreateElement("label")
							xjs.SetInnerText(l, name)
							l.SetAttribute("for", "jarChoose_"+v)

							jarSelect.AppendChild(r)
							jarSelect.AppendChild(l)
							jarSelect.AppendChild(xjs.CreateElement("br"))
							radios[num] = r
						}

						choose := xjs.CreateElement("input")
						choose.SetAttribute("type", "button")
						choose.SetAttribute("value", "Select")
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
		f.AppendChild(xjs.CreateElement("br"))

		f.AppendChild(urlLabel)
		f.AppendChild(urlInput)
		f.AppendChild(xjs.CreateElement("br"))

		f.AppendChild(uploadLabel)
		f.AppendChild(uploadInput)
		f.AppendChild(xjs.CreateElement("br"))

		f.AppendChild(fileLabel)
		f.AppendChild(fileInput)
		f.AppendChild(xjs.CreateElement("br"))

		f.AppendChild(submit)

		dom.GetWindow().Document().DocumentElement().AppendChild(o)
	}
}

func viewServer(c dom.Element, s Server) func(dom.Event) {
	return func(dom.Event) {
		d := xjs.CreateElement("div")
		od := overlay.New(d)
		d.AppendChild(xjs.SetInnerText(xjs.CreateElement("h1"), "Server Details"))

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

			nameLabel := xjs.CreateElement("label")
			xjs.SetInnerText(nameLabel, "Name")
			name := xjs.CreateElement("input").(*dom.HTMLInputElement)
			name.Value = s.Name
			name.Type = "text"

			d.AppendChild(nameLabel)
			d.AppendChild(name)
			d.AppendChild(xjs.CreateElement("br"))

			argsLabel := xjs.CreateElement("label")
			xjs.SetInnerText(argsLabel, "Arguments")

			d.AppendChild(argsLabel)

			argSpans := make([]*dom.HTMLSpanElement, len(s.Args))

			for num, arg := range s.Args {
				a := xjs.CreateElement("span").(*dom.HTMLSpanElement)
				a.SetAttribute("contenteditable", "true")
				a.SetAttribute("class", "sizeableInput")
				a.SetTextContent(arg)
				argSpans[num] = a
				d.AppendChild(a)
			}

			remove := xjs.CreateElement("input").(*dom.HTMLInputElement)
			remove.Type = "button"
			remove.Value = "-"
			remove.AddEventListener("click", false, func(dom.Event) {
				if len(argSpans) > 0 {
					d.RemoveChild(argSpans[len(argSpans)-1])
					argSpans = argSpans[:len(argSpans)-1]
				}
			})
			add := xjs.CreateElement("input").(*dom.HTMLInputElement)
			add.Type = "button"
			add.Value = "+"
			add.AddEventListener("click", false, func(dom.Event) {
				a := xjs.CreateElement("span").(*dom.HTMLSpanElement)
				a.SetAttribute("contenteditable", "true")
				a.SetAttribute("class", "sizeableInput")
				argSpans = append(argSpans, a)
				d.InsertBefore(a, remove)
			})

			d.AppendChild(remove)
			d.AppendChild(add)
			d.AppendChild(xjs.CreateElement("br"))
			d.AppendChild(xjs.CreateElement("br"))

			submit := xjs.CreateElement("input").(*dom.HTMLInputElement)
			submit.Value = "Make Changes"
			submit.SetAttribute("type", "button")
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
					errDiv := xjs.CreateElement("div")
					xjs.SetPreText(errDiv, err.Error())
					d.AppendChild(errDiv)
				}()
			})

			d.AppendChild(submit)

			d.AppendChild(xjs.CreateElement("br"))
			d.AppendChild(xjs.SetInnerText(xjs.CreateElement("label"), "Map"))
			if m.ID < 0 {
				d.AppendChild(xjs.SetInnerText(xjs.CreateElement("div"), "[Unassigned]"))
			} else {
				d.AppendChild(xjs.SetInnerText(xjs.CreateElement("div"), m.Name))
			}

		}()
	}
}

func serverConsole(sID int) func(dom.Element) {
	return func(c dom.Element) {
		c.AppendChild(xjs.CreateElement("textarea"))
		input := xjs.CreateElement("input").(*dom.HTMLInputElement)
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
