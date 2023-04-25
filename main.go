package main

import (
	"net/http"
)

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()
	corsMux := middlewareCors(mux)

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/", fileServer)

	mux.HandleFunc("/healthz", readinessHandler)

	assetsServer := http.FileServer(http.Dir("./assets/"))
	mux.Handle("/birbs/", http.StripPrefix("/birbs/", assetsServer))

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: corsMux,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
