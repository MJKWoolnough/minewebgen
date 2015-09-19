package main

import (
	"errors"
	"math/rand"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/MJKWoolnough/minecraft"
)

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

func (c *Config) Save(s Server, _ *struct{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.save()
	ns, ok := c.Servers[s.ID]
	if !ok {
		return ErrNoServer
	}
	s.Path = ns.Path
	s.status = ns.status
	c.Servers[s.ID] = s
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

type DefaultMap struct {
	Mode               int
	Name               string
	GameMode           int
	Seed               int64
	Structures, Cheats bool
}

func (c *Config) CreateDefaultMap(data DefaultMap, _ *struct{}) error {
	if data.Seed == 0 {
		data.Seed = rand.Int63()
	}
	d, err := setupMapDir()
	if err != nil {
		return err
	}
	p, err := minecraft.NewFilePath(d)
	if err != nil {
		return err
	}
	l, err := minecraft.NewLevel(p)
	if err != nil {
		return err
	}
	l.GameMode(int32(data.GameMode))
	l.LevelName(data.Name)
	switch data.Mode {
	case 0:
		l.Generator(minecraft.DefaultGenerator)
	case 1:
		l.Generator(minecraft.FlatGenerator)
	case 2:
		l.Generator(minecraft.LargeBiomeGenerator)
	case 3:
		l.Generator(minecraft.AmplifiedGenerator)
	case 4:
		l.Generator(minecraft.CustomGenerator)
	}
	l.Seed(data.Seed)
	l.AllowCommands(data.Cheats)
	l.MapFeatures(data.Structures)
	l.Save()
	c.newMap(data.Name, d)
	f, err := os.Create(path.Join(d, "server.properties"))
	if err != nil {
		return err
	}
	defer f.Close()
	m := DefaultSettings()
	m["gamemode"] = strconv.Itoa(data.GameMode)
	if !data.Structures {
		m["generate-structures"] = "false"
	}
	if data.GameMode == 3 {
		m["hardcore"] = "true"
	}
	m["level-name"] = data.Name
	m["level-seed"] = strconv.FormatInt(data.Seed, 10)
	switch data.Mode {
	case 0:
		m["level-type"] = minecraft.DefaultGenerator
	case 1:
		m["level-type"] = minecraft.FlatGenerator
	case 2:
		m["level-type"] = minecraft.LargeBiomeGenerator
	case 3:
		m["level-type"] = minecraft.AmplifiedGenerator
	case 4:
		m["level-type"] = minecraft.CustomGenerator
	case 5:
		m["level-type"] = minecraft.DebugGenerator
	}
	if err := m.WriteTo(f); err != nil {
		return err
	}
	return nil
}

var mapDirLock sync.Mutex

func setupMapDir() (string, error) {
	mapDirLock.Lock()
	defer mapDirLock.Unlock()
	num := 0
	for {
		dir := path.Join(config.MapsDir, strconv.Itoa(num))
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0777)
			if err != nil {
				return "", err
			}
			return dir, nil
		}
		num++
	}
}

type MapServer struct {
	Map, Server int
}

func (c *Config) SetMapServer(ms MapServer, _ *struct{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	m, ok := c.Maps[ms.Map]
	if !ok {
		return ErrNoMap
	}
	s, ok := c.Servers[ms.Server]
	if !ok {
		return ErrNoServer
	}
	m.Server = ms.Server
	s.Map = ms.Map
	c.Maps[ms.Map] = m
	c.Servers[ms.Server] = s
	return c.save()
}

// Errors
var ErrNoMap = errors.New("no map found")
