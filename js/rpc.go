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
	ID         int
	Name, Path string
	Args       []string
}

func ServerList() ([]Server, error) {
	var list []Server
	err := jrpc.Call("Server.List", nil, &list)
	return list, err
}

var emptyStruct = &struct{}{}

func SaveServer(s Server) error {
	return jrpc.Call("Server.Save", s, emptyStruct)
}

type Map struct {
	Name string
}

func MapList() ([]Map, error) {
	var list []Map
	err := jrpc.Call("Server.MapList", nil, &list)
	return list, err
}
