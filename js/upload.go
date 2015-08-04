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

var uploadDiv = xjs.CreateElement("div")

func upload(c dom.Element) {
	if !uploadDiv.HasChildNodes() {
		upl := xjs.CreateElement("input")
		upl.SetAttribute("name", "file")
		upl.SetAttribute("type", "file")
		upl.AddEventListener("change", false, func(e dom.Event) {
			fs := e.Target().(*dom.HTMLInputElement).Files()
			if len(fs) != 1 {
				return
			}
			file := files.NewFile(fs[0])
			length := file.Size()
			pb := progress.New(color.RGBA{255, 0, 0, 0}, color.RGBA{0, 0, 255, 0}, 400, 50)
			uploadDiv.RemoveChild(upl)
			status := xjs.CreateElement("div")
			xjs.SetInnerText(status, "Uploading...")
			uploadDiv.AppendChild(status)
			uploadDiv.AppendChild(pb)
			go func() {
				conn, err := websocket.Dial("ws://" + js.Global.Get("location").Get("host").String() + "/socket")
				if err != nil {
					xjs.SetInnerText(status, err.Error())
					return
				}
				dom.GetWindow().AddEventListener("beforeunload", false, func(_ dom.Event) {
					conn.Close()
				})
				defer conn.Close()
				w := byteio.StickyWriter{Writer: &byteio.LittleEndianWriter{Writer: conn}}
				r := byteio.StickyReader{Reader: &byteio.LittleEndianReader{conn}}
				w.WriteInt64(int64(length))
				if w.Err != nil {
					xjs.SetInnerText(status, err.Error())
					return
				}
				_, err = io.Copy(conn, pb.Reader(files.NewFileReader(file), file.Size()))
				if err != nil {
					xjs.SetInnerText(status, err.Error())
					return
				}
				statusCode := r.ReadUint8()
				if r.Err != nil {
					xjs.SetInnerText(status, err.Error())
					return
				}
				switch statusCode {
				case 0:
					readError(status, r)
					return
				case 1:
				default:
					xjs.SetInnerText(status, "unknown status")
					return
				}
				uploadDiv.RemoveChild(pb)
				width := r.ReadInt64()
				height := r.ReadInt64()
				if r.Err != nil {
					xjs.SetInnerText(status, err.Error())
					return
				}
				xjs.SetInnerText(status, strconv.FormatInt(width, 10)+"x"+strconv.FormatInt(height, 10))
				return
				canvas := xjs.CreateElement("canvas").(*dom.HTMLCanvasElement)
				canvas.SetAttribute("width", strconv.FormatInt(width, 10))
				canvas.SetAttribute("height", strconv.FormatInt(width, 10))
				ctx := canvas.GetContext2d()
				for {
					statusCode := r.ReadUint8()
					if r.Err != nil {
						xjs.SetInnerText(status, err.Error())
						return
					}
					switch statusCode {
					case 0:
						readError(status, r)
						return
					case 1:
						x := r.ReadInt64()
						y := r.ReadInt64()
						red := r.ReadUint8()
						green := r.ReadUint8()
						blue := r.ReadUint8()
						alpha := r.ReadUint8()
						if r.Err != nil {
							xjs.SetInnerText(status, err.Error())
							return
						}
						ctx.FillStyle = "rgba(" + strconv.Itoa(int(red)) + ", " + strconv.Itoa(int(green)) + ", " + strconv.Itoa(int(blue)) + ", " + strconv.FormatFloat(float64(alpha)/255, 'f', -1, 32) + ")"
						ctx.FillRect(int(x), int(y), 1, 1)
					case 255:
						return
					default:
						xjs.SetInnerText(status, "unknown status")
						return
					}
				}
			}()
		})
		uploadDiv.AppendChild(upl)
	}
	c.AppendChild(uploadDiv)
}

func readError(status dom.Element, r byteio.StickyReader) {
	length := r.ReadInt64()
	if r.Err != nil {
		xjs.SetInnerText(status, r.Err.Error())
		return
	}
	errStr := make([]byte, length)
	_, err := io.ReadFull(r.Reader, errStr)
	if err != nil {
		xjs.SetInnerText(status, err.Error())
		return
	}
	xjs.SetInnerText(status, string(errStr))
}
