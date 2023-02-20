package workers

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type TaskDistributor interface {
	SendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opts ...asynq.Option) error
}

type RedisTaskDistributor struct {
	logger *zap.SugaredLogger
	client *asynq.Client
}

func NewRedisTaskDistributor(clientOpt asynq.RedisClientOpt, logger *zap.SugaredLogger) TaskDistributor {
	client := asynq.NewClient(clientOpt)
	return &RedisTaskDistributor{
		client: client,
		logger: logger,
	}
}
