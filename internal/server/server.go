package server

import (
	"bufio"
	"context"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	listenerDeadline = 5 * time.Second
	connReadDeadline = 10 * time.Second
)

// Cache cache interface that used by the server
type Cache interface {
	Get(key string) (string, error)
	Set(key, value string, ttl time.Duration)
	Delete(key string)
}

// Server TCP server
type Server struct {
	address  string
	logger   *log.Logger
	protocol *protocol
}

// New returns new TCP server
func New(addr string, cache Cache, logger *log.Logger) *Server {
	return &Server{
		address: addr,
		logger:  logger,
		protocol: &protocol{
			cache: cache,
		},
	}
}

// Listen listener for TCP network
func (s *Server) Listen(ctx context.Context) error {
	addr, err := net.ResolveTCPAddr("tcp", s.address)
	if err != nil {
		return err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	wg := new(sync.WaitGroup)

	for {
		select {
		case <-ctx.Done():
			if err := l.Close(); err != nil {
				s.logger.Printf("Failed to close listener: %v", err)
			}
			wg.Wait()
			return nil
		default:
		}

		if err := l.SetDeadline(time.Now().Add(listenerDeadline)); err != nil {
			s.logger.Printf("Failed to set listener deadline: %v", err)
			continue
		}

		conn, err := l.Accept()
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			continue
		}
		if err != nil {
			s.logger.Printf("Failed to accept conn: %v", err)
			continue
		}

		wg.Add(1)
		go func() {
			s.handleConn(ctx, conn)
			wg.Done()
		}()
	}
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	s.logger.Printf("Client %s connected", conn.RemoteAddr().String())

	defer func() {
		if err := conn.Close(); err != nil {
			s.logger.Printf("Failed to close conn: %v", err)
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := conn.SetReadDeadline(time.Now().Add(connReadDeadline)); err != nil {
			s.logger.Printf("Failed to set read deadline: %v", err)
			continue
		}

		message, err := bufio.NewReader(conn).ReadString('\n')
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			continue
		}
		if err != nil {
			s.logger.Printf("Failed to read data: %v", err)
			break
		}

		message = strings.TrimSpace(message)
		params := strings.Fields(message)

		repl, err := s.protocol.exec(params)
		if err == errQuit {
			break
		}
		if err != nil {
			if _, err = conn.Write([]byte(err.Error() + "\n")); err != nil {
				s.logger.Printf("Failed to write error message: %v", err)
			}
			continue
		}

		if repl != "" {
			if _, err := conn.Write([]byte(repl + "\n")); err != nil {
				s.logger.Printf("Failed to write reply: %v", err)
			}
		}
	}
	s.logger.Printf("Client %s disconnected", conn.RemoteAddr().String())
}
