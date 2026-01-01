package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// System health metrics

	// httpRequestsTotal counts requests by endpoint and status code
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "oggole_http_requests_total",
		Help: "Total HTTP requests by endpoint and status",
	}, []string{"endpoint", "status"})

	// Feature health metrics

	// searchQueries counts total search queries
	searchQueries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oggole_search_queries_total",
		Help: "Total search queries",
	})

	// searchZeroResults counts searches returning zero results
	searchZeroResults = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oggole_search_zero_results_total",
		Help: "Searches returning zero results",
	})

	// Crawler/Indexing metrics

	// pagesIndexed counts pages successfully indexed via batch-pages API
	pagesIndexed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oggole_pages_indexed_total",
		Help: "Total pages indexed via batch-pages API",
	})

	// totalPages tracks current number of pages in database
	totalPages = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "oggole_pages_in_database",
		Help: "Current number of pages in database",
	})

	// Operational metrics

	// databaseErrors tracks database operation failures
	databaseErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oggole_database_errors_total",
		Help: "Total database errors",
	})

	// serviceUp tracks service health (1=up, 0=down)
	serviceUp = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "oggole_service_up",
		Help: "Service health: 1=up, 0=down",
	})
)
