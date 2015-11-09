package main

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"path"
	"strconv"

	"github.com/MJKWoolnough/memio"
	"github.com/MJKWoolnough/minecraft"
	"github.com/MJKWoolnough/minewebgen/internal/data"
)

type RPC struct {
	c *Config
}

func (r RPC) Settings(_ struct{}, settings *data.ServerSettings) error {
	*settings = r.c.Settings()
	return nil
}

func (r RPC) SetSettings(settings data.ServerSettings, _ *struct{}) error {
	settings.DirMaps = path.Clean(settings.DirMaps)
	settings.DirServers = path.Clean(settings.DirServers)
	if settings.DirMaps == settings.DirServers {
		return errors.New("map and server paths cannot be the same")
	}
	r.c.SetSettings(settings)
	go r.c.Save()
	return nil
}

func (r RPC) ServerName(_ struct{}, serverName *string) error {
	*serverName = r.c.Settings().ServerName
	return nil
}

func (r RPC) ServerList(_ struct{}, list *[]data.Server) error {
	r.c.Servers.mu.RLock()
	defer r.c.Servers.mu.RUnlock()
	*list = make([]data.Server, len(r.c.Servers.List))
	for n, s := range r.c.Servers.List {
		(*list)[n] = *s
	}
	return nil
}

func (r RPC) MapList(_ struct{}, list *[]data.Map) error {
	r.c.Maps.mu.RLock()
	defer r.c.Maps.mu.RUnlock()
	*list = make([]data.Map, len(r.c.Maps.List))
	for n, m := range r.c.Maps.List {
		(*list)[n] = *m
	}
	return nil
}

func (r RPC) Server(id int, s *data.Server) error {
	ser := r.c.Server(id)
	ser.RLock()
	defer ser.RUnlock()
	*s = *ser
	return nil
}

func (r RPC) Map(id int, m *data.Map) error {
	mp := r.c.Map(id)
	mp.RLock()
	defer mp.RUnlock()
	*m = *mp
	return nil
}

func (r RPC) SetServer(s data.Server, _ *struct{}) error {
	ser := r.c.Server(s.ID)
	if ser == nil {
		return ErrUnknownServer
	}
	if ser.State != data.StateStopped {
		return ErrServerRunning
	}
	ser.Lock()
	defer ser.Unlock()
	ser.Name = s.Name
	ser.Args = s.Args
	go r.c.Save()
	return nil
}

func (r RPC) SetMap(m data.Map, _ *struct{}) error {
	mp := r.c.Map(m.ID)
	if mp == nil {
		return ErrUnknownMap
	}
	mp.RLock()
	sID := mp.Server
	mp.RUnlock()
	if sID != -1 {
		ser := r.c.Server(sID)
		if ser != nil {
			ser.RLock()
			s := ser.State
			ser.RUnlock()
			if s != data.StateStopped {
				return ErrServerRunning
			}
		}
	}
	mp.Lock()
	defer mp.Unlock()
	mp.Name = m.Name
	go r.c.Save()
	return nil
}

func (r RPC) SetServerMap(ids [2]int, _ *struct{}) error {
	if ids[0] != -1 {
		serv := r.c.Server(ids[0])
		if serv == nil {
			return ErrUnknownServer
		}
		serv.RLock()
		mID := serv.Map
		s := serv.State
		serv.RUnlock()
		if s != data.StateStopped {
			return ErrServerRunning
		}
		if mID != -1 {
			if mID == ids[1] {
				return nil
			}
			mp := r.c.Map(mID)
			if mp != nil {
				mp.Lock()
				mp.Server = -1
				mp.Unlock()
			}
		}
		serv.Lock()
		serv.Map = ids[1]
		serv.Unlock()
	}
	if ids[1] != -1 {
		mp := r.c.Map(ids[1])
		if mp == nil {
			return ErrUnknownMap
		}
		mp.RLock()
		sID := mp.Server
		mp.RUnlock()
		if sID != -1 {
			serv := r.c.Server(sID)
			if serv != nil {
				serv.RLock()
				s := serv.State
				serv.RUnlock()
				if s != data.StateStopped {
					return ErrServerRunning
				}
				serv.Lock()
				serv.Map = -1
				serv.Unlock()
			}
		}
		mp.Lock()
		mp.Server = ids[0]
		mp.Unlock()
	}
	go r.c.Save()
	return nil
}

func (r RPC) ServerProperties(id int, sp *ServerProperties) error {
	s := r.c.Server(id)
	if s == nil {
		return ErrUnknownServer
	}
	s.RLock()
	p := s.Path
	s.RUnlock()
	f, err := os.Open(path.Join(p, "properties.server"))
	if err != nil {
		return err
	}
	defer f.Close()
	*sp = make(ServerProperties)
	return sp.ReadFrom(f)
}

