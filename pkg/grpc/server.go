package grpc

import (
	"context"
	"fmt"
	"net"

	proto "github.com/joshturge-io/auth/pkg/grpc/proto"
	"google.golang.org/grpc"
)

// Server is a grpc server
type Server struct {
	gs       *grpc.Server
	listener net.Listener
	serveErr error
}

// NewServer will create a new listener and server with registered services
func NewServer(addr string, services ...proto.Service) (*Server, error) {
	var (
		srv = &Server{}
		err error
	)
	srv.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("unable to listen to address: %s: %w", addr, err)
	}

	srv.gs = grpc.NewServer()

	for _, service := range services {
		service.Register(srv.gs)
	}

	return srv, nil
}

// Serve will start serving the grpc server, errors can be checked through the Err method
func (s *Server) Serve() {
	go func() {
		if err := s.gs.Serve(s.listener); err != nil {
			s.serveErr = err
		}
	}()
}

// Err will return any errors that accured while serving, returns nil when none
func (s Server) Err() error {
	return s.serveErr
}

// Close will close the grpc server, returns an error if context is done
func (s *Server) Close(ctx context.Context) error {
	done := make(chan struct{})
	defer close(done)
	go func() {
		s.gs.GracefulStop()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		s.gs.Stop()
		return ctx.Err()
	case <-done:
	}

	return nil
}
