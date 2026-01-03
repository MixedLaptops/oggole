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

	// Feature health metrics - Search tracking

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

	// searchQueryTerms tracks what users search for (top search terms)
	searchQueryTerms = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "oggole_search_query_terms_total",
		Help: "Count of each search query term",
	}, []string{"query"})

	// searchesPerSession tracks distribution of searches per user session
	searchesPerSession = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "oggole_searches_per_session",
		Help:    "Distribution of number of searches per user session before leaving",
		Buckets: []float64{1, 2, 3, 5, 10, 20, 50},
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

	// serviceUp tracks service health (1=up, 0=down)
	serviceUp = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "oggole_service_up",
		Help: "Service health: 1=up, 0=down",
	})
)
