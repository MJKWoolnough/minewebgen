package main

import (
	"image/color"
	"io"
	"strconv"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/gopherjs/files"
	"github.com/MJKWoolnough/gopherjs/progress"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"
	"honnef.co/go/js/dom"
)

var gDiv = xjs.CreateElement("div")

func generate(c dom.Element) {
	if !gDiv.HasChildNodes() {
		upl := xjs.CreateElement("input")
		upl.SetAttribute("name", "file")
		upl.SetAttribute("type", "file")
		upl.AddEventListener("change", false, func(e dom.Event) {
			fs := e.Target().(*dom.HTMLInputElement).Files()
			if len(fs) != 1 {
				return
			}
			file := files.NewFile(fs[0])
			length := file.Len()
			pb := progress.New(color.RGBA{255, 0, 0, 0}, color.RGBA{0, 0, 255, 0}, 400, 50)
			gDiv.RemoveChild(upl)
			status := xjs.CreateElement("div")
			xjs.SetInnerText(status, "Uploading...")
			gDiv.AppendChild(status)
			gDiv.AppendChild(pb)
			addRestart := func() {
				reset := xjs.CreateElement("input")
				reset.SetAttribute("type", "button")
				reset.SetAttribute("value", "Restart")
				reset.AddEventListener("click", false, func(dom.Event) {
					xjs.RemoveChildren(gDiv)
					generate(c)
				})
				gDiv.InsertBefore(reset, gDiv.FirstChild())
			}
			setError := func(err string) {
				xjs.SetInnerText(status, err)
				addRestart()
			}
			go func() {
				conn, err := websocket.Dial("ws://" + js.Global.Get("location").Get("host").String() + "/upload")
				if err != nil {
					setError(err.Error())
					return
				}
				defer removeCloser(closeOnExit(conn))
				defer conn.Close()
				w := byteio.StickyWriter{Writer: &byteio.LittleEndianWriter{Writer: conn}}
				r := byteio.StickyReader{Reader: &byteio.LittleEndianReader{conn}}
				uploadFile(0, pb.Reader(files.NewFileReader(file), length), w)
				if w.Err != nil {
					setError(w.Err.Error())
					return
				}
				statusCode := r.ReadUint8()
				if r.Err != nil {
					setError(r.Err.Error())
					return
				}

				readError := func() {
					length := r.ReadUint16()
					if r.Err != nil {
						setError(r.Err.Error())
						return
					}
					errStr := make([]byte, int(length))
					_, err := io.ReadFull(r.Reader, errStr)
					if err != nil {
						setError(err.Error())
						return
					}
					setError(string(errStr))
				}

				switch statusCode {
				case 0:
					readError()
					return
				case 1:
				default:
					setError("Unknown Status")
					return
				}
				gDiv.RemoveChild(pb)
				width := r.ReadInt32()
				height := r.ReadInt32()
				if r.Err != nil {
					setError(err.Error())
					return
				}
				xjs.SetInnerText(status, strconv.FormatInt(int64(width), 10)+"x"+strconv.FormatInt(int64(height), 10))
				canvas := xjs.CreateElement("canvas").(*dom.HTMLCanvasElement)
				canvas.SetAttribute("width", strconv.FormatInt(int64(width), 10))
				canvas.SetAttribute("height", strconv.FormatInt(int64(height), 10))
				canvas.Style().Set("width", strconv.FormatInt(int64(width*4), 10)+"px")
				canvas.Style().Set("height", strconv.FormatInt(int64(width*4), 10)+"px")
				ctx := canvas.GetContext2d()
				gDiv.AppendChild(canvas)
				for {
					statusCode := r.ReadUint8()
					if r.Err != nil {
						setError(r.Err.Error())
						return
					}
					switch statusCode {
					case 0:
						readError()
						return
					case 1:
						x := r.ReadInt32()
						y := r.ReadInt32()
						red := r.ReadUint8()
						green := r.ReadUint8()
						blue := r.ReadUint8()
						alpha := r.ReadUint8()
						if r.Err != nil {
							setError(r.Err.Error())
							return
						}
						ctx.FillStyle = "rgba(" + strconv.Itoa(int(red)) + ", " + strconv.Itoa(int(green)) + ", " + strconv.Itoa(int(blue)) + ", " + strconv.FormatFloat(float64(alpha)/255, 'f', -1, 32) + ")"
						ctx.FillRect(int(x), int(y), 1, 1)
					case 2:
						length := r.ReadUint16()
						message := make([]byte, length)
						r.Read(message)
						if r.Err != nil {
							setError(r.Err.Error())
							return
						}
						xjs.SetInnerText(status, string(message))
					case 255:
						addRestart()
						xjs.SetInnerText(status, "Done")
						return
					default:
						setError("Unknown Error")
						return
					}
				}
			}()
		})
		gDiv.AppendChild(upl)
	}
	c.AppendChild(gDiv)
}
