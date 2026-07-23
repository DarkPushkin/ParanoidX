// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

import (
	_ "embed"
	"net/http"
)

//go:embed dashboard.html
var dashboardHTML string


// DashboardHandler handles the DashboardHandler HTTP request.
// DashboardHandler serves the embedded HTML dashboard.
func DashboardHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(dashboardHTML))
	}
}
