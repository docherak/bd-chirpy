package main

import (
	"database/sql"
	"github.com/docherak/bd-chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	env            string
	jwtSecret      string
	polkApiSecret  string
}

func main() {
	godotenv.Load()

	environment := os.Getenv("PLATFORM")
	if environment == "" {
		log.Fatal("PLATFORM must be set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	polkaApiSecret := os.Getenv("POLKA_KEY")
	if polkaApiSecret == "" {
		log.Fatal("POLKA_KEY must be set")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)

	const port = "8080"
	const filepathRoot = "."

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		env:            environment,
		jwtSecret:      jwtSecret,
		polkApiSecret:  polkaApiSecret,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	// In REST, it's conventional to name all of your endpoints after the resource that they represent and for the name to be plural.
	mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUsersUpdate)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsGetAll)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpsGetSingle)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerChirpsDeleteSingle)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerPolkaEvents)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerTokenRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerTokenRevoke)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on: http://localhost:%s\n", port)
	log.Fatal(srv.ListenAndServe())
}
