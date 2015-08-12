package main

import (
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"
)

var jrpc *rpc.Client

func rpcInit() error {
	conn, err := websocket.Dial("ws://" + js.Global.Get("location").Get("host").String() + "/rpc")
	if err != nil {
		return err
	}
	closeOnExit(conn)
	jrpc = jsonrpc.NewClient(conn)
	return nil
}

func ServerName() (string, error) {
	var name string
	err := jrpc.Call("Server.Name", nil, &name)
	return name, err
}

type Server struct {
	Name string
}

func ServerList() ([]Server, error) {
	var list []Server
	err := jrpc.Call("Server.List", nil, &list)
	return list, err
}
