package main

import (
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"
)

type jrpc struct {
	rpc *rpc.Client
}

type Server struct {
	ID         int
	Name, Path string
	Args       []string
	Map        int
}

func (s Server) IsRunning() bool {
	return false
}

type Map struct {
	ID     int
	Name   string
	Server int
}

var (
	RPC         jrpc
	emptyStruct = &struct{}{}
)

func rpcInit() error {
	conn, err := websocket.Dial("ws://" + js.Global.Get("location").Get("host").String() + "/rpc")
	if err != nil {
		return err
	}
	closeOnExit(conn)
	RPC = jrpc{jsonrpc.NewClient(conn)}
	return nil
}

func (j jrpc) ServerName() (string, error) {
	var name string
	err := j.rpc.Call("RPC.Name", nil, &name)
	return name, err
}

func (j jrpc) ServerList() ([]Server, error) {
	var list []Server
	err := j.rpc.Call("RPC.ServerList", nil, &list)
	return list, err
}

func (j jrpc) GetServer(sID int) (Server, error) {
	if sID < 0 {
		return Server{ID: -1}, nil
	}
	var s Server
	err := j.rpc.Call("RPC.GetServer", sID, &s)
	return s, err
}

func (j jrpc) SetServer(s Server) error {
	return j.rpc.Call("RPC.SetServer", s, emptyStruct)
}

func (j jrpc) MapList() ([]Map, error) {
	var list []Map
	err := j.rpc.Call("RPC.MapList", nil, &list)
	return list, err
}

func (j jrpc) GetMap(mID int) (Map, error) {
	if mID < 0 {
		return Map{ID: -1}, nil
	}
	m := Map{ID: -1}
	err := j.rpc.Call("RPC.GetMap", mID, &m)
	return m, err
}

func (j jrpc) SetMap(m Map) error {
	return j.rpc.Call("RPC.SetMap", m, emptyStruct)
}

type MapServer struct {
	Map, Server int
}

func (j jrpc) SetServerMap(sID, mID int) error {
	return j.rpc.Call("RPC.SetMapServer", MapServer{mID, sID}, emptyStruct)
}

func (j jrpc) RemoveServerMap(mapID int) error {
	return j.rpc.Call("RPC.RemoveMapServer", mapID, emptyStruct)
}

type DefaultMap struct {
	Mode               int
	Name               string
	GameMode           int
	Seed               int64
	Structures, Cheats bool
}

func (j jrpc) CreateDefaultMap(data DefaultMap) error {
	return j.rpc.Call("RPC.CreateDefaultMap", data, emptyStruct)
}

type SuperFlatMap struct {
	DefaultMap
}

func (j jrpc) CreateSuperFlatMap(data SuperFlatMap) error {
	return nil
}

type CustomMap struct {
	DefaultMap
}

func (j jrpc) CreateCustomMap(data CustomMap) error {
	return nil
}
