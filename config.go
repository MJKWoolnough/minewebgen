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
	Servers    []Server
	Maps       []Map
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
		Servers:    make([]Server, 0),
		Maps:       make([]Map, 0),
		selected:   -1,
	}
	err = json.NewDecoder(f).Decode(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
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
	id := len(c.Servers)
	c.Servers = append(c.Servers, Server{ID: id, Name: name, Path: path})
	return id
}

func (c *Config) newMap(name, path string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := len(c.Maps)
	c.Maps = append(c.Maps, Map{ID: id, Name: name, Path: path})
	return id
}

func (c *Config) serverStatus(id int, status string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Servers[id].status = status
}
