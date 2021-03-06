package main // import "vimagination.zapto.org/minewebgen"

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

	"golang.org/x/net/websocket"
	"vimagination.zapto.org/httpdir"
	"vimagination.zapto.org/httpgzip"
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

	controller := NewController(c)
	rpc.RegisterName("RPC", RPC{controller})

	t := Transfer{c}
	con := Console{controller}
	http.Handle("/download/generator/", http.HandlerFunc(c.Generators.Download))
	http.Handle("/download/server/", http.HandlerFunc(c.Servers.Download))
	http.Handle("/download/maps/", http.HandlerFunc(c.Maps.Download))
	http.Handle("/transfer", websocket.Handler(t.Websocket))
	http.Handle("/console", websocket.Handler(con.Websocket))
	http.Handle("/rpc", websocket.Handler(func(conn *websocket.Conn) { jsonrpc.ServeConn(conn) }))
	http.Handle("/", httpgzip.FileServer(dir))
	l, err := net.Listen("tcp", c.ServerSettings.ListenAddr)
	if err != nil {
		fmt.Println(os.Stderr, err)
		os.Exit(3)
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
		close(cc)
		l.Close()
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
	if len(controller.running) > 0 {
		log.Println("Stopping all servers...")
		controller.stopAll()
		log.Println("...servers stopped")
	}
	if len(gp.cmds) > 0 {
		log.Println("Stopping all generators...")
		gp.StopAll()
		log.Println("...generators stopped")
	}
}
