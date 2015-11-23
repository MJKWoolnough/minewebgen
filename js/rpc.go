package main

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"strconv"

	"honnef.co/go/js/dom"

	"github.com/MJKWoolnough/gopherjs/xdom"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/MJKWoolnough/minewebgen/internal/data"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"
)

type jRPC struct {
	rpc *rpc.Client
}

func rpcInit() error {
	conn, err := websocket.Dial("ws://" + js.Global.Get("location").Get("host").String() + "/rpc")
	if err != nil {
		return err
	}
	conn.WebSocket.Call("addEventListener", "close", func(*js.Object) {
		xjs.RemoveChildren(dom.GetWindow().Document().(dom.HTMLDocument).Body()).AppendChild(xjs.SetInnerText(xdom.H1(), "Connection Lost"))
	}, false)
	dom.GetWindow().AddEventListener("beforeunload", false, func(dom.Event) {
		switch conn.ReadyState {
		case websocket.Connecting, websocket.Open:
			conn.Close()
		}
	})
	RPC = jRPC{jsonrpc.NewClient(conn)}
	return nil
}

var (
	RPC jRPC
	es  = &struct{}{}
)

func (j jRPC) Settings() (data.ServerSettings, error) {
	var s data.ServerSettings
	err := j.rpc.Call("RPC.Settings", nil, &s)
	return s, err
}

func (j jRPC) SetSettings(settings data.ServerSettings) error {
	return j.rpc.Call("RPC.SetSettings", settings, es)
}

func (j jRPC) ServerName() (string, error) {
	var name string
	err := j.rpc.Call("RPC.ServerName", nil, &name)
	return name, err
}

func (j jRPC) ServerList() ([]data.Server, error) {
	var list []data.Server
	err := j.rpc.Call("RPC.ServerList", nil, &list)
	return list, err
}

func (j jRPC) MapList() ([]data.Map, error) {
	var list []data.Map
	err := j.rpc.Call("RPC.MapList", nil, &list)
	return list, err
}

func (j jRPC) Server(id int) (data.Server, error) {
	var s data.Server
	err := j.rpc.Call("RPC.Server", id, &s)
	return s, err
}

func (j jRPC) Map(id int) (data.Map, error) {
	var m data.Map
	err := j.rpc.Call("RPC.Map", id, &m)
	return m, err
}

func (j jRPC) SetServer(s data.Server) error {
	return j.rpc.Call("RPC.SetServer", s, es)
}

func (j jRPC) SetMap(m data.Map) error {
	return j.rpc.Call("RPC.SetMap", m, es)
}

func (j jRPC) SetServerMap(serverID, mapID int) error {
	return j.rpc.Call("RPC.SetServerMap", [2]int{serverID, mapID}, es)
}

func (j jRPC) ServerProperties(id int) (map[string]string, error) {
	sp := make(map[string]string)
	err := j.rpc.Call("RPC.ServerProperties", id, &sp)
	return sp, err
}

func (j jRPC) SetServerProperties(id int, properties map[string]string) error {
	return j.rpc.Call("RPC.SetServerProperties", data.ServerProperties{id, properties}, es)
}

func (j jRPC) MapProperties(id int) (map[string]string, error) {
	mp := make(map[string]string)
	err := j.rpc.Call("RPC.MapProperties", id, &mp)
	return mp, err
}

func (j jRPC) SetMapProperties(id int, properties map[string]string) error {
	return j.rpc.Call("RPC.SetMapProperties", data.ServerProperties{id, properties}, es)
}

func (j jRPC) RemoveServer(sid string) error {
	id, err := strconv.Atoi(sid)
	if err != nil {
		return err
	}
	return j.rpc.Call("RPC.RemoveServer", id, es)
}

func (j jRPC) RemoveMap(sid string) error {
	id, err := strconv.Atoi(sid)
	if err != nil {
		return err
	}
	return j.rpc.Call("RPC.RemoveMap", id, es)
}

func (j jRPC) StartServer(id int) error {
	return j.rpc.Call("RPC.StartServer", id, es)
}

func (j jRPC) StopServer(id int) error {
	return j.rpc.Call("RPC.StopServer", id, es)
}

func (j jRPC) CreateDefaultMap(d data.DefaultMap) error {
	return j.rpc.Call("RPC.CreateDefaultMap", d, es)
}

func (j jRPC) CreateSuperflatMap(d data.SuperFlatMap) error {
	return j.rpc.Call("RPC.CreateSuperflatMap", d, es)
}

func (j jRPC) CreateCustomMap(d data.CustomMap) error {
	return j.rpc.Call("RPC.CreateCustomMap", d, es)
}

func (j jRPC) ServerEULA(id int) (string, error) {
	var d string
	err := j.rpc.Call("RPC.ServerEULA", id, &d)
	return d, err
}

func (j jRPC) SetServerEULA(id int, d string) error {
	return j.rpc.Call("RPC.SetServerEULA", data.ServerEULA{ID: id, EULA: d}, es)
}

func (j jRPC) WriteCommand(id int, command string) error {
	return j.rpc.Call("RPC.WriteCmd", data.WriteCmd{ID: id, Cmd: command}, es)
}

func (j jRPC) Generators() ([]string, error) {
	var gs []string
	err := j.rpc.Call("RPC.Generators", nil, &gs)
	return gs, err
}

func (j jRPC) Generator(name string) (data.Generator, error) {
	var g data.Generator
	err := j.rpc.Call("RPC.Generator", name, &g)
	return g, err
}

func (j jRPC) RemoveGenerator(sid string) error {
	return j.rpc.Call("RPC.RemoveGenerator", sid, es)
}
