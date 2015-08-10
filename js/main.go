package main

import (
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"

	"honnef.co/go/js/dom"
)

func closeOnExit(conn *websocket.Conn) func(*js.Object) {
	return dom.GetWindow().AddEventListener("beforeunload", false, func(_ dom.Event) {
		switch conn.ReadyState {
		case websocket.Connecting, websocket.Open:
			conn.Close()
		}
	})
}

func removeCloser(l func(*js.Object)) {
	dom.GetWindow().RemoveEventListener("beforeunload", false, l)
}

var jrpc *rpc.Client

func main() {
	dom.GetWindow().AddEventListener("load", false, func(_ dom.Event) {
		go func() {
			conn, err := websocket.Dial("ws://" + js.Global.Get("location").Get("host").String() + "/rpc")
			if err != nil {
				dom.GetWindow().Alert("Error connection to RPC server: " + err.Error())
				return
			}
			closeOnExit(conn)
			jrpc = jsonrpc.NewClient(conn)
			body := dom.GetWindow().Document().(dom.HTMLDocument).Body()
			xjs.RemoveChildren(body)
			var title string
			err = jrpc.Call("Server.Name", nil, &title)
			if err != nil {
				dom.GetWindow().Alert("Error connection to RPC server: " + err.Error())
				return
			}
			body.AppendChild(xjs.SetInnerText(xjs.CreateElement("h1"), title+" Server"))
			body.AppendChild(tabs.MakeTabs([]tabs.Tab{
				{"Maps", maps},
				{"Add", add},
			}))
		}()
	})
}
