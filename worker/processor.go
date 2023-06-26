package worker

import (
	"context"
	"github.com/hibiken/asynq"
	db "github.com/patchbrain/simple-bank/db/sqlc"
	"github.com/patchbrain/simple-bank/mail"
	"github.com/rs/zerolog/log"
)

const (
	QueueNameCritical = "critical"
	QueueNameDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	ProcessTask(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server     *asynq.Server
	store      db.Store
	mailSender mail.EmailSender
}

func NewRedisTaskProcessor(opt asynq.RedisConnOpt, store db.Store, mailSender mail.EmailSender) TaskProcessor {
	server := asynq.NewServer(opt, asynq.Config{
		Queues: map[string]int{
			QueueNameCritical: 10,
			QueueNameDefault:  5,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			log.Error().Err(err).Str("type", task.Type()).Bytes("payload", task.Payload()).Msg("fail to process task")
		}),
		Logger: NewLogger(),
	})
	p := RedisTaskProcessor{
		server:     server,
		store:      store,
		mailSender: mailSender,
	}
	return &p
}

func (r *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(VerifyEmailTaskName, r.ProcessTask)
	return r.server.Start(mux)
}
