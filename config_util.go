package main

import (
	"archive/zip"
	"io"
	"os"
	"path"
	"strconv"
	"sync"
)

var pathFind sync.Mutex

func freePath(p string) string {
	pathFind.Lock()
	defer pathFind.Unlock()
	for num := 0; num < 10000; num++ {
		dir := path.Join(p, strconv.Itoa(num))
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			os.MkdirAll(dir, 0755)
			return dir
		}
	}
	return ""
}

func archive(w io.Writer, p string) {
	p = path.Clean(p)
	zw := zip.NewWriter(w)
	defer zw.Close()
	paths := []string{p}
	for len(paths) > 0 {
		p := paths[0]
		paths = paths[1:]
		d, err := os.Open(p)
		if err != nil {
			continue
		}
		for {
			fs, err := d.Readdir(1)
			if err != nil {
				break
			}
			fname := path.Join(p, fs[0].Name())
			if fs[0].IsDir() {
				paths = append(paths, fname)
				continue
			}
			if fs[0].Mode()&os.ModeSymlink > 0 {
				continue
			}
			fh, _ := zip.FileInfoHeader(fs[0])
			fh.Name = fname[len(p)+1:]
			fw, err := zw.CreateHeader(fh)
			if err != nil {
				return
			}
			f, err := os.Open(fname)
			if err != nil {
				continue
			}
			_, err = io.Copy(fw, f)
			f.Close()
			if err != nil {
				return
			}
		}
	}
}
