package main

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
	Mode                           int
	Name                           string
	GameMode                       int
	Seed                           int64
	Structures, Cheats, BonusChest bool
}

func (c *Config) CreateDefaultMap(data DefaultMap, _ *struct{}) error {
	return nil
}
