package main

import (
	"encoding/json"
	"os"
	"sort"
	"sync"

	"github.com/MJKWoolnough/minewebgen/internal/data"
)

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
	c.ServerSettings.GeneratorExecutable = "generator"
	c.ServerSettings.GeneratorMaxMem = 512 * 1024 * 1024
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
	sort.Sort(&c.Servers)
	sort.Sort(&c.Maps)
	sort.Sort(&c.Generators)
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
