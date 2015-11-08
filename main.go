package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/signal"

	"github.com/MJKWoolnough/httpdir"
	"golang.org/x/net/websocket"
)

var dir http.FileSystem = httpdir.Default

func main() {
	configFile := flag.String("c", "config.json", "config file")
	flag.Parse()
	c, err := LoadConfig(*configFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rpc.RegisterName("RPC", RPC{c})

	http.Handle("/transfer", websocket.Handler(handleFile))
	http.Handle("/rpc", websocket.Handler(func(conn *websocket.Conn) { jsonrpc.ServeConn(conn) }))
	http.Handle("/", http.FileServer(dir))
	l, err := net.Listen("tcp", c.ServerSettings.ListenAddr)
	if err != nil {
		log.Println(err)
		return
	}

	cc := make(chan struct{})
	go func() {
		log.Println("Server Started")
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, os.Interrupt)
		select {
		case <-sc:
			log.Println("Closing")
		case <-cc:
		}
		signal.Stop(sc)
		close(sc)
		l.Close()
		close(cc)
	}()

	err = http.Serve(l, nil)
	select {
	case <-cc:
	default:
		log.Println(err)
		close(cc)
	}
	<-cc
	// Close all running minecraft servers before closing
}
