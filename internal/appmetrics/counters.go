package appmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SmartcarIngestTotalOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_smartcar_ingest_ops_total",
		Help: "Total smartcar ingest events started",
	})
	SmartcarIngestSuccessOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_smartcar_ingest_success_ops_total",
		Help: "Total succesful smartcar ingest events completed",
	})

	AutoPiIngestTotalOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_autopi_ingest_ops_total",
		Help: "Total AutoPi ingest events started",
	})
	AutoPiIngestSuccessOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_autopi_ingest_success_ops_total",
		Help: "Total successful AutoPi ingest events completed",
	})
	DrivlyIngestTotalOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_drivly_ingest_success_ops_total",
		Help: "Total successful Drivly used",
	})
	BlackbookRequestTotalOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_blackbook_request_success_ops_total",
		Help: "Total successful Blackbook used",
	})
	// Chat GPT Metrics
	OpenAITotalCallsOps = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "devices_api_openai_chatgpt_ops_total",
		Help:      "Total number of calls to Open AI ChatGPT",
		Namespace: "devices_api",
		Subsystem: "error_codes",
	})
	OpenAITotalFailedCallsOps = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "devices_api_openai_chatgpt_ops_failed_calls_total",
		Help:      "Total number of failed calls to Open AI ChatGPT",
		Namespace: "devices_api",
		Subsystem: "error_codes",
	})
	OpenAITotalTokensUsedOps = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "devices_api_openai_chatgpt_ops_total_token_used",
		Help:      "Total number of failed calls to Open AI ChatGPT",
		Namespace: "devices_api",
		Subsystem: "error_codes",
	})
	OpenAIResponseTimeOps = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "devices_api_openai_chatgpt_ops_request_duration_seconds",
		Help:      "Response duration of OpenAI ChatGPT in seconds",
		Namespace: "devices_api",
		Subsystem: "error_codes",
		Buckets:   []float64{0.1, 0.15, 0.2, 0.25, 0.3, 0.5, 0.7, 0.9, 10},
	}, []string{"status", "method"})
)
