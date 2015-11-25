package main

import (
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/MJKWoolnough/minewebgen/internal/data"
)

type Generators struct {
	mu   sync.RWMutex
	List []*data.Generator
}

func (gs *Generators) New(path string) *data.Generator {
	gPath := freePath(path)
	if gPath == "" {
		return nil
	}
	gs.mu.Lock()
	defer gs.mu.Unlock()
	id := 0
	for _, g := range gs.List {
		if g.ID >= id {
			id = g.ID + 1
		}
	}
	g := &data.Generator{
		ID:   id,
		Path: gPath,
		Name: "New Generator",
	}
	gs.List = append(gs.List, g)
	return g
}

func (gs *Generators) Get(id int) *data.Generator {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	for _, g := range gs.List {
		if g.ID == id {
			return g
		}
	}
	return nil
}

func (gs *Generators) Download(w http.ResponseWriter, r *http.Request) {
	b := path.Base(r.URL.Path)
	if len(b) < 5 || b[len(b)-4:] != ".zip" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	id, err := strconv.Atoi(b[:len(b)-4])
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
	}
	g := gs.Get(id)
	if g == nil {
		http.Error(w, "unknown generator", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	archive(w, g.Path)
}

func (gs *Generators) Remove(id int) error {
	gs.mu.Lock()
	gs.mu.Unlock()
	for n, g := range gs.List {
		if g.ID == id {
			copy(gs.List[n:], gs.List[n+1:])
			gs.List = gs.List[:len(gs.List)-1]
			os.RemoveAll(g.Path)
			break
		}
	}
	return nil
}

func (gs *Generators) Len() int {
	return len(gs.List)
}

func (gs *Generators) Less(i, j int) bool {
	return gs.List[i].ID < gs.List[j].ID
}

func (gs *Generators) Swap(i, j int) {
	gs.List[i], gs.List[j] = gs.List[j], gs.List[i]
}
