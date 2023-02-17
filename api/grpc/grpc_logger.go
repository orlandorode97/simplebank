package grpc

import (
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) LoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		now := time.Now()
		result, err := handler(ctx, req)
		duration := time.Since(now)
		statusCode := codes.Unknown
		if fromStatus, ok := status.FromError(err); ok {
			statusCode = fromStatus.Code()
		}

		if err != nil {
			s.logger.Errorw("recevied gRPC request: ",
				zap.Error(err),
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
				zap.String("code", statusCode.String()),
			)
			return result, err
		}

		s.logger.Infow("recevied gRPC request",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.String("code", statusCode.String()),
		)

		return result, err
	}
}
