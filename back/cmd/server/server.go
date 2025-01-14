package main

import (
	"fmt"
	"github.com/finkabaj/squid/back/internal/config"
	"github.com/finkabaj/squid/back/internal/controller"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"os"
	"time"

	"github.com/finkabaj/squid/back/internal/repository"
	"github.com/finkabaj/squid/back/internal/types"

	"github.com/finkabaj/squid/back/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		fmt.Printf("Error loading env variables: %s", err.Error())
		os.Exit(1)
	}

	if err = config.Initialize(); err != nil {
		fmt.Printf("Error initializing config: %s", err.Error())
	}

	fs, err := os.OpenFile(config.Data.FilenameLogOutput, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %s with error %s", config.Data.FilenameLogOutput, err.Error())
		os.Exit(1)
	}
	defer fs.Close()
	logger.InitLogger(fs)

	dbCredentials := types.DBCredentials{
		Host:     config.Data.PostgresHost,
		Port:     config.Data.PostgresPort,
		User:     config.Data.PostgresUser,
		Password: config.Data.PostgresPassword,
		Database: config.Data.PostgresDatabase,
	}

	if err = repository.Connect(dbCredentials); err != nil {
		logger.Logger.Fatal().Err(err).Stack().Msg("Error connecting to database")
	}

	defer repository.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	controller.RegisterAuthRoutes(r)

	server := http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Data.Host, config.Data.Port),
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  600 * time.Second,
	}

	server.ListenAndServe()
	defer server.Close()
}
