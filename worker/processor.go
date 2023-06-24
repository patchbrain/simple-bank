package worker

import (
	"context"
	"github.com/hibiken/asynq"
	db "github.com/patchbrain/simple-bank/db/sqlc"
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
	server *asynq.Server
	store  db.Store
}

func NewRedisTaskProcessor(opt asynq.RedisConnOpt, store db.Store) TaskProcessor {
	server := asynq.NewServer(opt, asynq.Config{
		Queues: map[string]int{
			QueueNameCritical: 10,
			QueueNameDefault:  5,
		},
	})
	p := RedisTaskProcessor{
		server: server,
		store:  store,
	}
	return &p
}

func (r *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(VerifyEmailTaskName, r.ProcessTask)
	return r.server.Start(mux)
}
