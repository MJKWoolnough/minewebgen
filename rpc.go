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

type RPC struct {
	c *Config
}

func (r RPC) Name(_ struct{}, serverName *string) error {
	r.c.mu.RLock()
	defer r.c.mu.RUnlock()
	*serverName = r.c.ServerName
	return nil
}

func (r RPC) ServerList(_ struct{}, list *[]Server) error {
	r.c.mu.RLock()
	defer r.c.mu.RUnlock()
	*list = make([]Server, 0, len(r.c.Servers))
	for _, s := range r.c.Servers {
		*list = append(*list, s)
	}
	return nil
}

func (r RPC) GetServer(sID int, s *Server) error {
	r.c.mu.Lock()
	defer r.c.mu.Unlock()
	ns, ok := r.c.Servers[sID]
	if !ok {
		return ErrNoServer
	}
	*s = ns
	return nil
}

func (r RPC) SetServer(s Server, _ *struct{}) error {
	r.c.mu.Lock()
	defer r.c.mu.Unlock()
	defer r.c.save()
	ns, ok := r.c.Servers[s.ID]
	if !ok {
		return ErrNoServer
	}
	s.Path = ns.Path
	s.status = ns.status
	r.c.Servers[s.ID] = s
	return nil
}

func (r RPC) ServerStart(sID int, _ *struct{}) error {
	controller.Start(sID)
}

func (r RPC) ServerStop(sID int, _ *struct{}) error {
	return controller.Stop(sID)
}

func (r RPC) MapList(_ struct{}, list *[]Map) error {
	r.c.mu.RLock()
	defer r.c.mu.RUnlock()
	*list = make([]Map, 0, len(r.c.Maps))
	for _, m := range r.c.Maps {
		*list = append(*list, m)
	}
	return nil
}

func (r RPC) GetMap(mID int, m *Map) error {
	r.c.mu.Lock()
	defer r.c.mu.Unlock()
	nm, ok := r.c.Maps[mID]
	if !ok {
		return ErrNoMap
	}
	*m = nm
	return nil
}

func (r RPC) SetMap(m Map, _ *struct{}) error {
	r.c.mu.Lock()
	defer r.c.mu.Unlock()
	defer r.c.save()
	nm, ok := r.c.Maps[m.ID]
	if !ok {
		return ErrNoMap
	}
	m.Path = nm.Path
	m.Server = nm.Server
	m.Status = nm.Status
	r.c.Maps[m.ID] = m
	return nil
}

type MapServer struct {
	Map, Server int
}

func (r RPC) SetMapServer(ms MapServer, _ *struct{}) error {
	r.c.mu.Lock()
	defer r.c.mu.Unlock()
	m, ok := r.c.Maps[ms.Map]
	if !ok {
		return ErrNoMap
	}
	if m.Server >= 0 {
		return ErrMapAlreadyAssigned
	}
	s, ok := r.c.Servers[ms.Server]
	if !ok {
		return ErrNoServer
	}
	if s.Map >= 0 {
		return ErrServerAlreadyAssigned
	}
	if s.state != StateStopped {
		return ErrServerRunning
	}
	m.Server = ms.Server
	s.Map = ms.Map
	r.c.Maps[ms.Map] = m
	r.c.Servers[ms.Server] = s
	return r.c.save()
}

func (r RPC) RemoveMapServer(mID int, _ *struct{}) error {
	r.c.mu.Lock()
	defer r.c.mu.Unlock()
	m, ok := r.c.Maps[mID]
	if !ok {
		return ErrNoMap
	}
	s, ok := r.c.Servers[m.Server]
	if !ok {
		return ErrNoServer
	}
	if s.state != StateStopped {
		return ErrServerRunning
	}
	m.Server = -1
	s.Map = -1
	r.c.Maps[mID] = m
	r.c.Servers[s.ID] = s
	return r.c.save()
}

type DefaultMap struct {
	Mode               int
	Name               string
	GameMode           int
	Seed               int64
	Structures, Cheats bool
}

func (r RPC) CreateDefaultMap(data DefaultMap, _ *struct{}) error {
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
	r.c.newMap(data.Name, d)
	f, err := os.Create(path.Join(d, "properties.map"))
	if err != nil {
		return err
	}
	defer f.Close()
	m := DefaultMapSettings()
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

// Errors
var (
	ErrNoMap                 = errors.New("no map found")
	ErrMapAlreadyAssigned    = errors.New("map already assigned")
	ErrServerAlreadyAssigned = errors.New("server already assigned")
)
