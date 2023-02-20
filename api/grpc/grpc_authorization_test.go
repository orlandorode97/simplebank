package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/orlandorode97/simple-bank/config"
	"github.com/orlandorode97/simple-bank/generated/simplebank"
	"github.com/orlandorode97/simple-bank/pkg/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var fakeToken = "Bearer v2.local.sIgVm0es9uswZliPdyXOOi99czPbpl41KOUu45e62BvCaL5H3kHNibrbRZkM1-wW091ARzNexLY8g0GZA0-WCNsgs8GZLClEk5TJbgQjf__yExZRh2qMnqxfVr_KS9WoqKVlU-WrAG6TRUXZo43OSJQkeNBnB8Gq4rN2A8HYeA3ms20up80dgz2rpY79F9ILvPrAIzxNkDSE51vAxv50BTShuel3F3hXgReHsDv2PJCnMBnMyE_AfePxJ6WJ1obXSIUpSsOQX6wjwdQdOIcXZ853c-NPYMVU-abXJhhLVvvHyNZPi1wcEvjt.eyJraWQiOiAiMTIzNDUifQ"

func TestUnaryInterceptor(t *testing.T) {
	conf := &config.Config{
		SymmetricKey: "iQ9m6CjMXwEFEdTDYLrLw3krZq6ewKep",
	}

	tokenMaker, err := token.NewPasetoMaker(conf.SymmetricKey)
	if err != nil {
		t.Fatal(err)
	}

	server := &GRPCServer{
		tokenMaker: tokenMaker,
	}

	accessToken, _, err := server.tokenMaker.CreateToken("orlandorode97", 1*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		desc     string
		req      interface{}
		info     *grpc.UnaryServerInfo
		handler  grpc.UnaryHandler
		metadata metadata.MD

		wantGRPCCode codes.Code
	}{
		{
			desc: "success - interceptor without any issue on protected RPCs",
			req: &simplebank.UpdateUserRequest{
				Username: "orlandorode97",
			},
			info: &grpc.UnaryServerInfo{
				FullMethod: updateUserRPC,
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
			metadata: metadata.MD{
				metadataAuthorizationHeader: []string{
					"Bearer " + accessToken,
				},
			},

			wantGRPCCode: codes.OK,
		},
		{
			desc: "success - interceptor without any issue non-protected RPCs",
			req: &simplebank.UpdateUserRequest{
				Username: "juanito97",
			},
			info: &grpc.UnaryServerInfo{
				FullMethod: "Login",
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},

			wantGRPCCode: codes.OK,
		},
		{
			desc: "failure - metadata not provided",
			req: &simplebank.UpdateUserRequest{
				Username: "orlandorode97",
			},
			info: &grpc.UnaryServerInfo{
				FullMethod: updateUserRPC,
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},

			wantGRPCCode: codes.Unauthenticated,
		},
		{
			desc: "failure - authorization header empty",
			req: &simplebank.UpdateUserRequest{
				Username: "orlandorode97",
			},
			info: &grpc.UnaryServerInfo{
				FullMethod: updateUserRPC,
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},

			metadata: metadata.MD{
				metadataAuthorizationHeader: make([]string, 0),
			},

			wantGRPCCode: codes.Unauthenticated,
		},
		{
			desc: "failure - authorization header different format",
			req: &simplebank.UpdateUserRequest{
				Username: "orlandorode97",
			},
			info: &grpc.UnaryServerInfo{
				FullMethod: updateUserRPC,
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},

			metadata: metadata.MD{
				metadataAuthorizationHeader: []string{
					"Bearer",
				},
			},

			wantGRPCCode: codes.Unauthenticated,
		},
		{
			desc: "failure - invalid token",
			req: &simplebank.UpdateUserRequest{
				Username: "orlandorode97",
			},
			info: &grpc.UnaryServerInfo{
				FullMethod: updateUserRPC,
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},

			metadata: metadata.MD{
				metadataAuthorizationHeader: []string{
					"Bearer v2.local.invalid",
				},
			},

			wantGRPCCode: codes.Unauthenticated,
		},
		{
			desc: "failure - permission denied",
			req: &simplebank.UpdateUserRequest{
				Username: "juanito97",
			},
			info: &grpc.UnaryServerInfo{
				FullMethod: updateUserRPC,
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},

			metadata: metadata.MD{
				metadataAuthorizationHeader: []string{
					"Bearer " + accessToken,
				},
			},

			wantGRPCCode: codes.PermissionDenied,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()
			ctx = metadata.NewIncomingContext(ctx, tc.metadata)
			_, err := server.AuthInterceptor()(ctx, tc.req, tc.info, tc.handler)
			resp, ok := status.FromError(err)
			if !ok {
				t.Fatal(err)
			}
			if resp.Code() != tc.wantGRPCCode {
				t.Errorf("response status: got %s want %s", resp.Code(), tc.wantGRPCCode)
			}
		})
	}
}
