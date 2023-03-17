package workers

import (
	"context"
	"html/template"

	"github.com/hibiken/asynq"
	"github.com/orlandorode97/simple-bank/mail"
	"github.com/orlandorode97/simple-bank/store"
	"go.uber.org/zap"
)

const (
	QueueCritial = "critical"
	QueueDefault = "defailt"
)

var tlp *template.Template

type TaskProcessor interface {
	Start() error
	SendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

type RedistTaskProcessor struct {
	server *asynq.Server
	store  store.Store
	logger *zap.SugaredLogger
	sender mail.EmailSender
}

// NewRedistTaskProcessor returns a *TaskProcessor
func NewRedistTaskProcessor(r asynq.RedisConnOpt, store store.Store, logger *zap.SugaredLogger, sender mail.EmailSender) TaskProcessor {

	tlp = template.Must(template.ParseFiles("templates/verification_email.gohtml"))

	taskProcessor := &RedistTaskProcessor{
		store:  store,
		logger: logger,
		sender: sender,
	}

	server := asynq.NewServer(r, asynq.Config{
		Queues: map[string]int{
			QueueCritial: 10,
			QueueDefault: 5,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			taskProcessor.logger.Errorw("unable to process task",
				zap.Error(err),
				zap.String("type", task.Type()),
				zap.ByteString("payload", task.Payload()))
		}),
	})

	taskProcessor.server = server
	return taskProcessor
}

// Start starts an asynq server by providing incoming handlers.
func (r *RedistTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(taskSendVerifyEmail, r.SendVerifyEmail)
	return r.server.Start(mux)
}
