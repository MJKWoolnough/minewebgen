// +build debug

package main

import "net/http"

func init() {
	hdir = http.Dir("")
}
