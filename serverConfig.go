package main

import (
	"bufio"
	"io"
	"strings"
)

func ReadConfig(r io.Reader) (map[string]string, error) {
	br := bufio.NewReader(r)
	data := make(map[string]string)
	for {
		l, err := br.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return data, nil
			}
			return nil, err
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

func SaveConfig(w io.Writer, c map[string]string) error {
	toWrite := make([]byte, 0, 1024)
	for k, v := range c {
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
