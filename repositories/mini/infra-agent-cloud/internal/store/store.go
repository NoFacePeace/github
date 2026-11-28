package store

import "infra-agent-cloud/internal/store/redis"

type Store struct {
	redis.Redis
}
