package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"

	"github.com/armon/circbuf"
)

const (
	StateStopped = iota
	StateLoading
	StateRunning
	StateShuttingDown
)

const BufferSize = 1024 * 512

var (
	saveCmd = []byte{'\r', '\n', 's', 'a', 'v', 'e', 'a', 'l', 'l', ' ', '\r', '\n'}
	stopCmd = []byte{'\r', '\n', 's', 't', 'o', 'p', ' ', '\r', '\n'}
)

// .*INFO.* Done ([0-9]+.[0-9]{3}s)! For help, type "help" or "?"
var doneRegex = regexp.MustCompile("Info.* Done \\([0-9]+\\.[0-9]{3}s\\)!")

type Controller struct {
	c       *Config
	running map[int]running
}

func (c *Controller) Start(sID int) error {
	c.c.mu.Lock()
	defer c.c.mu.Unlock()
	s, ok := c.c.Servers[sID]
	if !ok {
		return ErrNoServer
	}
	if s.state != StateStopped {
		return ErrServerRunning
	}
	if s.Map == -1 {
		return ErrNoMap
	}
	m, ok := c.c.Maps[s.Map]
	if !ok {
		return ErrNoMap // Shouldn't happen, different error?
	}
	err := os.Link(path.Join(s.Path, m.Name), m.Path)
	if err != nil {
		return err
	}
	ps, err := os.Open(path.Join(s.Path, "properties.server"))
	if err != nil {
		return err
	}
	defer ps.Close()
	pm, err := os.Open(path.Join(m.Path, "properties.map"))
	if err != nil {
		return err
	}
	defer pm.Close()
	f, err := os.Create(path.Join(s.Path, "server.properties"))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, ps)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, pm)
	if err != nil {
		return err
	}

	s.state = StateLoading
	c.c.Servers[sID] = s
	sc := make(chan struct{})
	c.running[s.ID] = running{shutdown: sc}
	go c.run(s, sc)
	return nil
}

func (c *Controller) Stop(sID int) error {
	c.c.mu.Lock()
	defer c.c.mu.Unlock()
	r, ok := c.running[sID]
	if !ok {
		return ErrServerNotRunning
	}
	close(r.shutdown)
	delete(c.running, sID)
	return nil
}

// runs in its own goroutine
func (c *Controller) run(s Server, shutdown chan struct{}) {
	cmd := exec.Command("java", append(s.Args, "-jar", "server.jar", "nogui")...)
	cmd.Dir = s.Path
	/*r.cb = circbuf.NewBuffer(BufferSize)
	cmd.Stdout = r.cb
	wp, _ := cmd.StdoutPipe()
	r.w.io.MultiWriter(r.cb, wp)
	*/
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
		// Write to Stderr
	} else {

		s.state = StateRunning
		c.c.mu.Lock()
		c.c.Servers[s.ID] = s
		c.c.mu.Unlock()

		died := make(chan struct{})
		go func() {
			select {
			case <-shutdown:
				// write stopCmd to stdin
				s.state = StateShuttingDown
				c.c.mu.Lock()
				c.c.Servers[s.ID] = s
				c.c.mu.Unlock()
			case <-died:
				c.c.mu.Lock()
				delete(c.running, s.ID)
				c.c.mu.Unlock()
			}
		}()

		cmd.Wait()
		shutdown = nil
		close(died)
	}
	s.state = StateStopped
	c.c.mu.Lock()
	c.c.Servers[s.ID] = s
	c.c.mu.Unlock()
}

type running struct {
	shutdown chan struct{}
	cb       *circbuf.Buffer
	stdin    io.Writer
	w        io.Writer
}

func (r *running) Write(p []byte) (int, error) {
	/*if r.id < 0 {
		return 0, ErrServerNotRunning
	}*/
	return r.w.Write(p)
}

// Errors
var (
	ErrServerRunning    = errors.New("server already running")
	ErrServerNotRunning = errors.New("server not running")
)
