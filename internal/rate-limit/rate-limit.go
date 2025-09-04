package ratelimit

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type TokenBucket struct {
	BucketSize int64
	RefillRate time.Duration
	Key        string
	sm         RateLimitStateManager
}

func NewTokenBucket(keyPrefix string, capacity int64, refillRate time.Duration, sm RateLimitStateManager) *TokenBucket {
	key := fmt.Sprintf("%s::rate-limiter::token-bucket::bucket", keyPrefix)

	ticker := time.NewTicker(refillRate)

	sm.Set(key, capacity, 0)

	go func() {
		for {
			select {
			case _ = <-ticker.C:
				{
					sm.Set(key, capacity, 0)
				}
			}
		}
	}()

	return &TokenBucket{
		BucketSize: capacity,
		RefillRate: refillRate,
		Key:        key,
		sm:         sm,
	}
}

func (tb TokenBucket) Allow() bool {
	bucketCounter, err := tb.sm.Get(tb.Key)
	if err != nil || bucketCounter <= 0 {
		return false
	}
	if _, err := tb.sm.Decr(tb.Key); err != nil {
		fmt.Println("Couldn't Decr the bucket counter. Error: ", err.Error())
		return false
	}

	return true
}

func (tb TokenBucket) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !tb.Allow() {
			c.JSON(
				http.StatusTooManyRequests,
				gin.H{},
			)
			c.Abort()
		}
	}
}
