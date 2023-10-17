package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucdoe/capstone/internal/app/databases"
)

const (
	RateLimitKeyPrefix    = "rate-limit:"
	HeaderRateLimit       = "X-RateLimit-Limit"
	HeaderRateLimitRemain = "X-RateLimit-Remaining"
	HeaderRateLimitReset  = "X-RateLimit-Reset"
	MaxRequestsPerMinute  = 45
	RateLimitWindow       = time.Minute
)

func getRateLimitKey(ip string) string {
	return RateLimitKeyPrefix + ip
}

func increaseRateLimitCounter(ip string) (int64, error) {
	ctx := context.Background()
	key := getRateLimitKey(ip)
	count, err := databases.RDB.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// If it's the first request, set an expiration time for it
	if count == 1 {
		databases.RDB.Expire(ctx, key, RateLimitWindow)
	}

	return count, nil
}

func isRateLimitExceeded(count int64) bool {
	return count > MaxRequestsPerMinute
}

func setRateLimitHeaders(c *gin.Context, count int64) {
	c.Header(HeaderRateLimit, strconv.Itoa(MaxRequestsPerMinute))
	c.Header(HeaderRateLimitRemain, strconv.Itoa(MaxRequestsPerMinute-int(count)))
	c.Header(HeaderRateLimitReset, strconv.Itoa(int(RateLimitWindow.Seconds())))
}

func RateLimit(c *gin.Context) {
	count, err := increaseRateLimitCounter(c.ClientIP())
	if err != nil {
		c.Error(errors.New("Redis not started, could not process Rate-Limit."))
		return
	}

	if isRateLimitExceeded(count) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
		c.Error(errors.New("Too many requests, please try again later.")).SetType(gin.ErrorTypePublic)
		c.Abort()
		return
	}

	setRateLimitHeaders(c, count)
	c.Next()
}
