package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

type apiConfig struct {
	fileserverHits int
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
	apiRouter.Post("/validate_chirp", validationHandler)
	router.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.metricsHandler)
	router.Mount("/admin", adminRouter)

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
