package main

import (
	"github.com/MJKWoolnough/gopherjs/style"
	"github.com/MJKWoolnough/gopherjs/tabs"
	"github.com/MJKWoolnough/gopherjs/xjs"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"

	"honnef.co/go/js/dom"
)

const css = `label {
	display : block;
	float : left;
	text-align : right;
	width : 200px;
}

label:after {
	content : ':';
}

.sizeableInput {
	border : 2px inset #DCDAD5;
	display : block;
	float : left;
	padding-left : 3px;
	padding-right : 3px;
	min-width : 50px;
	height : 20px;
	margin-top : 2px;
}
`

func init() {
	style.Add(css)
}

func closeOnExit(conn *websocket.Conn) func(*js.Object) {
	return dom.GetWindow().AddEventListener("beforeunload", false, func(dom.Event) {
		switch conn.ReadyState {
		case websocket.Connecting, websocket.Open:
			conn.Close()
		}
	})
}

func removeCloser(l func(*js.Object)) {
	dom.GetWindow().RemoveEventListener("beforeunload", false, l)
}

func main() {
	dom.GetWindow().AddEventListener("load", false, func(dom.Event) {
		go func() {
			err := rpcInit()
			if err != nil {
				dom.GetWindow().Alert("Error connection to RPC server: " + err.Error())
				return
			}
			body := dom.GetWindow().Document().(dom.HTMLDocument).Body()
			xjs.RemoveChildren(body)
			title, err := ServerName()
			if err != nil {
				dom.GetWindow().Alert(err.Error())
				return
			}
			body.AppendChild(xjs.SetInnerText(xjs.CreateElement("h1"), title+" Server"))
			tDoc, ok := dom.GetWindow().Document().(dom.HTMLDocument)
			if ok {
				tDoc.SetTitle(title + " Server")
			}
			body.AppendChild(tabs.MakeTabs([]tabs.Tab{
				{"Servers", servers},
				{"Maps", maps},
				{"Add", add},
			}))
		}()
	})
}
