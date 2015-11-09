package main

import (
	"errors"
	"image/color"
	"io"
	"sort"
	"strconv"
	"time"

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

func WrapEvent(f func(...dom.Element), c ...dom.Element) func(dom.Event) {
	return func(dom.Event) {
		go f(c...)
	}
}

func ReadError(r *byteio.StickyReader) error {
	s := ReadString(r)
	if r.Err != nil {
		return r.Err
	}
	return errors.New(s)
}

func WriteString(w *byteio.StickyWriter, s string) {
	w.WriteUint16(uint16(len(s)))
	w.Write([]byte(s))
}

func ReadString(r *byteio.StickyReader) string {
	length := r.ReadUint16()
	str := make([]byte, int(length))
	io.ReadFull(r, str)
	return string(str)
}

func transferFile(typeName, method string, typeID uint8, o *overlay.Overlay) dom.Node {
	name := xform.InputText("name", "")
	url := xform.InputRadio("url", "switch", true)
	upload := xform.InputRadio("upload", "switch", false)
	fileI := xform.InputUpload("")
	urlI := xform.InputURL("", "")
	s := xform.InputSubmit(method)

	name.Required = true

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

	f := xjs.AppendChildren(xdom.Form(), xjs.AppendChildren(xdom.Fieldset(),
		xjs.SetInnerText(xdom.Legend(), method+" "+typeName),
		xform.Label(typeName+" Name", "name"),
		name,
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
	))

	s.AddEventListener("click", false, func(e dom.Event) {
		if name.Value == "" {
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
		name.Disabled = true
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
				w.WriteUint8(typeID << 1)
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
				w.WriteUint8(typeID<<1 | 1)
				w.WriteInt32(int32(l))
				io.Copy(&w, pb.Reader(f, l))
			}

			d.RemoveChild(pb)
			xjs.SetInnerText(status, "Checking File")

			WriteString(&w, name.Value)

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
				case 2:
					// generate map
				case 255:
					uo.Close()
					return
				}
			}

		}()
	})
	return f
}

func editProperties(c dom.Element, name string, id int, rpcGet func(int) (map[string]string, error), rpcSet func(id int, properties map[string]string) error) {
	sp, err := rpcGet(id)
	if err != nil {
		c.AppendChild(xjs.SetInnerText(xdom.Div(), "Failed to get properties: "+err.Error()))
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
		toggle := xform.InputButton("", "-")
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

	add := xform.InputButton("", "Add")
	submit := xform.InputButton("", "Save")
	fs := xjs.AppendChildren(xdom.Fieldset(), xjs.AppendChildren(
		df,
		xjs.SetInnerText(xdom.Legend(), name+" Properties"),
		add,
		submit,
	))

	add.AddEventListener("click", false, func(dom.Event) {
		k := xform.InputSizeable("", "")
		v := xform.InputSizeable("", "")
		toggle := xform.InputButton("", "-")
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
			err := rpcSet(id, props)
			if err != nil {
				xjs.Alert("Error setting "+name+" properties: %s", err)
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
}
