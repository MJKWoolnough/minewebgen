package main

import (
	"encoding/json"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/MJKWoolnough/minewebgen/internal/config"
)

type Servers struct {
	mu   sync.RWMutex
	List []*config.Server
}

func (s *Servers) Get(id int) *config.Server {
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

func (s *Servers) New(path string) *config.Server {
	sPath := freePath(path)
	if sPath == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := 0
	for _, ser := range s.List {
		if ser.ID > id {
			id = ser.ID + 1
		}
	}
	ser := &config.Server{
		ID:   id,
		Path: sPath,
		Name: "New Server",
		Args: []string{"-Xmx1024M", "-Xms1024M"},
		Map:  -1,
	}
	s.List = append(s.List, ser)
	return ser
}

func (s *Servers) Remove(id int) *config.Server {
	s.mu.Lock()
	s.mu.Unlock()
	for n, ser := range s.List {
		if ser.ID == id {
			l := len(s.List)
			if l != n {
				s.List[n], s.List[l] = s.List[l], s.List[n]
			}
			s.List = s.List[:l-1]
			return ser
		}
	}
}

type Maps struct {
	mu   sync.RWMutex
	List []*config.Map
}

func (m *Maps) Get(id int) *config.Map {
	for _, maps := range m.List {
		if maps.ID == id {
			return maps
		}
	}
	return nil
}

func (m *Maps) Remove(id int) *config.Map {
	s.mu.Lock()
	s.mu.Unlock()
	for n, mp := range m.List {
		if mp.ID == id {
			l := len(m.List)
			if l != n {
				m.List[n], m.List[l] = m.List[l], m.List[n]
			}
			m.List = m.List[:l-1]
			return mp
		}
	}
}

func (m *Maps) New(path string) *config.Map {
	mPath := freePath(path)
	if mPath == "" {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	id := 0
	for _, mp := range m.List {
		if mp.ID > id {
			id = mp.ID + 1
		}
	}
	mp := &config.Map{
		ID:     id,
		Path:   mPath,
		Name:   "New Map",
		Server: -1,
	}
	m.List = append(m.List, mp)
	return mp
}

type Config struct {
	mu             sync.RWMutex
	ServerSettings config.ServerSettings

	Servers Servers
	Maps    Maps

	filename string
}

func LoadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	c := new(Config)
	c.ServerSettings.ServerName = "MineWebGen Server"
	c.ServerSettings.ListenAddr = ":8080"
	c.ServerSettings.DirServers = "servers"
	c.ServerSettings.DirMaps = "maps"
	c.filename = filename
	err = json.NewDecoder(f).Decode(c)
	if err != nil {
		return nil, err
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
	return json.NewEncoder(f).Encode(f)
}

func (c *Config) Server(id int) *config.Server {
	if id < 0 {
		return nil
	}
	return c.Servers.Get(id)
}

func (c *Config) Map(id int) *config.Map {
	if id < 0 {
		return nil
	}
	return c.Maps.Get(id)
}

func (c *Config) NewServer() *config.Server {
	p := c.Settings().DirServers
	return c.Servers.New(p)
}

func (c *Config) NewMap() *config.Map {
	p := c.Settings().DirMaps
	return c.Maps.New(p)
}

func (c *Config) RemoveServer(id int) *config.Server {
	return c.Servers.Remove(id)
}

func (c *Config) RemoveMap(id int) *config.Map {
	return c.Maps.Remove(id)
}

func (c *Config) Settings() config.ServerSettings {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ServerSettings
}

func (c *Config) SetSettings(s config.ServerSettings) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ServerSettings = s
}
