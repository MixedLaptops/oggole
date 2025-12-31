package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// System health metrics - detect issues as they occur

	// httpRequestsTotal counts total requests (baseline for error rate calculation)
	httpRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oggole_http_requests_total",
		Help: "Total number of HTTP requests",
	})

	// httpErrorsByCode tracks HTTP errors by status code
	httpErrorsByCode = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oggole_http_errors_total",
			Help: "Total HTTP errors by status code",
		},
		[]string{"code"}, // labels: "4xx" or "5xx"
	)

	// requestDuration tracks request duration to detect performance degradation
	requestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "oggole_request_duration_milliseconds",
		Help:    "Request duration distribution in milliseconds",
		Buckets: []float64{10, 50, 100, 250, 500, 1000, 2500},
	})

	// Feature health metrics

	// searchQueries counts total search queries
	searchQueries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oggole_search_queries_total",
		Help: "Total number of search queries performed",
	})

	// searchResultsCount tracks distribution of search result counts
	searchResultsCount = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "oggole_search_results_count",
		Help:    "Distribution of search result counts (quality indicator)",
		Buckets: []float64{0, 1, 5, 10, 25, 50},
	})

	// Operational metrics - for incident response

	// databaseErrors tracks database operation failures
	databaseErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oggole_database_errors_total",
		Help: "Total number of database errors",
	})
)
