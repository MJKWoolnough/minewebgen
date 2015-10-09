package main

import (
	"image/color"
	"strconv"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/gopherjs/files"
	"github.com/MJKWoolnough/gopherjs/progress"
	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"
	"honnef.co/go/js/dom"
)

var gDiv = xdom.Div()

func generate(c dom.Element) {
	if !gDiv.HasChildNodes() {
		nameLabel := xdom.Label()
		nameLabel.For = "name"
		xjs.SetInnerText(nameLabel, "Level Name")
		name := xdom.Input()
		name.Type = "name"
		name.SetID("name")

		uploadLabel := xdom.Label()
		uploadLabel.For = "file"
		xjs.SetInnerText(uploadLabel, "Input File")
		upload := xdom.Input()
		upload.Type = "file"
		upload.SetID("file")

		submit := xdom.Input()
		submit.Type = "button"
		submit.Value = "Generate"
		submit.AddEventListener("click", false, func(dom.Event) {
			nameStr := name.Value
			if len(nameStr) == 0 {
				return
			}
			fs := upload.Files()
			if len(fs) != 1 {
				return
			}
			file := files.NewFile(fs[0])
			length := file.Len()
			pb := progress.New(color.RGBA{255, 0, 0, 0}, color.RGBA{0, 0, 255, 0}, 400, 50)
			gDiv.RemoveChild(upload)
			status := xdom.Div()
			xjs.SetInnerText(status, "Uploading...")
			gDiv.AppendChild(status)
			gDiv.AppendChild(pb)
			addRestart := func() {
				reset := xdom.Input()
				reset.Type = "button"
				reset.Value = "Restart"
				reset.AddEventListener("click", false, func(dom.Event) {
					xjs.RemoveChildren(gDiv)
					generate(c)
				})
				gDiv.InsertBefore(reset, gDiv.FirstChild())
			}
			setError := func(err error) {
				xjs.SetInnerText(status, err.Error())
				addRestart()
			}
			go func() {
				conn, err := websocket.Dial("ws://" + js.Global.Get("location").Get("host").String() + "/upload")
				if err != nil {
					setError(err)
					return
				}
				defer removeCloser(closeOnExit(conn))
				defer conn.Close()
				w := &byteio.StickyWriter{Writer: &byteio.LittleEndianWriter{Writer: conn}}
				r := &byteio.StickyReader{Reader: &byteio.LittleEndianReader{conn}}
				uploadFile(0, pb.Reader(files.NewFileReader(file), length), w)
				writeString(w, nameStr)
				if w.Err != nil {
					setError(w.Err)
					return
				}
				statusCode := r.ReadUint8()
				if r.Err != nil {
					setError(r.Err)
					return
				}

				switch statusCode {
				case 0:
					setError(readError(r))
					return
				case 1:
				default:
					setError(ErrUnknown)
					return
				}
				gDiv.RemoveChild(pb)
				width := r.ReadInt32()
				height := r.ReadInt32()
				if r.Err != nil {
					setError(err)
					return
				}
				xjs.SetInnerText(status, strconv.FormatInt(int64(width), 10)+"x"+strconv.FormatInt(int64(height), 10))
				canvas := xdom.Canvas()
				canvas.Width = int(width)
				canvas.Height = int(height)
				canvas.Style().Set("width", strconv.FormatInt(int64(width*4), 10)+"px")
				canvas.Style().Set("height", strconv.FormatInt(int64(width*4), 10)+"px")
				ctx := canvas.GetContext2d()
				gDiv.AppendChild(canvas)
				for {
					statusCode := r.ReadUint8()
					if r.Err != nil {
						setError(r.Err)
						return
					}
					switch statusCode {
					case 0:
						setError(readError(r))
						return
					case 1:
						x := r.ReadInt32()
						y := r.ReadInt32()
						red := r.ReadUint8()
						green := r.ReadUint8()
						blue := r.ReadUint8()
						alpha := r.ReadUint8()
						if r.Err != nil {
							setError(r.Err)
							return
						}
						ctx.FillStyle = "rgba(" + strconv.Itoa(int(red)) + ", " + strconv.Itoa(int(green)) + ", " + strconv.Itoa(int(blue)) + ", " + strconv.FormatFloat(float64(alpha)/255, 'f', -1, 32) + ")"
						ctx.FillRect(int(x), int(y), 1, 1)
					case 2:
						length := r.ReadUint16()
						message := make([]byte, length)
						r.Read(message)
						if r.Err != nil {
							setError(r.Err)
							return
						}
						xjs.SetInnerText(status, string(message))
					case 255:
						addRestart()
						xjs.SetInnerText(status, "Done")
						return
					default:
						setError(ErrUnknown)
						return
					}
				}
			}()
		})
		gDiv.AppendChild(nameLabel)
		gDiv.AppendChild(name)
		gDiv.AppendChild(xdom.Br())
		gDiv.AppendChild(uploadLabel)
		gDiv.AppendChild(upload)
		gDiv.AppendChild(xdom.Br())
		gDiv.AppendChild(xdom.Br())
		gDiv.AppendChild(submit)
	}
	c.AppendChild(gDiv)
}
