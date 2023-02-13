package grpc

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	simplebankpb "github.com/orlandorode97/simple-bank/generated/simplebank"
	"github.com/orlandorode97/simple-bank/generated/sql/simplebanksql"
	"github.com/orlandorode97/simple-bank/pkg"
	"github.com/orlandorode97/simple-bank/pkg/validations"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const metadataUsergAgentKey = "user-agent"

func (s *GRPCServer) CreateUser(ctx context.Context, req *simplebankpb.CreateUserRequest) (*simplebankpb.CreateUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "CreateUserRequest is empty")
	}

	if err := isCreateUserReqValid(req); err != nil {
		return nil, err
	}

	hashed, err := pkg.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to hash password: %v", err)
	}

	args := simplebanksql.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashed,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	user, err := s.store.CreateUser(ctx, args)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "user already created: %v", err)
			}
		}
	}

	return &simplebankpb.CreateUserResponse{
		User: &simplebankpb.User{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
			CreatedAt:         timestamppb.New(user.CreateadAt),
		},
	}, nil
}

func (s *GRPCServer) Login(ctx context.Context, req *simplebankpb.LoginRequest) (*simplebankpb.LoginResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "LoginRequest is empty")
	}

	if err := isLoginRequestValid(req); err != nil {
		return nil, err
	}
	user, err := s.store.GetUser(ctx, req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user does not exist: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	if ok := pkg.ComparedPassoword(req.Password, user.HashedPassword); !ok {
		return nil, status.Error(codes.InvalidArgument, "incorrect username/password")
	}

	token, accessPayload, err := s.tokenMaker.CreateToken(user.Username, s.config.TokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to generate token: %v", err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.Username, s.config.TokenRefreshDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to generate refresh token: %v", err)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "unable to get incoming request metadata")
	}

	userAgentValues := md.Get(metadataUsergAgentKey)
	if len(userAgentValues) == 0 {
		return nil, status.Errorf(codes.Internal, "metadata user agent not provided")
	}

	peer, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "peer information about RPCs is not provided")
	}

	session, err := s.store.CreateSession(ctx, simplebanksql.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    userAgentValues[0],
		ClientIp:     peer.Addr.String(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create session: %v", err)
	}

	return &simplebankpb.LoginResponse{
		SessionId:             session.ID.String(),
		AccessToken:           token,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
		User: &simplebankpb.User{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
			CreatedAt:         timestamppb.New(user.CreateadAt),
		},
	}, nil
}

func (s *GRPCServer) UpdateUser(ctx context.Context, req *simplebankpb.UpdateUserRequest) (*simplebankpb.UpdateUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "UpdateUserRequest is empty")
	}

	if err := isUpdateUserReqValid(req); err != nil {
		return nil, err
	}

	args := simplebanksql.UpdateUserParams{
		Username: req.GetUsername(),
		FullName: sql.NullString{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: sql.NullString{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
	}

	if req.Password != nil {
		hashed, err := pkg.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "unable to hash password: %v", err)
		}
		args.HashedPassword = sql.NullString{
			String: hashed,
			Valid:  true,
		}
	}

	if args.HashedPassword.Valid {
		args.PasswordChangedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	updateUser, err := s.store.UpdateUser(ctx, args)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "username not found: %v", err)
		}

		return nil, status.Errorf(codes.NotFound, "unable to update user: %v", err)
	}

	return &simplebankpb.UpdateUserResponse{
		User: &simplebankpb.User{
			Username:          updateUser.Username,
			FullName:          updateUser.FullName,
			Email:             updateUser.Email,
			PasswordChangedAt: timestamppb.New(updateUser.PasswordChangedAt),
			CreatedAt:         timestamppb.New(updateUser.CreateadAt),
		},
	}, nil
}

func isCreateUserReqValid(req *simplebankpb.CreateUserRequest) error {
	createUserValidator := validations.NewCreateUserValidator(req)
	return validations.BuildErrDetails(createUserValidator, "CreateUserRequest error")
}

func isLoginRequestValid(req *simplebankpb.LoginRequest) error {
	loginValidator := validations.NewLoginValidator(req)
	return validations.BuildErrDetails(loginValidator, "LoginRequest error")
}

func isUpdateUserReqValid(req *simplebankpb.UpdateUserRequest) error {
	updateUserValidator := validations.NewUpdateUserValidator(req)
	return validations.BuildErrDetails(updateUserValidator, "UpdateUserRequest error")
}
