package main

import (
	"archive/zip"
	"encoding/json"
	"io"
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
	id := 0
	for _, mp := range m.List {
		if mp.ID >= id {
			id = mp.ID + 1
		}
	}
	mp := &data.Map{
		ID:     id,
		Path:   mPath,
		Name:   "New Map",
		Server: -1,
	}
	m.List = append(m.List, mp)
	return mp
}

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

type Config struct {
	mu             sync.RWMutex
	ServerSettings data.ServerSettings

	Servers Servers
	Maps    Maps

	Generators Generators

	filename string
}

func LoadConfig(filename string) (*Config, error) {
	c := new(Config)
	c.ServerSettings.ServerName = "MineWebGen"
	c.ServerSettings.ListenAddr = ":8080"
	c.ServerSettings.DirServers = "servers"
	c.ServerSettings.DirMaps = "maps"
	c.ServerSettings.DirGenerators = "generators"
	c.ServerSettings.GeneratorPath = "./generator"
	c.filename = filename
	f, err := os.Open(filename)
	if err == nil {
		defer f.Close()
		err = json.NewDecoder(f).Decode(c)
	}
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return c, nil
}

func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.Servers.mu.RLock()
	defer c.Servers.mu.RUnlock()
	c.Maps.mu.RLock()
	defer c.Maps.mu.RUnlock()
	f, err := os.Create(c.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(c)
}

func (c *Config) Server(id int) *data.Server {
	if id < 0 {
		return nil
	}
	return c.Servers.Get(id)
}

func (c *Config) Map(id int) *data.Map {
	if id < 0 {
		return nil
	}
	return c.Maps.Get(id)
}

func (c *Config) NewServer() *data.Server {
	p := c.Settings().DirServers
	return c.Servers.New(p)
}

func (c *Config) NewMap() *data.Map {
	p := c.Settings().DirMaps
	return c.Maps.New(p)
}

func (c *Config) RemoveServer(id int) {
	c.Servers.Remove(id)
}

func (c *Config) RemoveMap(id int) {
	c.Maps.Remove(id)
}

func (c *Config) Settings() data.ServerSettings {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ServerSettings
}

func (c *Config) SetSettings(s data.ServerSettings) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ServerSettings = s
}

func (c *Config) Generator(id int) *data.Generator {
	if id < 0 {
		return nil
	}
	return c.Generators.Get(id)
}

func (c *Config) NewGenerator() *data.Generator {
	p := c.Settings().DirGenerators
	return c.Generators.New(p)
}

func (c *Config) RemoveGenerator(id int) {
	c.Generators.Remove(id)
}
