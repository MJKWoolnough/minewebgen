package main

import (
	"image/color"
	"io"
	"io/ioutil"
	"os"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/ora"
	"golang.org/x/net/websocket"
)

type paint struct {
	color.Color
	X, Y int32
}

func socketHandler(conn *websocket.Conn) {
	conn.PayloadType = websocket.BinaryFrame
	r := byteio.StickyReader{Reader: &byteio.LittleEndianReader{conn}}
	w := byteio.StickyWriter{Writer: &byteio.LittleEndianWriter{Writer: conn}}
	length := r.ReadInt64()
	if r.Err != nil {
		writeError(&w, r.Err)
		return
	}
	f, err := ioutil.TempFile("", "mineWebGen")
	if err != nil {
		writeError(&w, err)
		return
	}
	defer os.Remove(f.Name())
	defer f.Close()
	n, err := io.Copy(f, io.LimitReader(conn, int64(length)))
	if err != nil {
		writeError(&w, err)
		return
	}
	if n != length {
		writeError(&w, io.EOF)
		return
	}
	f.Seek(0, 0)
	o, err := ora.Open(f, length)
	if err != nil {
		writeError(&w, err)
		return
	}
	if o.Layer("terrain") == nil {
		writeError(&w, layerError{"terrain"})
		return
	}
	if o.Layer("height") == nil {
		writeError(&w, layerError{"height"})
		return
	}
	b := o.Bounds()
	w.WriteUint8(1)
	w.WriteInt32(int32(b.Max.X) >> 4)
	w.WriteInt32(int32(b.Max.Y) >> 4)
	if w.Err != nil {
		writeError(&w, w.Err)
		return
	}
	c := make(chan paint, 1024)
	m := make(chan string, 4)
	e := make(chan error, 1)
	go buildMap(o, c, m, e)
Loop:
	for {
		select {
		case p := <-c:
			w.WriteUint8(1)
			w.WriteInt32(p.X)
			w.WriteInt32(p.Y)
			r, g, b, a := p.RGBA()
			w.WriteUint8(uint8(r >> 8))
			w.WriteUint8(uint8(g >> 8))
			w.WriteUint8(uint8(b >> 8))
			w.WriteUint8(uint8(a >> 8))
		case message := <-m:
			w.WriteUint8(2)
			w.WriteUint16(uint16(len(message)))
			w.Write([]byte(message))
		case err := <-e:
			if err == nil {
				break Loop
			}
			writeError(&w, err)
			return
		}
	}
	w.WriteUint8(255)
}

type layerError struct {
	name string
}

func (l layerError) Error() string {
	return "missing layer: " + l.name
}
