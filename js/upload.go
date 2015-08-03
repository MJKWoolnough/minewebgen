package main

import (
	"image/color"
	"io"

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
			uploadDiv.AppendChild(pb.HTMLCanvasElement)
			go func() {
				conn, err := websocket.Dial("ws://" + js.Global.Get("location").Get("host").String() + "/socket")
				if err != nil {
					xjs.SetInnerText(status, err.Error())
					return
				}
				defer conn.Close()
				ew := byteio.LittleEndianWriter{Writer: conn}
				er := byteio.LittleEndianReader{conn}
				_, err = ew.WriteInt64(int64(length))
				if err != nil {
					xjs.SetInnerText(status, err.Error())
					return
				}
				_, err = io.Copy(conn, pb.Reader(files.NewFileReader(file), file.Size()))
				if err != nil {
					xjs.SetInnerText(status, err.Error())
					return
				}
				statusCode, _, err := er.ReadUint8()
				if err != nil {
					xjs.SetInnerText(status, err.Error())
					return
				}
				switch statusCode {
				case 0:
					length, _, err := er.ReadInt64()
					if err != nil {
						xjs.SetInnerText(status, err.Error())
						return
					}
					errStr := make([]byte, length)
					_, err = io.ReadFull(conn, errStr)
					if err != nil {
						xjs.SetInnerText(status, err.Error())
						return
					}
					xjs.SetInnerText(status, string(errStr))
					return
				case 1:
					xjs.SetInnerText(status, "Done")
				}
			}()
		})
		uploadDiv.AppendChild(upl)
	}
	c.AppendChild(uploadDiv)
}
