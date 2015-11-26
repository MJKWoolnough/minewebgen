package data

import "sync"

type State int

const (
	StateStopped State = iota
	StateStarting
	StateRunning
	StateStopping
	StateBusy
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
	case StateBusy:
		return "Busy"
	}
	return ""
}

type ServerSettings struct {
	ServerName                         string
	ListenAddr                         string
	DirServers, DirMaps, DirGenerators string
	GeneratorExecutable                string
	GeneratorMaxMem                    uint64
}

type Server struct {
	ID   int
	Path string

	sync.RWMutex `json:"-"`
	Name         string
	Args         []string
	Map          int
	State        State
}

type Map struct {
	ID   int
	Path string

	sync.RWMutex `json:"-"`
	Name         string
	Server       int
}

type Generator struct {
	ID   int
	Path string
	Name string
}
