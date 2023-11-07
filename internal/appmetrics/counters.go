package appmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	DrivlyIngestTotalOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_drivly_ingest_success_ops_total",
		Help: "Total successful Drivly used",
	})

	// Chat GPT Metrics
	OpenAITotalCallsOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_error_codes_openai_requests_total",
		Help: "Total number of calls to Open AI ChatGPT",
	})
	OpenAITotalFailedCallsOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_error_codes_openai_failed_calls_total",
		Help: "Total number of failed calls to Open AI ChatGPT",
	})
	OpenAITotalTokensUsedOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_error_codes_openai_total_token_used",
		Help: "Total number of failed calls to Open AI ChatGPT",
	})
	OpenAIResponseTimeOps = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "devices_api_error_codes_openai_request_duration_seconds",
		Help:    "Response duration of OpenAI ChatGPT in seconds",
		Buckets: []float64{0.1, 0.15, 0.2, 0.25, 0.3, 0.5, 0.7, 0.9, 10},
	}, []string{"status"})

	GRPCRequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "devices_api_grpc_request_count",
			Help: "The total number of requests served by the GRPC Server",
		},
		[]string{"method", "status"},
	)

	GRPCPanicCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "devices_api_grpc_panic_count",
			Help: "The total number of panics served by the GRPC Server",
		},
	)

	GRPCResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "devices_api_grpc_response_time",
			Help:    "The response time distribution of the GRPC Server",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "status"},
	)

	HTTPRequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "devices_api_http_request_count",
			Help: "The total number of requests served by the Http Server",
		},
		[]string{"method", "path", "status"},
	)

	HTTPResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "devices_api_http_response_time",
			Help:    "The response time distribution of the Http Server",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	FingerprintRequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "devices_api_fingerprint_request_count",
			Help: "The total number of Fingerprint requests",
		},
		[]string{"protocol", "status"},
	)
)
