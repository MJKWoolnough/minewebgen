package main

import (
	"encoding/json"
	"os"
	"sync"
)

type Server struct {
	ID         uint32
	Name, Path string
}

type Config struct {
	mu         sync.RWMutex
	filename   string
	ServerName string
	Port       uint16
	Servers    map[uint32]Server
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
		Servers:    make(map[uint32]Server, 0),
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
