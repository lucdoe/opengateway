package middlewares

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucdoe/capstone_gateway/internal"
)

const (
	RateLimitKeyPrefix    = "rate-limit:"
	HeaderRateLimit       = "X-RateLimit-Limit"
	HeaderRateLimitRemain = "X-RateLimit-Remaining"
	HeaderRateLimitReset  = "X-RateLimit-Reset"
	MaxRequestsPerMinute  = 45
	RateLimitWindow       = time.Minute
	ThreeMegaByte         = 3145728
)

func BodyLimit(c *gin.Context) {
	var b []byte

	if c.Request.Body != nil {
		b, _ = io.ReadAll(c.Request.Body)
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(b))

	size := int64(len(b))

	if size > ThreeMegaByte {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request too big"})
		c.Abort()
		return
	}

	c.Next()
}

func RateLimit(c *gin.Context) {
	count, err := increaseRateLimitCounter(c.ClientIP())
	if err != nil {
		c.Error(errors.New("redis not started, could not process Rate-Limit"))
		return
	}

	if count > MaxRequestsPerMinute {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
		c.Error(errors.New("too many requests, please try again later")).SetType(gin.ErrorTypePublic)
		c.Abort()
		return
	}

	setRateLimitHeaders(c, count)
	c.Next()
}

func increaseRateLimitCounter(ip string) (int64, error) {
	ctx := context.Background()
	key := RateLimitKeyPrefix + ip

	count, err := internal.RDB.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if count == 1 {
		internal.RDB.Expire(ctx, key, RateLimitWindow)
	}

	return count, nil
}

func setRateLimitHeaders(c *gin.Context, count int64) {
	c.Header(HeaderRateLimit, strconv.Itoa(MaxRequestsPerMinute))
	c.Header(HeaderRateLimitRemain, strconv.Itoa(MaxRequestsPerMinute-int(count)))
	c.Header(HeaderRateLimitReset, strconv.Itoa(int(RateLimitWindow.Seconds())))
}
