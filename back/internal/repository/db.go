package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/finkabaj/squid/back/internal/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

var pool *pgxpool.Pool

func simpleDelete(id *string, tableName string, fieldName string) error {
	if id == nil {
		return errors.New("All arguments must be not nil")
	}

	query := fmt.Sprintf(`DELETE FROM "%s" WHERE "%s" = $1`, tableName, fieldName)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	row, err := pool.Query(ctx, query, *id)
	if err != nil {
		return errors.Wrap(err, "error executing query")
	}
	defer row.Close()

	if !row.Next() {
		return errors.New("no rows were deleted")
	}

	return nil
}

func setup() (err error) {
	ctx := context.Background()

	transaction, err := pool.Begin(ctx)

	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			transaction.Rollback(ctx)
		} else {
			transaction.Commit(ctx)
		}
	}()

	if _, err = transaction.Exec(ctx, `
		CREATE OR REPLACE FUNCTION update_updated_at_column()   
		RETURNS TRIGGER AS $$
		BEGIN
		    NEW."updatedAt" = CURRENT_TIMESTAMP;
		    RETURN NEW;   
		END;
		$$ language 'plpgsql';
		
		CREATE OR REPLACE FUNCTION set_created_at_column()   
		RETURNS TRIGGER AS $$
		BEGIN
		    NEW."createdAt" = CURRENT_TIMESTAMP;
		    NEW."updatedAt" = CURRENT_TIMESTAMP;
		    RETURN NEW;   
		END;
		$$ language 'plpgsql';	
		`); err != nil {
		return errors.Wrap(err, "Error creating update_updated_at_column and set_created_at_column functions")
	}

	if _, err = transaction.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS "users" (
            "id" VARCHAR(255) PRIMARY KEY,
            "username" VARCHAR(255) NOT NULL,
			"firstName" VARCHAR(255) NOT NULL,	
            "lastName" VARCHAR(255) NOT NULL,
            "dateOfBirth" DATE NOT NULL,
            "email" VARCHAR(255) NOT NULL UNIQUE,
            "passwordHash" VARCHAR(255) NOT NULL,
            "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            "updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );
        CREATE INDEX IF NOT EXISTS idx_users_email ON "users"("email");
    
		DROP TRIGGER IF EXISTS set_created_at_column on "public"."users";
		CREATE TRIGGER set_users_created_at
            BEFORE INSERT ON "users"
            FOR EACH ROW
            EXECUTE FUNCTION set_created_at_column();

		DROP TRIGGER IF EXISTS update_users_updated_at on "public"."users";
        CREATE TRIGGER update_users_updated_at
            BEFORE UPDATE ON "users"
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
        `); err != nil {
		return errors.Wrap(err, "error creating users table")
	}

	if _, err = transaction.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "refreshTokens" (
		    "id" VARCHAR(255) PRIMARY KEY,
		    "userId" VARCHAR(255) NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
		    "tokenHash" VARCHAR(255) NOT NULL,
		    "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			"expiresAt" TIMESTAMP NOT NULL
		);
        CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON "refreshTokens"("userId");
        CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON "refreshTokens"("expiresAt");

		DROP TRIGGER IF EXISTS set_refresh_tokens_created_at on "public"."refreshTokens";
		CREATE TRIGGER set_refresh_tokens_created_at
            BEFORE INSERT ON "refreshTokens"
            FOR EACH ROW
            EXECUTE FUNCTION set_created_at_column();
		`); err != nil {
		return errors.Wrap(err, "error creating refreshTokens table")
	}

	return
}

func Connect(credentials types.DBCredentials) (err error) {
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", credentials.User, credentials.Password, credentials.Host, credentials.Port, credentials.Database)
	pool, err = pgxpool.New(context.Background(), connStr)

	if err != nil {
		return errors.Wrap(err, "error on db setup")
	}

	if err = Status(); err != nil {
		return errors.Wrap(err, "error on checking db status")
	}

	err = setup()

	return
}

func Close() {
	pool.Close()
}

func Status() (err error) {
	err = pool.Ping(context.Background())
	return
}

func CreateUser(id *string, passwordHash *string, user *types.RegisterUser) (types.User, error) {
	if id == nil || passwordHash == nil || user == nil {
		return types.User{}, errors.New("All arguments must be not nil")
	}

	query := `
        INSERT INTO "users" ("id", "username", "firstName", "lastName", "dateOfBirth", "email", "passwordHash")
        VALUES ($1, $2, $3, $4, $5, $6, $7) 
        RETURNING *
    `

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	row, err := pool.Query(ctx, query, *id, user.Username, user.FirstName, user.LastName, user.DateOfBirth, user.Email, passwordHash)
	if err != nil {
		return types.User{}, errors.Wrap(err, "error executing query")
	}
	defer row.Close()

	newUser, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[types.User])

	if err != nil {
		return types.User{}, errors.Wrap(err, "error on collecting row")
	}

	return newUser, nil
}

func GetUser(id *string, email *string) (types.User, error) {
	if id == nil && email == nil {
		return types.User{}, errors.New("At least one arguments must not be nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	var row pgx.Rows
	var err error

	if id != nil {
		query := `SELECT * FROM "users" WHERE "id" = $1`
		row, err = pool.Query(ctx, query, *id)
	} else {
		query := `SELECT * FROM "users" WHERE "email" = $1`
		row, err = pool.Query(ctx, query, *email)
	}

	if err != nil {
		return types.User{}, errors.Wrap(err, "error executing query")
	}
	defer row.Close()

	user, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[types.User])

	if err == pgx.ErrNoRows {
		return types.User{}, errors.New("user not found")
	} else if err != nil {
		return types.User{}, errors.Wrap(err, "error on collecting row")
	}

	return user, nil
}

func DeleteUser(id *string) error {
	return simpleDelete(id, "users", "id")
}

func CreateRefreshToken(id *string, userID *string, tokenHash *string, expiresAt *time.Time) (types.RefreshToken, error) {
	if id == nil || userID == nil || tokenHash == nil || expiresAt == nil {
		return types.RefreshToken{}, errors.New("All arguments must be not nil")
	}

	query := `
        INSERT INTO "refreshTokens" ("id", "userId", "tokenHash", "expiresAt")
        VALUES ($1, $2, $3, $4) 
        RETURNING *
    `

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	row, err := pool.Query(ctx, query, *id, *userID, *tokenHash, *expiresAt)
	if err != nil {
		return types.RefreshToken{}, errors.Wrap(err, "error executing query")
	}
	defer row.Close()

	refreshToken, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[types.RefreshToken])

	if err != nil {
		return types.RefreshToken{}, errors.Wrap(err, "error on collecting row")
	}

	return refreshToken, nil
}

func DeleteRefreshToken(id *string) error {
	return simpleDelete(id, "refreshTokens", "id")
}

func GetRefreshToken(id *string) (types.RefreshToken, error) {
	if id == nil {
		return types.RefreshToken{}, errors.New("All arguments must be not nil")
	}

	query := `SELECT * FROM "refreshTokens" WHERE "id" = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	row, err := pool.Query(ctx, query, *id)
	if err != nil {
		return types.RefreshToken{}, errors.Wrap(err, "error executing query")
	}
	defer row.Close()

	refreshToken, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[types.RefreshToken])

	if err == pgx.ErrNoRows {
		return types.RefreshToken{}, errors.New("refresh token not found")
	} else if err != nil {
		return types.RefreshToken{}, errors.Wrap(err, "error on collecting row")
	}

	return refreshToken, nil
}
