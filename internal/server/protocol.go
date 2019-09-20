package server

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var errQuit = errors.New("quit")

type protocol struct {
	cache Cache
}

// get <key>
func (p *protocol) get(params []string) (string, error) {
	if len(params) != 2 {
		return "", errors.New("GET insufficient number of params")
	}
	return p.cache.Get(params[1])
}

// set <key> <value> <ttl>
func (p *protocol) set(params []string) (string, error) {
	if len(params) != 4 {
		return "", errors.New("SET insufficient number of params")
	}

	d, err := time.ParseDuration(params[3])
	if err != nil {
		return "", errors.New("SET failed to parse duration")
	}

	p.cache.Set(params[1], params[2], d)

	return "OK", nil
}

// del <key>
func (p *protocol) del(params []string) (string, error) {
	if len(params) != 2 {
		return "", errors.New("DEL insufficient number of params")
	}

	p.cache.Delete(params[1])

	return "OK", nil
}

// quit
func (p *protocol) quit() (string, error) {
	return "Bye", errQuit
}

func (p *protocol) exec(params []string) (string, error) {
	if len(params) == 0 {
		return "", nil
	}

	switch {
	case strings.EqualFold(params[0], "get"):
		return p.get(params)
	case strings.EqualFold(params[0], "set"):
		return p.set(params)
	case strings.EqualFold(params[0], "del"):
		return p.del(params)
	case strings.EqualFold(params[0], "quit"):
		return p.quit()
	}

	return "", fmt.Errorf("invalid command %q", params[0])
}
