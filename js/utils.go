package main

import "honnef.co/go/js/dom"

func WrapEvent(f func(...dom.Element), c ...dom.Element) func(dom.Event) {
	return func(dom.Event) {
		go f(c...)
	}
}
