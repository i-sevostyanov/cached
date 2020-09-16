package protocol

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrQuit = errors.New("quit")

type Command string

const (
	Get    Command = "get"
	Set    Command = "set"
	Delete Command = "del"
	Stats  Command = "stats"
	Quit   Command = "quit"
)

type Cache interface {
	Get(key string) (string, error)
	Set(key, value string, ttl time.Duration)
	Delete(key string)
	Stats() (hit, miss, size int64)
}

type protocol struct {
	cache Cache
}

func New(cache Cache) *protocol {
	return &protocol{
		cache: cache,
	}
}

func (p *protocol) Exec(command string) (string, error) {
	cmd := strings.TrimSpace(command)
	params := strings.Fields(cmd)

	if len(params) == 0 {
		return "", nil
	}

	switch Command(params[0]) {
	case Get:
		return p.get(params)
	case Set:
		return p.set(params)
	case Delete:
		return p.del(params)
	case Stats:
		return p.stats()
	case Quit:
		return p.quit()
	default:
		return "", fmt.Errorf("invalid command %q", params[0])
	}
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

	ttl, err := time.ParseDuration(params[3])
	if err != nil {
		return "", errors.New("SET failed to parse ttl")
	}

	p.cache.Set(params[1], params[2], ttl)

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

// stats
func (p *protocol) stats() (string, error) {
	hit, miss, size := p.cache.Stats()
	return fmt.Sprintf("Hit: %d, Miss: %d, Size: %d", hit, miss, size), nil
}

// quit
func (p *protocol) quit() (string, error) {
	return "Bye", ErrQuit
}
