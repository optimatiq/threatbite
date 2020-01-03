package middlewares

import (
	"time"

	"github.com/coinpaprika/ratelimiter"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/common/log"
)

// NewRatelimiter returns rate limiter.
func NewRatelimiter() echo.MiddlewareFunc {
	var maxLimit int64 = 600
	windowSize := 60 * time.Second
	dataStore := ratelimiter.NewMapLimitStore(2*windowSize, windowSize)
	rateLimiter := ratelimiter.New(dataStore, maxLimit, windowSize)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.RealIP()
			limitStatus, err := rateLimiter.Check(key)
			if err != nil {
				log.Error(err)
				return next(c)
			}

			if limitStatus.IsLimited {
				return echo.ErrTooManyRequests
			}

			if err := rateLimiter.Inc(key); err != nil {
				log.Error(err)
			}

			return next(c)
		}
	}
}
