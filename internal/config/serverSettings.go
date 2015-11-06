package config

import "sync"

type State int

const (
	StateStopped State = iota
	StateStarting
	StateRunning
	StateStopping
)

func (s State) String() string {
	switch s {
	case StateStopped:
		return "Stopped"
	case StateStarting:
		return "Starting"
	case StateRunning:
		return "Running"
	case StateStopping:
		return "Stopping"
	}
	return ""
}

type ServerSettings struct {
	ServerName          string
	ListenAddr          string
	DirServers, DirMaps string
}

type Server struct {
	ID   int
	Path string

	mu    sync.RWMutex
	Name  string
	Args  []string
	Map   int
	State State `json:",omitempty"`
}

type Map struct {
	ID   int
	Path string

	mu     sync.RWMutex
	Name   string
	Server int
}
