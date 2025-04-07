package cache

import (
	"expo-open-ota/config"
	"log"
	"sync"
)

type Cache interface {
	Get(key string) string
	Set(key string, value string, ttl *int) error
	Delete(key string)
	Clear() error
}

type CacheType string

const (
	LocalCacheType CacheType = "local"
	RedisCacheType CacheType = "redis"
)

func ResolveCacheType() CacheType {
	cacheType := config.GetEnv("CACHE_MODE")
	if cacheType == "redis" {
		return RedisCacheType
	}
	return LocalCacheType
}

var (
	cacheInstance Cache
	once          sync.Once
)

func GetCache() Cache {
	once.Do(func() {
		cacheType := ResolveCacheType()
		switch cacheType {
		case LocalCacheType:
			cacheInstance = NewLocalCache()
		case RedisCacheType:
			host := config.GetEnv("REDIS_HOST")
			password := config.GetEnv("REDIS_PASSWORD")
			port := config.GetEnv("REDIS_PORT")
			instance := NewRedisCache(host, password, port)
			if instance == nil {
				log.Printf("WARN: Failed to initialize Redis cache (host: %s, port: %s). Falling back to local cache.", host, port)
				cacheInstance = NewLocalCache()
			} else {
				cacheInstance = instance
			}
		default:
			panic("Unknown cache type")
		}
	})
	return cacheInstance
}
