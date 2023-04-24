package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	corsMux := middlewareCors(mux)

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/", fileServer)

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