func (r RPC) SetServerProperties(sp data.ServerProperties, _ *struct{}) error {
	s := r.c.Server(sp.ID)
	if s == nil {
		return ErrUnknownServer
	}
	s.RLock()
	p := s.Path
	s.RUnlock()
	f, err := os.Create(path.Join(p, "properties.server"))
	if err != nil {
		return err
	}
	defer f.Close()
	return ServerProperties(sp.Properties).WriteTo(f)
}

func (r RPC) MapProperties(id int, mp *ServerProperties) error {
	m := r.c.Map(id)
	if m == nil {
		return ErrUnknownMap
	}
	m.RLock()
	p := m.Path
	m.RUnlock()
	f, err := os.Open(path.Join(p, "properties.map"))
	if err != nil {
		return err
	}
	defer f.Close()
	*mp = make(ServerProperties)
	return mp.ReadFrom(f)
}

func (r RPC) SetMapProperties(sp data.ServerProperties, _ *struct{}) error {
	m := r.c.Map(sp.ID)
	if m == nil {
		return ErrUnknownMap
	}
	m.RLock()
	p := m.Path
	m.RUnlock()
	f, err := os.Create(path.Join(p, "properties.map"))
	if err != nil {
		return err
	}
	defer f.Close()
	return ServerProperties(sp.Properties).WriteTo(f)
}

func (r RPC) RemoveServer(id int, _ *struct{}) error {
	s := r.c.Server(id)
	if s == nil {
		return ErrUnknownServer
	}
	s.Lock()
	defer s.Unlock()
	if s.State != data.StateStopped {
		return ErrServerRunning
	}
	if s.Map >= 0 {
		m := r.c.Map(s.Map)
		m.Lock()
		m.Server = -1
		m.Unlock()
	}
	s.ID = -1
	r.c.RemoveServer(id)
	go r.c.Save()
	return nil
}

func (r RPC) RemoveMap(id int, _ *struct{}) error {
	m := r.c.Map(id)
	if m == nil {
		return ErrUnknownMap
	}
	m.Lock()
	defer m.Unlock()
	if m.Server >= 0 {
		s := r.c.Server(m.Server)
		m.Lock()
		defer m.Unlock()
		if s.State != data.StateStopped {
			return ErrServerRunning
		}
		m.Server = -1
	}
	m.ID = -1
	r.c.RemoveMap(id)
	go r.c.Save()
	return nil
}

func (r RPC) CreateDefaultMap(data data.DefaultMap, _ *struct{}) error {
	return r.createMap(data, "")
}

func (r RPC) createMap(data data.DefaultMap, generatorSettings string) error {
	if data.Seed == 0 {
		data.Seed = rand.Int63()
	}
	m := r.c.NewMap()
	if m == nil {
		return errors.New("failed to create map")
	}
	m.Lock()
	defer m.Unlock()
	p, err := minecraft.NewFilePath(m.Path)
	if err != nil {
		r.c.RemoveMap(m.ID)
		return err
	}
	l, err := minecraft.NewLevel(p)
	if err != nil {
		r.c.RemoveMap(m.ID)
		return err
	}
	l.GameMode(data.GameMode)
	l.LevelName(data.Name)
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
	if generatorSettings != "" {
		l.GeneratorOptions(generatorSettings)
	}
	l.Save()
	f, err := os.Create(path.Join(m.Path))
	if err != nil {
		r.c.RemoveMap(m.ID)
		return err
	}
	defer f.Close()
	ms := DefaultMapSettings()
	ms["gamemode"] = strconv.Itoa(int(data.GameMode))
	if !data.Structures {
		ms["generate-structures"] = "false"
	}
	if data.GameMode == 3 {
		ms["hardcore"] = "true"
	}
	if generatorSettings != "" {
		ms["generator-settings"] = generatorSettings
	}
	ms["level-seed"] = strconv.FormatInt(data.Seed, 10)
	ms["motd"] = data.Name
	switch data.Mode {
	case 0:
		ms["level-type"] = minecraft.DefaultGenerator
	case 1:
		ms["level-type"] = minecraft.FlatGenerator
	case 2:
		ms["level-type"] = minecraft.LargeBiomeGenerator
	case 3:
		ms["level-type"] = minecraft.AmplifiedGenerator
	case 4:
		ms["level-type"] = minecraft.CustomGenerator
	case 5:
		ms["level-type"] = minecraft.DebugGenerator
	}
	if err := ms.WriteTo(f); err != nil {
		return err
	}
	go r.c.Save()
	return nil
}

func (r RPC) CreateSuperflatMap(data data.SuperFlatMap, _ *struct{}) error {
	return r.createMap(data.DefaultMap, data.GeneratorSettings)
}

func (r RPC) CreateCustomMap(data data.CustomMap, _ *struct{}) error {
	// check settings for validity
	var buf []byte
	err := json.NewEncoder(memio.Create(&buf)).Encode(data.GeneratorSettings)
	if err != nil {
		return err
	}
	return r.createMap(data.DefaultMap, string(buf))
}

// Errors

var (
	ErrUnknownServer = errors.New("unknown server")
	ErrUnknownMap    = errors.New("unknown map")
	ErrServerRunning = errors.New("server running")
)
