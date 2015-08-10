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
	"runtime"

	"golang.org/x/net/websocket"
)

func main() {
	config := flag.String("-c", "config.json", "config file")
	flag.Parse()

	conf, err := loadConfig(*config)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = rpc.RegisterName("Server", conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	http.Handle("/upload", websocket.Handler(uploadHandler))
	http.Handle("/rpc", websocket.Handler(func(conn *websocket.Conn) { jsonrpc.ServeConn(conn) }))
	http.Handle("/", http.FileServer(dir))
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Port))
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
