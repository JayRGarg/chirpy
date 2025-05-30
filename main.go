package main

import (
	"fmt"
	"log"
	"sync/atomic"
	"net/http"
)

func main() {
	filepathRoot := "."
	port := "8080"
	mux := http.NewServeMux()

	cfg := &apiConfig{}

	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.Handle("GET /admin/metrics", cfg.middlewareMetricsGet())
	mux.Handle("POST /admin/reset", cfg.middlewareMetricsReset())
	mux.HandleFunc("GET /api/healthz", handleHealthz)
	server := http.Server{
		Addr: 		":"+port,
		Handler: 	mux,
	}
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMetricsGet() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		metricsTemplate := `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
		`
		responseBody := fmt.Sprintf(metricsTemplate, cfg.fileserverHits.Load())
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(responseBody))
		if err != nil {
			log.Printf("Error writing metrics response to client: %v", err)
			return 
		}
	})
}

func (cfg *apiConfig) middlewareMetricsReset() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Store(0)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(fmt.Sprintf("Hits reset to 0")))
		if err != nil {
			log.Printf("Error writing metrics reset status to client: %v", err)
			return
		}
		
	})
}

func handleHealthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Printf("Error writing health response to client: %v", err)
		return
	}
}
