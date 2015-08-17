package main

import (
	"errors"
	"io"
	"os/exec"
	"path"

	"github.com/armon/circbuf"
)

type running struct {
	id    int
	cmd   *exec.Cmd
	cb    *circbuf.Buffer
	stdin io.Writer
	w     io.Writer
}

func (r *running) Write(p []byte) (int, error) {
	if r.id < 0 {
		return 0, ErrServerNotRunning
	}
	return r.w.Write(p)
}

var serverRunning = &running{id: -1}

func stopServer() error {
	if serverRunning.id < 0 {
		return ErrServerNotRunning
	}
	serverRunning.stdin.Write(stopCmd)
	err := serverRunning.cmd.Wait()
	serverRunning = &running{id: -1}
	return err
}

func saveServer() error {
	if serverRunning.id < 0 {
		return ErrServerNotRunning
	}
	_, err := serverRunning.stdin.Write(saveCmd)
	return err
}

var (
	saveCmd = []byte{'\r', '\n', 's', 'a', 'v', 'e', 'a', 'l', 'l', ' ', '\r', '\n'}
	stopCmd = []byte{'\r', '\n', 's', 't', 'o', 'p', ' ', '\r', '\n'}
)

func (c *Config) startServer(id int) error {
	if serverRunning.id < 0 {
		return ErrServerRunning
	}
	r := &running{id: id}
	s := c.Servers[id]
	d, e := path.Split(s.Path)
	r.cmd = exec.Command(e, s.Args...)
	r.cmd.Dir = d
	stdin, err := r.cmd.StdinPipe()
	if err != nil {
		return nil
	}
	r.cb, err = circbuf.NewBuffer(1024 * 1024)
	if err != nil {
		return nil
	}
	r.stdin = stdin
	r.w = io.MultiWriter(stdin, r.cb)
	r.cmd.Stdout = r.cb
	r.cmd.Stderr = r.cb
	serverRunning = r
	return serverRunning.cmd.Start()
}

// Errors
var (
	ErrServerRunning    = errors.New("server already running")
	ErrServerNotRunning = errors.New("server not running")
)
