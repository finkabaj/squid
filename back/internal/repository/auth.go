package repository

import (
	"context"
	"time"

	"github.com/finkabaj/squid/back/internal/types"
	"github.com/finkabaj/squid/back/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

func CreateUser(ctx context.Context, id *string, passwordHash *string, user *types.RegisterUser) (types.User, error) {
	if id == nil || passwordHash == nil || user == nil {
		return types.User{}, errors.New("All arguments must be not nil")
	}

	newUser, err := insertReturning[types.User](ctx, `
        INSERT INTO "users" ("id", "username", "firstName", "lastName", "dateOfBirth", "email", "passwordHash")
        VALUES ($1, $2, $3, $4, $5, $6, $7) 
        RETURNING *
    `, *id, user.Username, user.FirstName, user.LastName, user.DateOfBirth, user.Email, passwordHash)
	if err != nil {
		return types.User{}, err
	}

	return newUser, nil
}

func GetUser(ctx context.Context, id *string, email *string) (types.User, error) {
	if id == nil && email == nil {
		return types.User{}, errors.New("At least one arguments must not be nil")
	}

	var err error
	var query string
	var user types.User

	if id != nil {
		query = `SELECT * FROM "users" WHERE "id" = $1`
		user, err = selectOneReturning[types.User](ctx, query, *id)
	} else {
		query = `SELECT * FROM "users" WHERE "email" = $1`
		user, err = selectOneReturning[types.User](ctx, query, *email)
	}

	return user, err
}

func DeleteUser(ctx context.Context, id *string) error {
	return simpleDelete(ctx, id, "users", "id")
}

func CreateRefreshToken(ctx context.Context, id *string, userID *string, expiresAt *time.Time) (types.RefreshToken, error) {
	if id == nil || userID == nil || expiresAt == nil {
		return types.RefreshToken{}, errors.New("All arguments must be not nil")
	}

	refreshToken, err := insertReturning[types.RefreshToken](ctx, `
        INSERT INTO "refreshTokens" ("id", "userID", "expiresAt")
        VALUES ($1, $2, $3) 
        RETURNING *
    `, *id, *userID, *expiresAt)

	return refreshToken, err
}

func DeleteRefreshToken(ctx context.Context, userID *string) error {
	return simpleDelete(ctx, userID, "refreshTokens", "userID")
}

func GetRefreshToken(ctx context.Context, id *string) (types.RefreshToken, error) {
	if id == nil {
		return types.RefreshToken{}, errors.New("All arguments must be not nil")
	}

	refreshToken, err := selectOneReturning[types.RefreshToken](ctx, `SELECT * FROM "refreshTokens" WHERE "id" = $1`, *id)

	return refreshToken, err
}

func UpdateUser(ctx context.Context, user *types.User, updateUser *types.UpdateUser, passwordHash *string) (types.User, error) {
	if (updateUser == nil) == (passwordHash == nil) {
		return types.User{}, errors.New("Eather updateUser or passwordHash should not be nil")
	}

	var row pgx.Rows
	var err error

	if updateUser != nil {
		query := `UPDATE "users" SET "username"=$1, "firstName"=$2, "lastName"=$3, "dateOfBirth"=$4 WHERE "id"=$5 RETURNING *`
		row, err = pool.Query(ctx, query,
			utils.UpdateSelector(updateUser.Username, &user.Username),
			utils.UpdateSelector(updateUser.FirstName, &user.FirstName),
			utils.UpdateSelector(updateUser.LastName, &user.LastName),
			utils.UpdateSelector(updateUser.DateOfBirth, &user.DateOfBirth),
			user.ID,
		)
	} else {
		query := `UPDATE "users" SET "passwordHash"=$1 WHERE "id"=$2 RETURNING *`
		row, err = pool.Query(ctx, query, passwordHash, user.ID)
	}

	if err != nil {
		return types.User{}, errors.Wrap(err, "error executing query")
	}
	defer row.Close()

	busser, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[types.User])

	if errors.Is(err, pgx.ErrNoRows) {
		return types.User{}, err
	} else if err != nil {
		return types.User{}, errors.Wrap(err, "error on collecting row")
	}

	return busser, nil
}
