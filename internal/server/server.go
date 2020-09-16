package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/i-sevostyanov/chached/internal/protocol"
)

const (
	listenerDeadline = 5 * time.Second
	connReadDeadline = 10 * time.Second
)

type Protocol interface {
	Exec(command string) (string, error)
}

// Server TCP server
type Server struct {
	address  string
	logger   *log.Logger
	protocol Protocol
}

// New returns new TCP server
func New(addr string, protocol Protocol, logger *log.Logger) *Server {
	return &Server{
		address:  addr,
		logger:   logger,
		protocol: protocol,
	}
}

// Listen listener for TCP network
func (s *Server) Listen(ctx context.Context) error {
	addr, err := net.ResolveTCPAddr("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to resolve address: %w", err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start tcp server: %w", err)
	}

	wg := new(sync.WaitGroup)

	for {
		select {
		case <-ctx.Done():
			if err = listener.Close(); err != nil {
				s.logger.Printf("Failed to close listener: %v", err)
			}
			wg.Wait()
			return nil
		default:
		}

		if err = listener.SetDeadline(time.Now().Add(listenerDeadline)); err != nil {
			s.logger.Printf("Failed to set listener deadline: %v", err)
			continue
		}

		conn, err := listener.Accept()
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

	reader := bufio.NewReader(conn)

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
		}

		if err := conn.SetReadDeadline(time.Now().Add(connReadDeadline)); err != nil {
			s.logger.Printf("Failed to set read deadline: %v", err)
			continue
		}

		command, err := reader.ReadString('\n')
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			continue
		}
		if err != nil {
			s.logger.Printf("Failed to read data: %v", err)
			break
		}

		repl, err := s.protocol.Exec(command)
		if errors.Is(err, protocol.ErrQuit) {
			break
		}
		if err != nil {
			if _, err = conn.Write([]byte(err.Error() + "\n")); err != nil {
				s.logger.Printf("Failed to write error message: %v", err)
			}
			continue
		}

		if repl != "" {
			if _, err = conn.Write([]byte(repl + "\n")); err != nil {
				s.logger.Printf("Failed to write reply: %v", err)
			}
		}
	}

	s.logger.Printf("Client %s disconnected", conn.RemoteAddr().String())
}
