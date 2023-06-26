package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	db "github.com/patchbrain/simple-bank/db/sqlc"
	"github.com/patchbrain/simple-bank/util"
	"github.com/rs/zerolog/log"
)

const VerifyEmailTaskName = "task:send_verify_email"

type VerifyEmailPayload struct {
	Username string
}

func (r *RedisTaskDistributor) EnqueueTask(ctx context.Context, payload VerifyEmailPayload, opts ...asynq.Option) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("fail to marshal payload: %w", err)
	}
	task := asynq.NewTask(VerifyEmailTaskName, data)
	taskInfo, err := r.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("fail to enqueue task: %w", err)
	}

	log.Info().Str("username", payload.Username).
		Str("task_state", taskInfo.State.String()).
		Int("task_max_retry", taskInfo.MaxRetry).
		Str("task_name", VerifyEmailTaskName).
		Str("task_type", taskInfo.Type).
		Str("task_queue", taskInfo.Queue).
		Str("task_group", taskInfo.Group).
		Msg("distribute a verify email task")

	return nil
}

func (r *RedisTaskProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload VerifyEmailPayload
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("fail to unmarshal payload: %w", asynq.SkipRetry)
	}

	user, err := r.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("cannot find the user: %w", err)
		}
		return fmt.Errorf("fail to get user: %w", err)
	}

	// 创建email_verify
	verify, err := r.store.CreateEmailVerify(ctx, db.CreateEmailVerifyParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.RandomString(32),
	})
	if err != nil {
		return fmt.Errorf("failed to create email verify: %w", err)
	}

	path := "/v1/verify_email"
	query := fmt.Sprintf("id=%d&secret_code=%s", verify.ID, verify.SecretCode)
	urlVerify := fmt.Sprintf("localhost:8080%s?%s", path, query)
	subject := "Welcome to Simple Bank"
	to := []string{user.Email}
	content := fmt.Sprintf(`Hello %s,<br/>
	Thank you for registering with us!<br/>
	Please <a href="%s">click here</a> to verify your email.<br/>`, user.Username, urlVerify)
	err = r.mailSender.SendEmail(subject, to, content, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("fail to send email: %w", err)
	}

	log.Info().Str("username", user.Username).
		Str("email", user.Email).
		Msg("send a verify email")

	return nil
}
