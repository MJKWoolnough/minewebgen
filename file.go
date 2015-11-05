package main

import (
	"net/http"

	"github.com/MJKWoolnough/httpdir"
)

var dir http.FileSystem = httpdir.Default
