package consumer

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"aegis/consts"
	redisinfra "aegis/infra/redis"

	goredis "github.com/redis/go-redis/v9"
)

type namespaceLockState struct {
	EndTime int64
	TraceID string
}

type namespaceLockStore struct {
	client *redisinfra.Gateway
}

func newNamespaceLockStore(client *redisinfra.Gateway) namespaceLockStore {
	return namespaceLockStore{client: client}
}

func (s namespaceLockStore) key(namespace string) string {
	return fmt.Sprintf(consts.NamespaceKeyPattern, namespace)
}

func (s namespaceLockStore) read(ctx context.Context, namespace string) (*namespaceLockState, error) {
	endTimeStr, err := s.client.HashGet(ctx, s.key(namespace), "end_time")
	if err != nil && err != goredis.Nil {
		return nil, err
	}

	traceID, err := s.client.HashGet(ctx, s.key(namespace), "trace_id")
	if err != nil && err != goredis.Nil {
		return nil, err
	}

	if endTimeStr == "" {
		return &namespaceLockState{TraceID: traceID}, nil
	}

	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		return nil, err
	}

	return &namespaceLockState{EndTime: endTime, TraceID: traceID}, nil
}

func (s namespaceLockStore) readFromHash(reader goredis.HashCmdable, ctx context.Context, namespace string) (*namespaceLockState, error) {
	endTimeStr, err := reader.HGet(ctx, s.key(namespace), "end_time").Result()
	if err != nil && err != goredis.Nil {
		return nil, err
	}

	traceID, err := reader.HGet(ctx, s.key(namespace), "trace_id").Result()
	if err != nil && err != goredis.Nil {
		return nil, err
	}

	if endTimeStr == "" {
		return &namespaceLockState{TraceID: traceID}, nil
	}

	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		return nil, err
	}

	return &namespaceLockState{EndTime: endTime, TraceID: traceID}, nil
}

func (s namespaceLockStore) write(ctx context.Context, namespace string, endTime int64, traceID string) error {
	return s.client.HashSet(ctx, s.key(namespace), map[string]any{
		"end_time": endTime,
		"trace_id": traceID,
	})
}

func (s namespaceLockStore) acquire(ctx context.Context, namespace string, endTime time.Time, traceID string, now time.Time) error {
	return s.client.Watch(ctx, func(tx *goredis.Tx) error {
		state, err := s.readFromHash(tx, ctx, namespace)
		if err != nil {
			return err
		}
		if state.TraceID != "" && state.TraceID != traceID && now.Unix() < state.EndTime {
			return fmt.Errorf("namespace %s is locked by %s until %v",
				namespace, state.TraceID, time.Unix(state.EndTime, 0).Format(time.RFC3339))
		}
		_, err = tx.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
			pipe.HSet(ctx, s.key(namespace), "end_time", endTime.Unix())
			pipe.HSet(ctx, s.key(namespace), "trace_id", traceID)
			return nil
		})
		return err
	}, s.key(namespace))
}

func (s namespaceLockStore) release(ctx context.Context, namespace, traceID string, releasedAt time.Time) error {
	state, err := s.read(ctx, namespace)
	if err != nil && err != goredis.Nil {
		return fmt.Errorf("failed to get current trace_id: %v", err)
	}
	if state != nil && state.TraceID != traceID && state.TraceID != "" {
		return fmt.Errorf("cannot release lock: namespace %s is not owned by trace_id %s (current owner: %s)",
			namespace, traceID, state.TraceID)
	}
	return s.write(ctx, namespace, releasedAt.Unix(), "")
}

func (s namespaceLockStore) isActive(ctx context.Context, namespace string, now time.Time) (bool, error) {
	state, err := s.read(ctx, namespace)
	if err == goredis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if state.TraceID == "" {
		return false, nil
	}
	return now.Unix() < state.EndTime, nil
}
