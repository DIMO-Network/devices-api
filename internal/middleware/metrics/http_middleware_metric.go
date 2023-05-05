package metrics

import (
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

func HTTPMetricsPrometheusMiddleware(c *fiber.Ctx) error {
	start := time.Now()
	method := c.Route().Method

	err := c.Next()
	status := fiber.StatusInternalServerError
	if err != nil {
		if e, ok := err.(*fiber.Error); ok {
			// Get correct error code from fiber.Error type
			status = e.Code
		}

	} else {
		status = c.Response().StatusCode()
	}

	path := c.Route().Name
	statusCode := strconv.Itoa(status)

	appmetrics.HTTPRequestCount.WithLabelValues(method, path, statusCode).Inc()

	defer func() {
		appmetrics.HTTPResponseTime.With(prometheus.Labels{
			"method": method,
			"path":   path,
			"status": statusCode,
		}).Observe(time.Since(start).Seconds())
	}()

	return err
}
