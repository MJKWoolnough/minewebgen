package main

import (
	"encoding/json"
	"os"
)

type Server struct {
	Name, Path string
}

type Config struct {
	filename   string
	ServerName string
	Port       uint16
	Servers    []Server
}

func loadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	c := &Config{
		Port: 8080,
	}
	err = json.NewDecoder(f).Decode(c)
	if err != nil {
		return nil, err
	}
	c.filename = filename
	return c, nil
}

func (c *Config) save() error {
	f, err := os.Create(c.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(c)
}

func (c *Config) Name(_ struct{}, serverName *string) error {
	*serverName = c.ServerName
	return nil
}
