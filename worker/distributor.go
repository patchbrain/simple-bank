package worker

import (
	"context"
	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	EnqueueTask(ctx context.Context, payload VerifyEmailPayload, opts ...asynq.Option) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisConnOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{client: client}
}
