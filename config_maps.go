package main

import (
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"sync"

	"github.com/MJKWoolnough/minewebgen/internal/data"
)

type Maps struct {
	mu   sync.RWMutex
	List []*data.Map
}

func (m *Maps) Get(id int) *data.Map {
	for _, maps := range m.List {
		if maps.ID == id {
			return maps
		}
	}
	return nil
}

func (m *Maps) Remove(id int) {
	m.mu.Lock()
	m.mu.Unlock()
	for n, mp := range m.List {
		if mp.ID == id {
			copy(m.List[n:], m.List[n+1:])
			m.List = m.List[:len(m.List)-1]
			os.RemoveAll(mp.Path)
			break
		}
	}
}

func (m *Maps) Download(w http.ResponseWriter, r *http.Request) {
	b := path.Base(r.URL.Path)
	if len(b) < 5 || b[len(b)-4:] != ".zip" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	id, err := strconv.Atoi(b[:len(b)-4])
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
	}
	mp := m.Get(id)
	if mp == nil {
		http.Error(w, "unknown map", http.StatusNotFound)
		return
	}
	mp.Lock()
	defer mp.Unlock()
	if mp.Server != -1 {
		http.Error(w, "server attached", http.StatusBadGateway)
		return
	}
	mp.Server = -2
	mp.Unlock()
	w.Header().Set("Content-Type", "application/zip")
	archive(w, mp.Path)
	mp.Lock()
	mp.Server = -1
}

func (m *Maps) New(path string) *data.Map {
	mPath := freePath(path)
	if mPath == "" {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	id := len(m.List)
	for n, mp := range m.List {
		if n != mp.ID {
			id = n
			break
		}
	}
	mp := &data.Map{
		ID:     id,
		Path:   mPath,
		Name:   "New Map",
		Server: -1,
	}
	m.List = append(m.List, mp)
	sort.Sort(m)
	return mp
}

func (m *Maps) Len() int {
	return len(m.List)
}

func (m *Maps) Less(i, j int) bool {
	return m.List[i].ID < m.List[j].ID
}

func (m *Maps) Swap(i, j int) {
	m.List[i], m.List[j] = m.List[j], m.List[i]
}
