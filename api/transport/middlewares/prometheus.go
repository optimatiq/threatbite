package middlewares

// Code from: https://github.com/0neSe7en/echo-prometheus modified in order to use with echo/v4

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusConfig is a configuration struct for prometheus middleware.
type PrometheusConfig struct {
	Skipper   middleware.Skipper
	Namespace string
}

// DefaultPrometheusConfig default settings for prometheus middleware.
var DefaultPrometheusConfig = PrometheusConfig{
	Skipper:   middleware.DefaultSkipper,
	Namespace: "echo",
}

var (
	echoReqQPS      *prometheus.CounterVec
	echoReqDuration *prometheus.SummaryVec
	echoOutBytes    prometheus.Summary
)

func initCollector(namespace string) {
	echoReqQPS = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_request_total",
			Help:      "HTTP requests processed.",
		},
		[]string{"code", "method", "host", "url"},
	)
	echoReqDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latencies in seconds.",
		},
		[]string{"method", "host", "url"},
	)
	echoOutBytes = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Name:      "http_response_size_bytes",
			Help:      "HTTP response bytes.",
		},
	)
	prometheus.MustRegister(echoReqQPS, echoReqDuration, echoOutBytes)
}

// NewPrometheus returns prometheus middleware with default settings.
func NewPrometheus() echo.MiddlewareFunc {
	return NewPrometheusWithConfig(DefaultPrometheusConfig)
}

// NewPrometheusWithConfig returns prometheus middleware for given configuration.
func NewPrometheusWithConfig(config PrometheusConfig) echo.MiddlewareFunc {
	initCollector(config.Namespace)
	if config.Skipper == nil {
		config.Skipper = DefaultPrometheusConfig.Skipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()
			start := time.Now()

			if err := next(c); err != nil {
				c.Error(err)
			}
			uri := req.URL.Path
			status := strconv.Itoa(res.Status)
			elapsed := time.Since(start).Seconds()
			bytesOut := float64(res.Size)
			echoReqQPS.WithLabelValues(status, req.Method, req.Host, uri).Inc()
			echoReqDuration.WithLabelValues(req.Method, req.Host, uri).Observe(elapsed)
			echoOutBytes.Observe(bytesOut)
			return nil
		}
	}
}
