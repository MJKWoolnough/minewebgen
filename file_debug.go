// +build debug

package main

import "net/http"

func init() {
	dir = http.Dir("")
}
