package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/MJKWoolnough/byteio"
	"github.com/MJKWoolnough/ora"
	"golang.org/x/net/websocket"
)

func writeError(w *byteio.StickyWriter, err error) {
	w.WriteUint8(0)
	errStr := []byte(err.Error())
	w.WriteUint16(uint16(len(errStr)))
	w.Writer.Write(errStr)
	fmt.Println("error:", err)
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
	terrain := o.Layer("terrain")
	if terrain == nil {
		writeError(&w, layerError{"terrain"})
		return
	}
	heightMap := o.Layer("height")
	if heightMap == nil {
		writeError(&w, layerError{"height"})
		return
	}
	b := o.Bounds()
	w.WriteUint8(1)
	w.WriteInt32(int32(b.Max.X))
	w.WriteInt32(int32(b.Max.Y))
	if w.Err != nil {
		writeError(&w, w.Err)
		return
	}
	for i := 0; i < 255; i++ {
		for j := 0; j < 10; j++ {
			w.WriteUint8(1)
			w.WriteInt32(int32(i))
			w.WriteInt32(int32(j))
			w.WriteUint8(uint8(i + j*10))
			w.WriteUint8(uint8(i + j*10))
			w.WriteUint8(uint8(i + j*10))
			w.WriteUint8(255)
			if w.Err != nil {
				writeError(&w, w.Err)
				return
			}
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

var port = flag.Uint("-p", 8080, "server port")

func main() {
	http.Handle("/socket", websocket.Handler(socketHandler))
	http.Handle("/", http.FileServer(dir))
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Println(err)
		return
	}

	c := make(chan os.Signal, 1)
	go func() {
		defer l.Close()
		log.Println("Server Started")
		signal.Notify(c, os.Interrupt)
		defer signal.Stop(c)
		<-c
		close(c)
		log.Println("Closing")
	}()

	err = http.Serve(l, nil)
	select {
	case <-c:
	default:
		close(c)
		log.Println(err)
	}
}
