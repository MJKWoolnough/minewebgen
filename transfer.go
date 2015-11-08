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

var transferFuncs = [...]func(Transfer, *byteio.StickyReader, *byteio.StickyWriter, *os.File) error{
	Transfer.server,
	Transfer.maps,
	Transfer.generate,
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
	r := byteio.StickyReader{Reader: &byteio.LittleEndianReader{conn}}
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

	if transferType&1 == 0 {
		url := readString(r)
		if r.Err != nil {
			return err
		}
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		w.WriteInt64(resp.ContentLength)
		_, err = io.Copy(f, downloadProgress{resp.Body, w})
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
	}
	f.Seek(0, 0)
	err = transferFuncs[transferType>>1](t, r, w, f)
	if err != nil {
		return err
	}
	w.WriteUint8(255)
	return nil
}

func (Transfer) maps(*byteio.StickyReader, *byteio.StickyWriter, *os.File) error {
	return nil
}

func (Transfer) generate(*byteio.StickyReader, *byteio.StickyWriter, *os.File) error {
	return nil
}
