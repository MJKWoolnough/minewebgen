package main

import (
	"encoding/json"
	"os"
	"sync"
)

type Server struct {
	ID         int
	Name, Path string
	Args       []string
	status     string
	State      int `json:",omitempty"`
	Map        int
}

type Map struct {
	ID         int
	Name, Path string
	Status     string
	Server     int
}

type Config struct {
	mu         sync.RWMutex
	filename   string
	ServerName string
	ServersDir string
	MapsDir    string
	Port       uint16
	Servers    serverMap
	Maps       mapMap
	selected   int
}

type serverMap map[int]*Server

func (m serverMap) MarshalJSON() ([]byte, error) {
	s := make([]*Server, 0, len(m))
	for _, v := range m {
		v.State = 0
		s = append(s, v)
	}
	return json.Marshal(s)
}

func (m serverMap) UnmarshalJSON(b []byte) error {
	var s []*Server
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	for _, v := range s {
		m[v.ID] = v
	}
	return nil
}

type mapMap map[int]*Map

func (m mapMap) MarshalJSON() ([]byte, error) {
	s := make([]*Map, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}
	return json.Marshal(s)
}

func (m mapMap) UnmarshalJSON(b []byte) error {
	var s []*Map
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	for _, v := range s {
		m[v.ID] = v
	}
	return nil
}

func loadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	c := &Config{
		filename:   filename,
		ServerName: "Minecraft",
		Port:       8080,
		Servers:    make(serverMap),
		Maps:       make(mapMap),
		selected:   -1,
	}
	err = json.NewDecoder(f).Decode(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) save() error {
	f, err := os.Create(c.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(c)
}

func (c *Config) createServer(name, path string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.save()
	id := 0
	for {
		_, ok := c.Servers[id]
		if !ok {
			break
		}
		id++
	}
	c.Servers[id] = &Server{ID: id, Name: name, Path: path, Args: []string{"-Xmx1024M", "-Xms1024M"}, Map: -1}
	return id
}

func (c *Config) newMap(name, path string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.save()
	id := 0
	for {
		_, ok := c.Maps[id]
		if !ok {
			break
		}
		id++
	}
	c.Maps[id] = &Map{ID: id, Name: name, Path: path, Server: -1}
	return id
}

func (c *Config) serverStatus(id int, status string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s, ok := c.Servers[id]
	if !ok {
		return
	}
	s.status = status
	//c.Servers[id] = s
}
