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
	Map        int
}

func ServerList() ([]Server, error) {
	var list []Server
	err := jrpc.Call("Server.List", nil, &list)
	return list, err
}

func (s Server) IsRunning() bool {
	return false
}

var emptyStruct = &struct{}{}

func SaveServer(s Server) error {
	return jrpc.Call("Server.Save", s, emptyStruct)
}

type Map struct {
	Name   string
	Server int
}

func MapList() ([]Map, error) {
	var list []Map
	err := jrpc.Call("Server.MapList", nil, &list)
	return list, err
}

type DefaultMap struct {
	Mode               int
	Name               string
	GameMode           int
	Seed               int64
	Structures, Cheats bool
}

func CreateDefaultMap(data DefaultMap) error {
	return jrpc.Call("Server.CreateDefaultMap", data, emptyStruct)
}

type SuperFlatMap struct {
	DefaultMap
}

func CreateSuperFlatMap(data SuperFlatMap) error {
	return nil
}

type CustomMap struct {
	DefaultMap
}

func CreateCustomMap(data CustomMap) error {
	return nil
}
