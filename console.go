package main

import (
	"errors"
	"io"
	"os"
	"path"
	"time"

	"gopkg.in/fsnotify.v1"

	"github.com/MJKWoolnough/byteio"
	"golang.org/x/net/websocket"
)

type Console struct {
	c *Controller
}

func (c Console) Websocket(conn *websocket.Conn) {
	conn.PayloadType = websocket.BinaryFrame
	r := byteio.StickyReader{Reader: &byteio.LittleEndianReader{Reader: conn}}
	w := byteio.StickyWriter{Writer: &byteio.LittleEndianWriter{Writer: conn}}

	err := c.handle(&r, &w)
	if err != nil {
		writeError(&w, err)
	}
}

var logPaths = []string{
	"logs/server.log",
	"logs/latest.log",
	"server.log",
}

func (c Console) handle(r *byteio.StickyReader, w *byteio.StickyWriter) error {
	id := int(r.ReadInt32())
	if r.Err != nil {
		return r.Err
	}
	s := c.c.c.Server(id)
	if s == nil {
		return ErrUnknownServer
	}
	s.RLock() //Needed? Path never gets changed!
	p := s.Path
	s.RUnlock()

	var (
		f       *os.File
		err     error
		logPath string
	)

	for _, lp := range logPaths {
		logPath = path.Join(p, lp)
		f, err = os.Open(logPath)
		if err == nil {
			break
		}
	}
	if f == nil {
		return errors.New("unable to open log file")
	}
	defer func() {
		f.Close()
	}()

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	logDir := path.Dir(logPath)
	fsw.Add(logDir)
	defer fsw.Remove(logDir)

	pw := partWriter{w}

	io.Copy(pw, f)

	if w.Err != nil {
		return w.Err
	}
	t := time.NewTimer(time.Second * 10)
	for {
		select {
		case ev := <-fsw.Events:
			switch ev.Op {
			case fsnotify.Create:
				if ev.Name == logPath {
					f, err = os.Open(logPath)
					if err != nil {
						return err
					}
				}
			case fsnotify.Write:
				if ev.Name == logPath {
					io.Copy(pw, f)
					if w.Err != nil {
						return w.Err
					}
				}
			case fsnotify.Remove:
				if ev.Name == logPath {
					f.Close()
				}
			}
		case err = <-fsw.Errors:
		case <-t.C:
			w.WriteUint8(2) //ping
			if w.Err != nil {
				return w.Err
			}
		}
		t.Reset(time.Second * 10)
	}
}

type partWriter struct {
	*byteio.StickyWriter
}

func (pw partWriter) Write(p []byte) (int, error) {
	l := len(p)
	for len(p) > 0 {
		b := p
		if len(b) > 65535 {
			b = p[:65535]

		}
		p = p[len(b):]
		pw.WriteUint8(1)
		pw.WriteUint16(uint16(len(b)))
		pw.StickyWriter.Write(b)
	}
	return l, nil
}
