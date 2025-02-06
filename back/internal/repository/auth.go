package repository

import (
	"context"
	"time"

	"github.com/finkabaj/squid/back/internal/types"
	"github.com/pkg/errors"
)

func CreateUser(ctx context.Context, id *string, passwordHash *string, user *types.RegisterUser) (types.User, error) {
	if id == nil || passwordHash == nil || user == nil {
		return types.User{}, errors.New("All arguments must be not nil")
	}

	return queryOneReturning[types.User](ctx, `
        INSERT INTO "users" ("id", "username", "firstName", "lastName", "dateOfBirth", "email", "passwordHash")
        VALUES ($1, $2, $3, $4, $5, $6, $7) 
        RETURNING *
    `, id, user.Username, user.FirstName, user.LastName, user.DateOfBirth, user.Email, passwordHash)
}

func GetUser(ctx context.Context, id *string, email *string) (types.User, error) {
	if id == nil && email == nil {
		return types.User{}, errors.New("At least one arguments must not be nil")
	}

	if id != nil {
		return queryOneReturning[types.User](ctx, `SELECT * FROM "users" WHERE "id" = $1`, id)
	} else {
		return queryOneReturning[types.User](ctx, `SELECT * FROM "users" WHERE "email" = $1`, email)
	}
}

func DeleteUser(ctx context.Context, id *string) (err error) {
	_, err = queryOneReturning[types.User](ctx, `DELETE FROM "users" WHERE "id" = $1`, id)

	return
}

func CreateRefreshToken(ctx context.Context, id *string, userID *string, expiresAt *time.Time) (types.RefreshToken, error) {
	if id == nil || userID == nil || expiresAt == nil {
		return types.RefreshToken{}, errors.New("All arguments must be not nil")
	}

	return queryOneReturning[types.RefreshToken](ctx, `
        INSERT INTO "refreshTokens" ("id", "userID", "expiresAt")
        VALUES ($1, $2, $3) 
        RETURNING *
    `, id, userID, expiresAt)
}

func DeleteRefreshToken(ctx context.Context, userID *string) (err error) {
	_, err = queryOneReturning[types.RefreshToken](ctx, `DELETE FROM "refreshTokens" WHERE "userID" = $1`, userID)
	return
}

func GetRefreshToken(ctx context.Context, id *string) (types.RefreshToken, error) {
	if id == nil {
		return types.RefreshToken{}, errors.New("All arguments must be not nil")
	}

	return queryOneReturning[types.RefreshToken](ctx, `SELECT * FROM "refreshTokens" WHERE "id" = $1`, id)
}

func UpdateUser(ctx context.Context, user *types.User, updateUser *types.UpdateUser, passwordHash *string) (types.User, error) {
	if (updateUser == nil) == (passwordHash == nil) {
		return types.User{}, errors.New("Either updateUser or passwordHash should not be nil")
	}

	if updateUser != nil {
		return queryOneReturning[types.User](ctx, `UPDATE "users" 
            SET "username"=COALESCE($1, "username"),
            "firstName"=COALESCE($2, "firstName"),
            "lastName"=COALESCE($3, "lastName"),
            "dateOfBirth"=COALESCE($4, "dateOfBirth")
            WHERE "id"=$5
            RETURNING *`,
			updateUser.Username,
			updateUser.FirstName,
			updateUser.LastName,
			updateUser.DateOfBirth,
			user.ID,
		)
	} else {
		return queryOneReturning[types.User](ctx, `UPDATE "users" SET "passwordHash"=$1 WHERE "id"=$2 RETURNING *`, passwordHash, user.ID)
	}
}
