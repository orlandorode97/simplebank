package grpc

import (
	"strings"

	simplebankpb "github.com/orlandorode97/simple-bank/generated/simplebank"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	metadataAuthorizationHeader = "authorization"

	updateUserRPC = "/simplebank.SimplebankService/UpdateUser"
)

var protectedRPCs = map[string]bool{
	updateUserRPC: true,
}

func (s *GRPCServer) AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		_, ok := protectedRPCs[info.FullMethod]
		if ok {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Errorf(codes.Unauthenticated, "metadata for %v is not provided", info.FullMethod)
			}

			authValues := md.Get(metadataAuthorizationHeader)
			if len(authValues) == 0 {
				return nil, status.Errorf(codes.Unauthenticated, "metadata authorization not provided")
			}
			authHeader := strings.Fields(authValues[0])
			if len(authHeader) < 2 {
				return nil, status.Error(codes.Unauthenticated, "metadata authorization header format is invalid")
			}

			accessToken := authHeader[1]
			payload, err := s.tokenMaker.VerfifyToken(accessToken)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "%v", err)
			}

			switch info.FullMethod {
			case updateUserRPC:
				updateUserRequest, _ := req.(*simplebankpb.UpdateUserRequest)
				if updateUserRequest.GetUsername() != payload.Username {
					return nil, status.Errorf(codes.PermissionDenied, "unable to update user information")
				}
			}
		}

		return handler(ctx, req)
	}
}
