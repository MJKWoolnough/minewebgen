package main

import (
	"strings"

	"github.com/gopherjs/gopherjs/js"

	"honnef.co/go/js/dom"
)

func Fragment() dom.Node {
	return dom.WrapNode(js.Global.Get("document").Call("createDocumentFragment"))
}

func RemoveChildren(node dom.Node) dom.Node {
	for node.HasChildNodes() {
		node.RemoveChild(node.LastChild())
	}
	return node
}

func SetInnerText(node dom.Node, text string) dom.Node {
	RemoveChildren(node)
	node.AppendChild(dom.GetWindow().Document().CreateTextNode(text))
	return node
}

func SetPreText(node dom.Node, text string) dom.Node {
	RemoveChildren(node)
	for n, part := range strings.Split(text, "\n") {
		if n > 0 {
			node.AppendChild(CreateElement("br"))
		}
		node.AppendChild(dom.GetWindow().Document().CreateTextNode(part))
	}
	return node
}

func CreateElement(name string) dom.Element {
	return dom.GetWindow().Document().CreateElement(name)
}

type Tab struct {
	Name string
	Func func(dom.Element)
}

func MakeTabs(t []Tab) dom.Node {
	f := Fragment()
	if len(t) < 0 {
		return f
	}
	tabsDiv := CreateElement("div")
	contents := CreateElement("div")
	tabsDiv.Class().SetString("tabs")
	contents.Class().SetString("content")
	tabs := make([]dom.Element, len(t))
	for n := range t {
		func() {
			i := n
			tabs[n] = SetInnerText(CreateElement("div"), t[i].Name).(dom.Element)
			tabs[n].AddEventListener("click", false, func(e dom.Event) {
				if tabs[i].Class().String() == "selected" {
					return
				}
				for _, tab := range tabs {
					tab.Class().Remove("selected")
				}
				RemoveChildren(contents)
				tabs[i].Class().Add("selected")
				t[i].Func(contents)
			})
			tabsDiv.AppendChild(tabs[n])
		}()
	}
	t[0].Func(contents)
	tabs[0].Class().Add("selected")
	f.AppendChild(tabsDiv)
	f.AppendChild(contents)
	return f
}

func maps(c dom.Element) {
	SetInnerText(c, "MAPS")
}

func upload(c dom.Element) {
	upl := CreateElement("input")
	upl.SetAttribute("name", "file")
	upl.SetAttribute("type", "file")
	c.AppendChild(upl)
	upl.AddEventListener("change", false, func(e dom.Event) {
		files := e.Target().(*dom.HTMLInputElement).Files()
		if len(files) != 1 {
			return
		}
		println(files[0].Get("name"))
	})
}

func main() {
	dom.GetWindow().AddEventListener("load", false, func(e dom.Event) {
		body := dom.GetWindow().Document().(dom.HTMLDocument).Body()
		RemoveChildren(body)
		body.AppendChild(SetInnerText(CreateElement("h1"), "Ferrumwood Server"))
		body.AppendChild(MakeTabs([]Tab{
			{"Maps", maps},
			{"Upload", upload},
		}))
	})
}
