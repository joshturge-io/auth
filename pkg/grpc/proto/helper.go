package proto_auth

import grpc "google.golang.org/grpc"

type Service interface {
	Register(*grpc.Server)
}
