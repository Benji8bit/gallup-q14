package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/sapiens-solutions/gallup-q14/internal/analytics"
	"github.com/sapiens-solutions/gallup-q14/internal/api"
	"github.com/sapiens-solutions/gallup-q14/internal/config"
	"github.com/sapiens-solutions/gallup-q14/internal/delivery"
	"github.com/sapiens-solutions/gallup-q14/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := store.New(cfg.DBPath, time.Now)
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	defer db.Close()

	router := chi.NewRouter()
	router.Use(chimw.RequestID)
	router.Use(chimw.RealIP)
	router.Use(chimw.Logger)
	router.Use(chimw.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSOrigin},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	h := api.NewHandler(db, analytics.New(), cfg.AdminPassword, cfg.DBPath)
	h.RegisterRoutes(router)

	delivery.StartPeriodicSync(cfg.DBPath, cfg.DeliverySyncInterval, log.Default())

	distPath := filepath.Clean(filepath.Join("..", "frontend", "dist"))
	if pathExists(distPath) {
		router.Get("/*", spaStaticHandler(distPath))
		log.Printf("static frontend enabled from %s", distPath)
	} else {
		router.Get("/", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte("Gallup Q14 backend is running"))
		})
	}

	addr := ":" + cfg.Port
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func pathExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func spaStaticHandler(distPath string) http.HandlerFunc {
	fileServer := http.StripPrefix("/", http.FileServer(http.Dir(distPath)))
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		cleanPath := strings.TrimPrefix(r.URL.Path, "/")
		target := filepath.Join(distPath, filepath.Clean(cleanPath))
		target = filepath.Clean(target)
		if !strings.HasPrefix(target, filepath.Clean(distPath)) {
			http.NotFound(w, r)
			return
		}

		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
	}
}
