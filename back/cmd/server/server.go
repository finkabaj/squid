package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/finkabaj/squid/back/internal/config"
	"github.com/finkabaj/squid/back/internal/controller"
	myMiddleware "github.com/finkabaj/squid/back/internal/middleware"
	"github.com/finkabaj/squid/back/internal/websocket"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"

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
	r.Use(cors.Default().Handler)

	wsServer := websocket.NewServer()

	controller.NewKanbanController(wsServer).RegisterKanbanRoutes(r)
	controller.NewAuthController().RegisterAuthRoutes(r)

	r.With(myMiddleware.ValidateJWT).HandleFunc("/ws", wsServer.HandleWs)

	server := http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Data.Host, config.Data.Port),
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  600 * time.Second,
	}

	logger.Logger.Info().Msgf("Server starting on %s", server.Addr)
	if err = server.ListenAndServe(); err != nil {
		logger.Logger.Fatal().Err(err).Msg("Error starting server")
	}
}
