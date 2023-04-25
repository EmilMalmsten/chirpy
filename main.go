package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hits: %v", cfg.fileserverHits)
}

func main() {
	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	router := chi.NewRouter()

	fileServer := apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))
	router.Mount("/", fileServer)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", readinessHandler)
	apiRouter.Get("/metrics", apiCfg.metricsHandler)
	router.Mount("/api", apiRouter)

	corsMux := middlewareCors(router)

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: corsMux,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
