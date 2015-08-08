package main

import (
	"encoding/json"
	"os"
)

type Server struct {
	Name, Path string
}

type Config struct {
	Port    uint16
	Servers []Server
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
	return c, nil
}
