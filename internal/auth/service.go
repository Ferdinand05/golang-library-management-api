package auth

import (
	"context"
	"ferdinand/library-management-system-api/internal/crypto"
	"ferdinand/library-management-system-api/internal/user"
)

type Service interface {
	Login(ctx context.Context, request LoginRequest) (string, error)
}

type service struct {
	userRepo user.Repository
	jwtService JWTGenerator
}

func NewService(userRepo user.Repository,jwtService JWTGenerator ) *service {
	return &service{
		userRepo: userRepo,
		jwtService: jwtService,
	}
}

func (s *service) Login(ctx context.Context,request LoginRequest) (string,error) {

	user,err := s.userRepo.FindByEmail(ctx,request.Email)
	if err != nil {
		return "",ErrInvalidCredentials
	}

	checked := crypto.CheckPassword(request.Password,user.Password)

	if !checked {
		return "",ErrInvalidCredentials
	}

	token,err := s.jwtService.GenerateToken(int64(user.ID),user.Email,user.Role)
	if err != nil {
		return "",err
	}

	return token,nil
}
