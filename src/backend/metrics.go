package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// httpRequestsTotal counts total HTTP requests
	httpRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oggole_http_requests_total",
		Help: "Total number of HTTP requests",
	})

	// loginAttempts tracks login attempts by status (success/failure)
	loginAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oggole_login_attempts_total",
			Help: "Total login attempts by status",
		},
		[]string{"status"}, // labels: "success" or "failure"
	)

	// activeSessions tracks current number of active user sessions
	activeSessions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "oggole_active_sessions",
		Help: "Number of active user sessions",
	})

	// searchQueries counts total search queries
	searchQueries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oggole_search_queries_total",
		Help: "Total number of search queries",
	})
)
