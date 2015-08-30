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
}

type Map struct {
	ID         int
	Name, Path string
	Status     string
}

type Config struct {
	mu         sync.RWMutex
	filename   string
	ServerName string
	ServersDir string
	Port       uint16
	Servers    map[int]Server
	Maps       map[int]Map
	selected   int
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
		Servers:    make(map[int]Server),
		Maps:       make(map[int]Map),
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

func (c *Config) Name(_ struct{}, serverName *string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	*serverName = c.ServerName
	return nil
}

func (c *Config) List(_ struct{}, list *[]Server) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	*list = make([]Server, 0, len(c.Servers))
	for _, s := range c.Servers {
		*list = append(*list, s)
	}
	return nil
}

func (c *Config) MapList(_ struct{}, list *[]Map) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	*list = make([]Map, 0, len(c.Maps))
	for _, m := range c.Maps {
		*list = append(*list, m)
	}
	return nil
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
	c.Servers[id] = Server{ID: id, Name: name, Path: path}
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
	c.Maps[id] = Map{ID: id, Name: name, Path: path}
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
	c.Servers[id] = s
}
