package service

import (
	"context"
	"log"

	"github.com/joshturge-io/auth/pkg/auth"
	proto "github.com/joshturge-io/auth/pkg/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type GRPCAuthService struct {
	srv *auth.Service
	lg  *log.Logger
}

func NewGRPCAuthService(as *auth.Service, lg *log.Logger) *GRPCAuthService {
	return &GRPCAuthService{as, lg}
}

func (ga *GRPCAuthService) Login(ctx context.Context, cred *proto.Credentials) (*proto.Session,
	error) {
	session, err := ga.srv.SessionWithChallenge(ctx, cred.GetUsername(), cred.GetPassword())
	if err != nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "failed to create session from challenge: %w",
			err)
	}

	return &proto.Session{UserId: session.UserId, Jwt: session.JWT, RefreshToken: session.Refresh}, nil
}

func (ga *GRPCAuthService) Refresh(ctx context.Context, sess *proto.Session) (*proto.Session, error) {
	session, err := ga.srv.Renew(ctx, &auth.Session{UserId: sess.GetUserId(),
		Refresh: sess.GetRefreshToken(), JWT: sess.GetJwt()})
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to renew session: %w", err)
	}

	return &proto.Session{UserId: session.UserId, Jwt: session.JWT, RefreshToken: session.Refresh}, nil
}

func (ga *GRPCAuthService) ValidateJWT(ctx context.Context, jw *proto.JWT) (*proto.ValidityStatus,
	error) {
	isValid, err := ga.srv.IsValidJWT(jw.GetToken())
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to validate token: %w", err)
	}

	return &proto.ValidityStatus{Valid: isValid}, nil
}

func (ga *GRPCAuthService) Logout(ctx context.Context, sess *proto.Session) (*proto.LogoutStatus,
	error) {
	if err := ga.srv.DestroySession(ctx, &auth.Session{UserId: sess.GetUserId(),
		Refresh: sess.GetRefreshToken(), JWT: sess.GetJwt()}); err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to destroy session: %w", err)
	}

	return &proto.LogoutStatus{UserId: sess.GetUserId(), Success: true, Msg: "user has been logged out"},
		nil
}

func (ga *GRPCAuthService) Register(s *grpc.Server) {
	proto.RegisterAuthenticationServer(s, ga)
}

type AuthorisationService struct{}

func NewAuthorisationService() *AuthorisationService {
	return &AuthorisationService{}
}

func (as *AuthorisationService) Check(ctx context.Context, perm *proto.Permission) (*proto.CheckStatus,
	error) {
	return nil, grpc.Errorf(codes.Unimplemented, "not implemented")
}

func (as *AuthorisationService) Register(s *grpc.Server) {
	proto.RegisterAuthorisationServer(s, as)
}
