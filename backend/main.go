package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/abdul/vortyx/backend/internal/server"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// Connect to Database
	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Initialize Server
	handler := server.NewRouter(dbPool)

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Starting server on %s\n", addr)

	// Start HTTP Server
	err = http.ListenAndServe(addr, handler)
	if err != nil {
		log.Fatal(err)
	}
}
