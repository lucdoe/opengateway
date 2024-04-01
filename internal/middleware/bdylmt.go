package middleware

import (
	"net/http"
)

type BodyLimiter interface {
	LimitRequestBody(next http.Handler) http.Handler
}

type concreteBodyLimiter struct {
	Limit int64
}

func NewBodyLimiter(limit int64) BodyLimiter {
	return &concreteBodyLimiter{Limit: limit}
}

func (bl *concreteBodyLimiter) LimitRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, bl.Limit)
		next.ServeHTTP(w, r)
	})
}
