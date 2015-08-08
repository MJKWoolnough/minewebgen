package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"

	"golang.org/x/net/websocket"
)

var port = flag.Uint("-p", 8080, "server port")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.Handle("/upload", websocket.Handler(uploadHandler))
	http.Handle("/rpc", websocket.Handler(rpcHandler))
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
