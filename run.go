package main

import (
	"errors"
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

// 2015-09-27 15:33:41 [INFO] [Minecraft-Server] Done (3.959s)! For help, type "help" or "?"
var doneRegex = regexp.MustCompile("^[0-9]{4} [0-9]{2}:[0-9]{2}:[0-9}{2} \\[Info\\] \\[Minecraft-Server\\] Done ")

type controller struct {
	c       *Config
	running map[int]running
}

func (c *controller) Start(sID int) error {
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
	s.state = StateLoading
	c.c.Servers[sID] = s
	go c.run(s)
	return nil
}

func (c *controller) Stop(sID int) {
	c.c.mu.RLock()
	defer c.c.mu.RUnlock()
	r, ok := c.running[sID]
	if !ok {
		return ErrServerNotRunning
	}
	close(r.shutdown)
	delete(c.running, sID)
	return nil
}

// runs in its own goroutine
func (c *controller) run(s Server) {
	r := running{
		shutdown: make(chan struct{}),
	}
	c.c.mu.Lock()
	c.running[s.ID] = r
	c.c.mu.Unlock()
	cmd := exec.Command(path.Join(s.Path, "server.jar"), s.Args...)
	cmd.Dir = s.Path
	/*r.cb = circbuf.NewBuffer(BufferSize)
	cmd.Stdout = r.cb
	wp, _ := cmd.StdoutPipe()
	r.w.io.MultiWriter(r.cb, wp)
	*/
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
}

type running struct {
	shutdown chan struct{}
	cb       *circbuf.Buffer
	stdin    io.Writer
	w        io.Writer
}

func (r *running) Write(p []byte) (int, error) {
	if r.id < 0 {
		return 0, ErrServerNotRunning
	}
	return r.w.Write(p)
}

// Errors
var (
	ErrServerRunning    = errors.New("server already running")
	ErrServerNotRunning = errors.New("server not running")
)
