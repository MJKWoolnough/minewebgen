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

func writeError(w *byteio.LittleEndianWriter, err error) {
	ew := byteio.LittleEndianWriter{Writer: w}
	ew.WriteUint8(0)
	errStr := []byte(err.Error())
	ew.WriteInt64(int64(len(errStr)))
	ew.Write(errStr)
	fmt.Println("error:", err)
}

func socketHandler(conn *websocket.Conn) {
	conn.PayloadType = websocket.BinaryFrame
	r := byteio.LittleEndianReader{conn}
	w := byteio.LittleEndianWriter{Writer: conn}
	length, _, err := r.ReadInt64()
	if err != nil {
		writeError(&w, err)
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
	_, err = w.WriteUint8(1)
	if err != nil {
		fmt.Println(err)
		writeError(&w, err)
		return
	}
	b := o.Bounds()
	_, err = w.WriteInt64(int64(b.Max.X))
	if err != nil {
		fmt.Println(err)
		writeError(&w, err)
		return
	}
	_, err = w.WriteInt64(int64(b.Max.Y))
	if err != nil {
		fmt.Println(err)
		writeError(&w, err)
		return
	}
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
