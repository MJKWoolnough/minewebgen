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
			uploadDiv.AppendChild(xjs.CreateElement("br"))
			uploadDiv.AppendChild(pb.HTMLCanvasElement)
			go func() {
				w, err := websocket.Dial("ws://" + js.Global.Get("location").Get("host").String() + "/socket")
				if err != nil {
					println(err)
					return
				}
				defer w.Close()
				er := byteio.LittleEndianWriter{Writer: w}
				_, err = er.WriteInt64(int64(length))
				if err != nil {
					println(err)
					return
				}
				_, err = io.Copy(w, pb.Reader(files.NewFileReader(file), file.Size()))
				if err != nil {
					println(err)
					return
				}
				println("done")
			}()
		})
		uploadDiv.AppendChild(upl)
	}
	c.AppendChild(uploadDiv)
}
