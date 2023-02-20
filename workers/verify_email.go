package workers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/orlandorode97/simple-bank/mail"
	"go.uber.org/zap"
)

const (
	taskSendVerifyEmail = "task:send_verify_email"

	emailSubject = "Verification email"
)

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// SendVerifyEmail of RedisTaskDistributor creates a task to enqueue.
func (r *RedisTaskDistributor) SendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opts ...asynq.Option) error {
	// marshal the payload of the task to enqueue
	jsonPaylod, err := json.Marshal(&payload)
	if err != nil {
		return err
	}

	// create a new task of the name task:send_verify_email
	task := asynq.NewTask(taskSendVerifyEmail, jsonPaylod, opts...)
	// enqueue task with context
	_, err = r.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return err
	}

	r.logger.Infow("task enqueued",
		zap.String("type", task.Type()),
		zap.ByteString("payload", task.Payload()))

	return nil
}

// SendVerifyEmail of RedistTaskProcessor processes the enqueued task and sends verification email.
func (r *RedistTaskProcessor) SendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	payload := PayloadSendVerifyEmail{}
	// Unmarshal payload of the task
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("unable to unmarshal payload: %w", asynq.SkipRetry)
	}

	_, err = r.store.GetUser(ctx, payload.Username)
	if err != nil {
		// Aboid SkipRetry to retry task processing
		return fmt.Errorf("unable to get user: %w", err)
	}

	var body bytes.Buffer
	data := &mail.EmailBody{
		Username: payload.Username,
		URL:      "http://google.com",
		Today:    time.Now(),
	}

	if err = tlp.ExecuteTemplate(&body, "verification_email.gohtml", &data); err != nil {
		return fmt.Errorf("unable to execute verification_email template: %w", err)
	}

	if err = r.sender.SendEmail(emailSubject, body.String(), []string{payload.Email}, nil, nil, nil); err != nil {
		return fmt.Errorf("unable to send verification email: %w", err)
	}

	//Send email to user
	r.logger.Infow("task processed",
		zap.String("type", task.Type()),
		zap.ByteString("payload", task.Payload()))

	return nil
}
