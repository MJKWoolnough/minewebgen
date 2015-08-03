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
	"golang.org/x/net/websocket"
)

func socketHandler(conn *websocket.Conn) {
	r := byteio.LittleEndianReader{conn}
	length, _, err := r.ReadInt64()
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err := ioutil.TempFile("", "mineWebGen")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.Remove(f.Name())
	_, err = io.Copy(f, io.LimitReader(conn, int64(length)))
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Seek(0, 0)
	io.Copy(os.Stdout, f)
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
