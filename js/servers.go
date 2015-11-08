package main

import (
	"image/color"
	"io"
	"strconv"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/gopherjs/files"
	"github.com/MJKWoolnough/gopherjs/overlay"
	"github.com/MJKWoolnough/gopherjs/progress"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xform"
	"github.com/MJKWoolnough/gopherjs/xjs"
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
	for _, serv := range s {
		c.AppendChild(xjs.SetInnerText(xdom.Div(), serv.Name))
	}
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
