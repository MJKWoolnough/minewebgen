package main

import (
	"archive/zip"
	"encoding/json"
	"image/color"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"sync"

	"github.com/MJKWoolnough/minecraft"
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
			l := len(s.List)
			if l != n {
				s.List[n], s.List[l-1] = s.List[l-1], s.List[n]
			}
			s.List = s.List[:l-1]
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
			l := len(m.List)
			if l != n {
				m.List[n], m.List[l-1] = m.List[l-1], m.List[n]
			}
			m.List = m.List[:l-1]
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
	mu    sync.RWMutex
	list  map[string]*generator
	names []string
}

func (gs *Generators) Get(name string) *generator {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.list[name]
}

func (gs *Generators) Names() []string {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	n := make([]string, len(gs.names))
	copy(n, gs.names)
	return n
}

var empty = struct {
	Palette color.Palette
	Biomes  []minecraft.Biome
	Blocks  []blocks
}{
	color.Palette{color.RGBA{}},
	[]minecraft.Biome{minecraft.Plains},
	[]blocks{blocks{}},
}

func (gs *Generators) Load(gPath string) error {
	d, err := os.Open(gPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	fs, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	gs.list = make(map[string]*generator)
	gs.names = make([]string, 0, 32)
	for _, name := range fs {
		if len(name) < 5 {
			continue
		}
		if name[len(name)-4:] != ".gen" {
			continue
		}
		g := new(generator)
		f, err := os.Open(path.Join(gPath, name))
		if err != nil {
			continue
		}
		err = json.NewDecoder(f).Decode(g)
		if err != nil {
			continue
		}
		gName := name[:len(name)-4]

		// Sanity Checks
		if len(g.Terrain.Blocks) < len(g.Terrain.Colours) {
			g.Terrain.Colours = g.Terrain.Colours[:len(g.Terrain.Blocks)]
		}
		if len(g.Terrain.Colours) == 0 {
			g.Terrain.Colours = empty.Palette
			g.Terrain.Blocks = empty.Blocks
		}
		if len(g.Terrain.Colours) == g.Terrain.Blocks {
			g.Terrain.Blocks = append(g.Terrain.Blocks, minecraft.Block{ID: 9}) // water block
		}
		if len(g.Biomes.Values) < len(g.Biomes.Colours) {
			g.Biomes.Colours = g.Biomes.Colours[:len(g.Biomes.Values)]
		}
		if len(g.Biomes.Colours) == 0 {
			g.Biomes.Colours = empty.Palette
			g.Biomes.Values = empty.Biomes
		}
		if len(g.Plants.Blocks) < len(g.Plants.Colours) {
			g.Plants.Colours = g.Plants.Colours[:len(g.Plants.Blocks)]
		}
		if len(g.Plants.Colours) == 0 {
			g.Plants.Colours = empty.Palette
			g.Plants.Blocks = empty.Blocks
		}

		gs.list[gName] = g
		gs.names = append(gs.names, gName)

	}
	sort.Strings(gs.names)
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
	c.filename = filename
	f, err := os.Open(filename)
	if err == nil {
		defer f.Close()
		err = json.NewDecoder(f).Decode(c)
		if err == nil {
			err = c.Generators.Load(c.ServerSettings.DirGenerators)
		}
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
