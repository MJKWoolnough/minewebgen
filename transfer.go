package main

import (
	"errors"
	"io"
	"net/http"
	"os"

	"golang.org/x/net/websocket"
	"vimagination.zapto.org/byteio"
	"vimagination.zapto.org/minewebgen/internal/data"
)

var transferFuncs = [...]func(Transfer, string, *byteio.StickyLittleEndianReader, *byteio.StickyLittleEndianWriter, *os.File, int64) error{
	Transfer.server,
	Transfer.maps,
	Transfer.generate,
	Transfer.generator,
}

type downloadProgress struct {
	io.Reader
	*byteio.StickyLittleEndianWriter
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
	r := byteio.StickyLittleEndianReader{Reader: conn}
	w := byteio.StickyLittleEndianWriter{Writer: conn}

	err := t.handle(&r, &w)
	if err != nil {
		writeError(&w, err)
	}
}

func (t Transfer) handle(r *byteio.StickyLittleEndianReader, w *byteio.StickyLittleEndianWriter) error {
	transferType := r.ReadUint8()

	if transferType>>1 > uint8(len(transferFuncs)) {
		return errors.New("invalid transfer type")
	}

	f, err := os.CreateTemp("", "mineWebGen")
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
