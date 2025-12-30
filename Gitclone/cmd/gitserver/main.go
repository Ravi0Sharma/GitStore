package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gitclone/internal/metadata"
	httptransport "gitclone/internal/transport/http"
)

const defaultPort = "8080"
const defaultRepoBase = "./data/repos"
const defaultDBPath = "./data/db"

func main() {
	// Load configuration from environment
	repoBase := os.Getenv("GITSTORE_REPO_BASE")
	if repoBase == "" {
		repoBase = defaultRepoBase
	}

	dbPath := os.Getenv("GITSTORE_DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Convert paths to absolute for consistency
	repoBaseAbs, err := filepath.Abs(repoBase)
	if err != nil {
		log.Fatalf("Failed to get absolute path for repo base: %v", err)
	}
	repoBase = repoBaseAbs

	dbPathAbs, err := filepath.Abs(dbPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path for db: %v", err)
	}
	dbPath = dbPathAbs

	// Initialize metadata store
	metaStore, err := metadata.NewStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize metadata store: %v", err)
	}
	defer metaStore.Close()

	// Create server instance
	server := httptransport.NewServer(repoBase, metaStore)

	log.Printf("Repository base directory (absolute): %s", repoBase)
	log.Printf("Metadata database path (absolute): %s", dbPath)

	// Setup router
	handler := httptransport.NewRouter(server)

	log.Printf("Starting GitStore server on port %s, repo base: %s", port, repoBase)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
