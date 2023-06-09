package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/emilmalmsten/chirpy/internal/jsonDB"
	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits int
	DB             *jsonDB.DB
	jwtSecret      string
	polkaApiKey    string
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func filterProfanity(message string) string {
	//message = strings.ToLower(message)
	bannedWords := []string{"kerfuffle", "sharbert", "fornax"}
	const replacement = "****"

	for _, word := range bannedWords {
		message = strings.Replace(message, word, replacement, -1)
	}
	return message
}

func main() {
	godotenv.Load()
	//jwtSecret := os.Getenv("JWT_SECRET")
	jwtSecret := "test123"
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	//polkaApiKey := os.Getenv("POLKA_API_KEY")
	polkaApiKey := "123test"
	if polkaApiKey == "" {
		log.Fatal("POLKA_API_KEY environment variable is not set")
	}
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	db, err := jsonDB.NewDB(exPath + "/db.json")
	if err != nil {
		panic(err)
	}

	apiCfg := apiConfig{
		fileserverHits: 0,
		DB:             db,
		jwtSecret:      jwtSecret,
		polkaApiKey:    polkaApiKey,
	}

	router := chi.NewRouter()

	fileServer := apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))
	router.Mount("/", fileServer)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", readinessHandler)

	apiRouter.Post("/chirps", apiCfg.handlerPostChirp)
	apiRouter.Get("/chirps", apiCfg.handlerGetChirps)
	apiRouter.Get("/chirps/{chirpID}", apiCfg.handlerGetChirpById)
	apiRouter.Delete("/chirps/{chirpID}", apiCfg.handlerDeleteChirp)

	apiRouter.Post("/users", apiCfg.handlerUsersCreate)
	apiRouter.Put("/users", apiCfg.handlerUsersUpdate)
	apiRouter.Post("/login", apiCfg.handlerUsersLogin)
	apiRouter.Post("/refresh", apiCfg.handlerRefresh)
	apiRouter.Post("/revoke", apiCfg.handlerRevoke)

	apiRouter.Post("/polka/webhooks", apiCfg.handlerUpgradeMembership)

	router.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.metricsHandler)
	router.Mount("/admin", adminRouter)

	corsMux := middlewareCors(router)

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: corsMux,
	}

	fmt.Printf("server running on: %s\n", server.Addr)

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
