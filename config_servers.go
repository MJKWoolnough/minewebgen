package main

import (
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/MJKWoolnough/minewebgen/internal/data"
)

type Servers struct {
	mu   sync.RWMutex
	List []*data.Server
}

func (s *Servers) Get(id int) *data.Server {
	for _, ser := range s.List {
		if ser.ID == id {
			return ser
		}
	}
	return nil
}

func (s *Servers) New(path string) *data.Server {
	sPath := freePath(path)
	if sPath == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := 0
	for _, ser := range s.List {
		if ser.ID >= id {
			id = ser.ID + 1
		}
	}
	ser := &data.Server{
		ID:   id,
		Path: sPath,
		Name: "New Server",
		Args: []string{"-Xmx1024M", "-Xms1024M"},
		Map:  -1,
	}
	s.List = append(s.List, ser)
	return ser
}

func (s *Servers) Remove(id int) {
	s.mu.Lock()
	s.mu.Unlock()
	for n, ser := range s.List {
		if ser.ID == id {
			copy(s.List[n:], s.List[n+1:])
			s.List = s.List[:len(s.List)-1]
			os.RemoveAll(ser.Path)
			break
		}
	}
}

func (s *Servers) Download(w http.ResponseWriter, r *http.Request) {
	b := path.Base(r.URL.Path)
	if len(b) < 5 || b[len(b)-4:] != ".zip" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	id, err := strconv.Atoi(b[:len(b)-4])
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
	}
	serv := s.Get(id)
	if serv == nil {
		http.Error(w, "unknown sevrer", http.StatusNotFound)
		return
	}
	serv.Lock()
	defer serv.Unlock()
	if serv.State != data.StateStopped {
		http.Error(w, "server running", http.StatusBadGateway)
		return
	}
	serv.State = data.StateBusy
	serv.Unlock()
	w.Header().Set("Content-Type", "application/zip")
	archive(w, serv.Path)
	serv.Lock()
	serv.State = data.StateStopped
}

func (s *Servers) Len() int {
	return len(s.List)
}

func (s *Servers) Less(i, j int) bool {
	return s.List[i].ID < s.List[j].ID
}

func (s *Servers) Swap(i, j int) {
	s.List[i], s.List[j] = s.List[j], s.List[i]
}
