package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/finkabaj/squid/back/internal/repository"
	"github.com/finkabaj/squid/back/internal/types"

	"github.com/finkabaj/squid/back/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		fmt.Printf("Error loading env variables: %s", err.Error())
		os.Exit(1)
	}

	fnameLogOut := os.Getenv("FNAME_LOG_OUT")
	fs, err := os.OpenFile(fnameLogOut, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %s with error %s", fnameLogOut, err.Error())
		os.Exit(1)
	}
	defer fs.Close()
	logger.InitLogger(fs)

	dbCredentials := types.DBCredentials{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Database: os.Getenv("POSTGRES_DB"),
	}

	if err = repository.Connect(dbCredentials); err != nil {
		logger.Logger.Fatal().Err(err).Stack().Msg("Error connecting to database")
	}

	defer repository.Close()

	r := chi.NewRouter()
	if os.Getenv("ENV") == "development" {
		r.Use(middleware.Logger)
	} else {
		r.Use(middleware.Recoverer)
	}
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	server := http.Server{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  600 * time.Second,
	}

	server.ListenAndServe()
	defer server.Close()
}
