package main

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/minewebgen/internal/data"
	"golang.org/x/net/websocket"
)

var transferFuncs = [...]func(Transfer, string, *byteio.StickyReader, *byteio.StickyWriter, *os.File, int64) error{
	Transfer.server,
	Transfer.maps,
	Transfer.generate,
	Transfer.generator,
}

type downloadProgress struct {
	io.Reader
	*byteio.StickyWriter
}

func (d downloadProgress) Read(b []byte) (int, error) {
	n, err := d.Reader.Read(b)
	d.WriteUint8(1)
	d.WriteInt32(int32(n))
	return n, err
}

type Transfer struct {
	c *Config
}

func (t Transfer) Websocket(conn *websocket.Conn) {
	conn.PayloadType = websocket.BinaryFrame
	r := byteio.StickyReader{Reader: &byteio.LittleEndianReader{Reader: conn}}
	w := byteio.StickyWriter{Writer: &byteio.LittleEndianWriter{Writer: conn}}

	err := t.handle(&r, &w)
	if err != nil {
		writeError(&w, err)
	}
}

func (t Transfer) handle(r *byteio.StickyReader, w *byteio.StickyWriter) error {
	transferType := r.ReadUint8()

	if transferType>>1 > uint8(len(transferFuncs)) {
		return errors.New("invalid transfer type")
	}

	f, err := ioutil.TempFile("", "mineWebGen")
	if err != nil {
		return err
	}
	var size int64
	if transferType&1 == 0 {
		url := data.ReadString(r)
		if r.Err != nil {
			return err
		}
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		w.WriteInt32(int32(resp.ContentLength))
		size, err = io.Copy(f, downloadProgress{resp.Body, w})
		resp.Body.Close()
		if err != nil {
			return err
		}
	} else {
		length := r.ReadInt32()
		_, err := io.Copy(f, io.LimitReader(r, int64(length)))
		if err != nil {
			return err
		}
		size = int64(length)
	}
	f.Seek(0, 0)
	name := data.ReadString(r)
	if r.Err != nil {
		return r.Err
	}
	err = transferFuncs[transferType>>1](t, name, r, w, f, size)
	if err != nil {
		return err
	}
	w.WriteUint8(255)
	return nil
}
