package main

import (
	"errors"
	"image/color"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/MJKWoolnough/byteio"
	"golang.org/x/net/websocket"
)

type paint struct {
	color.Color
	X, Y int32
}

func uploadHandler(conn *websocket.Conn) {
	conn.PayloadType = websocket.BinaryFrame
	r := byteio.StickyReader{Reader: &byteio.LittleEndianReader{conn}}
	w := byteio.StickyWriter{Writer: &byteio.LittleEndianWriter{Writer: conn}}
	uploadType := r.ReadUint8()
	if uploadType > 2 {
		writeError(&w, ErrInvalidType)
		return
	}
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
	if uploadType == 3 {
		url := make([]byte, length)
		r.Read(url)
		if r.Err != nil {
			writeError(&w, r.Err)
			return
		}
		resp, err := http.Get(string(url))
		if err != nil {
			writeError(&w, err)
			return
		}
		_, err = io.Copy(f, resp.Body)
		resp.Body.Close()
		if err != nil {
			writeError(&w, err)
			return
		}
	} else {
		n, err := io.Copy(f, io.LimitReader(conn, int64(length)))
		if err != nil {
			writeError(&w, err)
			return
		}
		if n != length {
			writeError(&w, io.EOF)
			return
		}
	}
	f.Seek(0, 0)
	switch uploadType {
	case 0:
		err = generate(f, &r, &w)
	case 1, 2:
		err = unpack(f, &r, &w)
	}
	if err != nil {
		writeError(&w, err)
		return
	}
	if r.Err != nil {
		writeError(&w, r.Err)
		return
	}
	if w.Err != nil {
		writeError(&w, w.Err)
		return
	}
	w.WriteUint8(255)
}

var ErrInvalidType = errors.New("invalid upload type")
