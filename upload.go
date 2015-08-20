package main

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/MJKWoolnough/byteio"
	"golang.org/x/net/websocket"
)

var uploadFuncs = [...]func(f *os.File, r *byteio.StickyReader, w *byteio.StickyWriter) error{
	generate,
	unpack,
	unpack,
	setupServer,
	setupServer,
}

func uploadHandler(conn *websocket.Conn) {
	conn.PayloadType = websocket.BinaryFrame
	r := byteio.StickyReader{Reader: &byteio.LittleEndianReader{conn}}
	w := byteio.StickyWriter{Writer: &byteio.LittleEndianWriter{Writer: conn}}
	uploadType := r.ReadUint8()
	if uploadType >= uint8(len(uploadFuncs)) {
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
	if uploadType&1 == 1 {
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
	err = uploadFuncs[uploadType](f, &r, &w)
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
