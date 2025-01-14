package config

import (
	"github.com/pkg/errors"
	"os"
	"strconv"
)

type Config struct {
	Env                   string
	Host                  string
	Port                  int
	SaltRounds            int
	RefreshTokenExpHours  int
	AccessTokenExpMinutes int
	FilenameLogOutput     string
	JWTSecret             []byte
	PostgresHost          string
	PostgresPort          int
	PostgresUser          string
	PostgresPassword      string
	PostgresDatabase      string
}

var Data Config

func Initialize() error {
	port := os.Getenv("PORT")
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return errors.Wrap(err, "port is not a number")
	}

	saltRounds := os.Getenv("SALT_ROUNDS")
	saltRoundsInt, err := strconv.Atoi(saltRounds)
	if err != nil {
		return errors.Wrap(err, "salt rounds is not a number")
	}

	refreshTokenExpHours := os.Getenv("REFRESH_TOKEN_EXP_H")
	refreshTokenExpHoursInt, err := strconv.Atoi(refreshTokenExpHours)
	if err != nil {
		return errors.Wrap(err, "refresh token exp hours is not a number")
	}

	accessTokenExpMinutes := os.Getenv("ACCESS_TOKEN_EXP_M")
	accessTokenExpMinutesInt, err := strconv.Atoi(accessTokenExpMinutes)
	if err != nil {
		return errors.Wrap(err, "auth token exp minutes is not a number")
	}

	postgresPort := os.Getenv("POSTGRES_PORT")
	postgresPortInt, err := strconv.Atoi(postgresPort)
	if err != nil {
		return errors.Wrap(err, "postgres port is not a number")
	}

	Data = Config{
		Env:                   os.Getenv("ENV"),
		Host:                  os.Getenv("HOST"),
		Port:                  portInt,
		SaltRounds:            saltRoundsInt,
		RefreshTokenExpHours:  refreshTokenExpHoursInt,
		AccessTokenExpMinutes: accessTokenExpMinutesInt,
		FilenameLogOutput:     os.Getenv("FNAME_LOG_OUT"),
		JWTSecret:             []byte(os.Getenv("JWT_SECRET")),
		PostgresHost:          os.Getenv("POSTGRES_HOST"),
		PostgresPort:          postgresPortInt,
		PostgresUser:          os.Getenv("POSTGRES_USER"),
		PostgresPassword:      os.Getenv("POSTGRES_PASSWORD"),
		PostgresDatabase:      os.Getenv("POSTGRES_DB"),
	}

	return nil
}
