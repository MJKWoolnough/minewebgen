package main

import (
	"bufio"
	"io"
	"strings"
)

type ServerProperties map[string]string

var (
	defaultMapSettings = ServerProperties{
		"allow-flight":                 "false",
		"allow-nether":                 "true",
		"announce-player-achievements": "true",
		"difficulty":                   "1",
		"enable-command-block":         "false",
		"force-gamemode":               "false",
		"gamemode":                     "0",
		"generate-structures":          "true",
		"generator-settings":           "",
		"hardcore":                     "false",
		"level-seed":                   "",
		"level-type":                   "DEFAULT",
		"max-build-height":             "256",
		"max-world-size":               "29999984",
		"motd":                         "A MineWebGen Server",
		"player-idle-timeout":          "0",
		"pvp":                  "true",
		"resource-pack":        "",
		"resource-pack-hash":   "",
		"spawn-animals":        "true",
		"spawn-monsters":       "true",
		"spawn-npcs":           "true",
		"spawn-protection":     "16",
		"use-native-transport": "true",
		"view-distance":        "10",
		"grow-trees":           "true",
	}
	defaultServerSettings = ServerProperties{
		"enable-query":                  "false",
		"enable-rcon":                   "false",
		"max-players":                   "20",
		"max-tick-time":                 "60000",
		"network-compression-threshold": "256",
		"online-mode":                   "true",
		"op-permission-level":           "4",
		"query.port":                    "25565",
		"rcon.password":                 "",
		"rcon.port":                     "25575",
		"server-ip":                     "",
		"server-port":                   "25565",
		"snooper-enabled":               "false",
		"white-list":                    "false",
		"verify-names":                  "true",
		"admin-slot":                    "false",
		"public":                        "true",
		"server-name":                   "",
		"max-connections":               "3",
	}
)

func DefaultMapSettings() ServerProperties {
	return defaultMapSettings.Clone()
}

func DefaultServerSettings() ServerProperties {
	return defaultServerSettings.Clone()
}

func (s ServerProperties) Clone() ServerProperties {
	m := make(ServerProperties)
	for k, v := range s {
		m[k] = v
	}
	return m
}

func (s ServerProperties) ReadFrom(r io.Reader) error {
	br := bufio.NewReader(r)
	data := make(map[string]string)
	for {
		l, err := br.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if l[0] == '#' {
			continue
		}
		if l[len(l)-1] == '\r' {
			l = l[:len(l)-1]
		}
		parts := strings.SplitN(l, "=", 2)
		if len(parts) != 2 {
			continue
		}
		data[parts[0]] = parts[1]
	}
}

func (s ServerProperties) WriteTo(w io.Writer) error {
	toWrite := make([]byte, 0, 1024)
	for k, v := range s {
		toWrite = toWrite[:0]
		toWrite = append(toWrite, k...)
		toWrite = append(toWrite, '=')
		toWrite = append(toWrite, v...)
		toWrite = append(toWrite, '\n')
		_, err := w.Write(toWrite)
		if err != nil {
			return err
		}
	}
	return nil
}
