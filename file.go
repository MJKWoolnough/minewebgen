package main

import (
	"net/http"
	"time"

	"github.com/MJKWoolnough/httpdir"
)

var (
	dir                  = httpdir.New(time.Now())
	hdir http.FileSystem = dir
)
