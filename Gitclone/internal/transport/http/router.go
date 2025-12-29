package http

import (
	"net/http"
)

// NewRouter configures all routes and returns the mux
func NewRouter(s *Server) http.Handler {
	mux := http.NewServeMux()

	// Repo list and creation
	mux.HandleFunc("/api/repos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.handleListRepos(w, r)
		} else if r.Method == http.MethodPost {
			s.handleCreateRepo(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Repo-specific routes
	mux.HandleFunc("/api/repos/", s.handleRepoRoutes)

	return corsMiddleware(mux)
}

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

