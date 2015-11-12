package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"

	"github.com/MJKWoolnough/minewebgen/internal/data"
)

type runner struct {
	s        *data.Server
	shutdown chan struct{}
	io.Writer
}

type Controller struct {
	c *Config

	mu      sync.RWMutex
	running map[int]*runner
}

func (c Controller) Run(id int) error {
	s := c.c.Server(id)
	if s == nil {
		return ErrUnknownServer
	}
	s.Lock()
	defer s.Unlock()
	if s.State != data.StateStopped {
		return ErrServerRunning
	}
	m := c.c.Map(s.Map)
	if m == nil {
		return ErrUnknownServer
	}
	m.RLock()
	defer m.RUnlock()
	mapPath := m.Path
	if !path.IsAbs(mapPath) {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		mapPath = path.Join(pwd, mapPath)
	}
	serverMapPath := path.Join(s.Path, "world")
	if err := os.Remove(serverMapPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Symlink(mapPath, serverMapPath); err != nil {
		return err
	}
	sp := make(ServerProperties)
	f, err := os.Open(path.Join(s.Path, "properties.server"))
	if err != nil {
		return err
	}
	sp.ReadFrom(f)
	f.Close()
	if err != nil {
		return err
	}
	f, err = os.Open(path.Join(m.Path, "properties.map"))
	if err != nil {
		return err
	}
	err = sp.ReadFrom(f)
	f.Close()
	if err != nil {
		return err
	}
	sp["level-name"] = "world"
	f, err = os.Open(path.Join(s.Path, "server.properties"))
	if err != nil {
		return err
	}
	sp.WriteTo(f)
	f.Close()
	s.State = data.StateStarting
	r := &runner{
		s:        s,
		shutdown: make(chan struct{}, 1),
	}
	go c.run(r)
	return nil
}

func (c *Controller) Stop(id int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	r, ok := c.running[id]
	if !ok {
		return errors.New("server not running")
	}
	close(r.shutdown)
	delete(c.running, id)
	return nil
}

var stopCmd = []byte{'\r', '\n', 's', 't', 'o', 'p', '\r', '\n'}

// runs in its own goroutine
func (c *Controller) run(r *runner) {
	cmd := exec.Command("java", append(r.s.Args, "-jar", "server.jar", "nogui")...)
	cmd.Dir = r.s.Path
	r.Writer, _ = cmd.StdinPipe()
	err := cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		c.mu.Lock()
		c.running[r.s.ID] = r
		c.mu.Unlock()
		r.s.Lock()
		r.s.State = data.StateRunning
		r.s.Unlock()
		died := make(chan struct{})
		go func() {
			select {
			case <-r.shutdown:
				r.s.Lock()
				r.s.State = data.StateStopping
				r.s.Unlock()
				t := time.NewTimer(time.Second * 10)
				defer t.Stop()
				for i := 0; i < 6; i++ {
					r.Write(stopCmd)
					select {
					case <-died:
						return
					case <-t.C:
					}
				}
				cmd.Process.Kill()
			case <-died:
				c.mu.Lock()
				delete(c.running, r.s.ID)
				c.mu.Unlock()
			}
		}()
		cmd.Wait()
		r.shutdown = nil
		close(died)
	}
	r.s.Lock()
	r.s.State = data.StateStopped
	r.s.Unlock()
}

func (c *Controller) WriteCmd(id int, cmd string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	r, ok := c.running[id]
	if !ok {
		return ErrUnknownServer
	}
	toWrite := make([]byte, 0, len(cmd)+4)
	toWrite = append(toWrite, '\r', '\n')
	toWrite = append(toWrite, cmd...)
	toWrite = append(toWrite, '\r', '\n')
	_, err := r.Write(toWrite)
	return err
}
