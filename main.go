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
	mux.Handle("/metrics", cfg.middlewareMetricsGet())
	mux.Handle("/reset", cfg.middlewareMetricsReset())
	mux.HandleFunc("/healthz", handleHealthz)
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

type IncHitsHandler struct{
	inner 		http.Handler
	cfg 		*apiConfig
}
func (oH *IncHitsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	oH.cfg.fileserverHits.Add(1)
	oH.inner.ServeHTTP(w, r)
}
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return &IncHitsHandler{inner: next, cfg: cfg}
}

type GetHitsHandler struct{
	cfg 		*apiConfig
}
func (oH *GetHitsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Hits: %v", oH.cfg.fileserverHits.Load())))
}
func (cfg *apiConfig) middlewareMetricsGet() http.Handler {
	return &GetHitsHandler{cfg: cfg}
}

type ResetHitsHandler struct{
	cfg 		*apiConfig
}
func (oH *ResetHitsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	oH.cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) middlewareMetricsReset() http.Handler {
	return &ResetHitsHandler{cfg: cfg}
}

func handleHealthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
