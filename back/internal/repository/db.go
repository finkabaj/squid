package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/finkabaj/squid/back/internal/types"
	"github.com/finkabaj/squid/back/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

var pool *pgxpool.Pool

func simpleDelete(ctx context.Context, id *string, tableName string, fieldName string) error {
	if id == nil {
		return errors.New("All arguments must be not nil")
	}

	query := fmt.Sprintf(`DELETE FROM "%s" WHERE "%s" = $1`, tableName, fieldName)

	result, err := pool.Exec(ctx, query, *id)
	if err != nil {
		return errors.Wrap(err, "error executing delete")
	}

	if result.RowsAffected() == 0 {
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
		    "userID" VARCHAR(255) NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
		    "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			"expiresAt" TIMESTAMP NOT NULL
		);
        CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON "refreshTokens"("userID");
        CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON "refreshTokens"("expiresAt");
		`); err != nil {
		return errors.Wrap(err, "error creating refreshTokens table")
	}

	if _, err = transaction.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "projects" (
		    "id" VARCHAR(255) PRIMARY KEY,
		    "creatorID" VARCHAR(255) NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
		    "name" VARCHAR(50) NOT NULL,
		    "description" VARCHAR(500),
		    "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    "updatedAt" TIMESTAMP NOT NULL
		);

		CREATE TABLE "projectAdmins" (
    		"projectID" VARCHAR(255) REFERENCES "projects"("id") ON DELETE CASCADE,
    		"userID" VARCHAR(255) REFERENCES "users"("id") ON DELETE CASCADE,
    		PRIMARY KEY ("projectID", "userID")
		);

		CREATE TABLE "projectMembers" (
		    "projectID" VARCHAR(255) REFERENCES "projects"("id") ON DELETE CASCADE,
		    "userID" VARCHAR(255) REFERENCES "users"("id") ON DELETE CASCADE,
		    PRIMARY KEY ("projectID", "userID")
		);

		CREATE INDEX idx_project_admins_user ON "projectAdmins"("userID");
		CREATE INDEX idx_project_members_user ON "projectMembers"("userID");

		DROP TRIGGER IF EXISTS update_projects_updated_at on "public"."projects";
        CREATE TRIGGER update_projects_updated_at
            BEFORE UPDATE ON "projects"
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();

	`); err != nil {
		return errors.Wrap(err, "error creating project table")
	}

	if _, err = transaction.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "kanbanColumns" (
		    "id" VARCHAR(255) PRIMARY KEY,
		    "projectID" VARCHAR(255) NOT NULL REFERENCES "projects"("id") ON DELETE CASCADE,
		    "name" VARCHAR(50) NOT NULL,
		    "order" INTEGER NOT NULL,
		    "labelID" VARCHAR(255) REFERENCES "kanbanColumnLabels"("id")
		);

		CREATE INDEX idx_kanban_columns_project ON "kanbanColumns"("projectID");
	`); err != nil {
		return errors.Wrap(err, "error creating kanbanColumns table")
	}

	if _, err = transaction.Exec(ctx, `
		CREATE TYPE IF NOT EXISTS "specialTags" AS ENUM ('TODO', 'IN_PROGRESS', 'TESTING', 'COMPLETED');

		CREATE TABLE IF NOT EXISTS "kanbanColumnLabels" (
		    "id" VARCHAR(255) PRIMARY KEY,
		    "projectID" VARCHAR(255) NOT NULL REFERENCES "projects"("id") ON DELETE CASCADE,
		    "specialTag" "specialTags",
		    "name" VARCHAR(50) NOT NULL,
		    "color" VARCHAR(255) NOT NULL
		);

		CREATE INDEX idx_kanban_column_labels_project ON "kanbanColumnLabels"("projectID");
	`); err != nil {
		return errors.Wrap(err, "error creating kanbanColumnLabels table")
	}

	if _, err = transaction.Exec(ctx, `
		CREATE TYPE IF NOT EXISTS "priorities" AS ENUM ('LOW', 'MEDIUM', 'HIGH');

		CREATE TABLE IF NOT EXISTS "kanbanRows" (
		    "id" VARCHAR(255) PRIMARY KEY,
			"columnID" VARCHAR(255) NOT NULL REFERENCES "kanbanColumns"("id") ON DELETE CASCADE,
		    "name" VARCHAR(50) NOT NULL,
			"description" VARCHAR(500),
		    "order" INTEGER NOT NULL,
		    "creatorID" VARCHAR(255) NOT NULL REFERENCES "users"("id"),
		    "priority" "priorities",
			"labelID" VARCHAR(255) REFERENCES "kanbanRowLabels"("id"),
			"createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			"updatedAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			"dueDate" TIMESTAMP
		);

		CREATE INDEX idx_kanban_rows_column ON "kanbanRows"("columnID");

		CREATE TABLE "kanbanRowAssignees" (
		    "rowID" VARCHAR(255) REFERENCES "kanbanRows"("id") ON DELETE CASCADE,
		    "userID" VARCHAR(255) REFERENCES "users"("id") ON DELETE CASCADE,
		    PRIMARY KEY ("rowID", "userID")
		);

		CREATE INDEX idx_kanban_row_assignees_user ON "kanbanRowAssignees"("userID");

		DROP TRIGGER IF EXISTS update_kanban_rows_updated_at on "public"."kanbanRows";
        CREATE TRIGGER update_kanban_rows_updated_at
            BEFORE UPDATE ON "kanbanRows"
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
	`); err != nil {
		return errors.Wrap(err, "error creating kanbanRows table")
	}

	if _, err = transaction.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "kanbanRowLabels" (
		    "id" VARCHAR(255) PRIMARY KEY,
			"projectID" VARCHAR(255) NOT NULL REFERENCES "projects"("id") ON DELETE CASCADE,
		    "name" VARCHAR(50) NOT NULL,
			"color" VARCHAR(255) NOT NULL
		);

		CREATE INDEX idx_kanban_row_labels_project ON "kanbanRowLabels"("projectID");
	`); err != nil {
		return errors.Wrap(err, "error creating kanbanRowLabels table")
	}

	if _, err = transaction.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "historyPoints" (
		    "id" VARCHAR(255) PRIMARY KEY,
			"rowID" VARCHAR(255) NOT NULL REFERENCES "kanbanRows"("id") ON DELETE CASCADE,
		    "userID" VARCHAR(255) NOT NULL REFERENCES "users"("id"),
		    "text" VARCHAR(255) NOT NULL,
		    "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX idx_history_points_row ON "historyPoints"("rowID");
	`); err != nil {
		return errors.Wrap(err, "error creating historyPoints table")
	}

	if _, err = transaction.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "checkLists" (
		    "id" VARCHAR(255) PRIMARY KEY,
		    "rowID" VARCHAR(255) NOT NULL REFERENCES "kanbanRows"("id") ON DELETE CASCADE
		);

		CREATE INDEX idx_points_checklist ON "points"("checkListID");
	`); err != nil {
		return errors.Wrap(err, "error creating checkLists table")
	}

	if _, err = transaction.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "points" (
			"id" VARCHAR(255) PRIMARY KEY,
		    "checkListID" VARCHAR(255) NOT NULL REFERENCES "checkLists"("id") ON DELETE CASCADE,
		    "name" VARCHAR(50) NOT NULL,
		    "description" VARCHAR(255) NOT NULL,
		    "completed" BOOLEAN NOT NULL DEFAULT FALSE,
		    "completedAt" TIMESTAMP,
		    "completedBy" VARCHAR(255) REFERENCES "users"("id")
		);
	`); err != nil {
		return errors.Wrap(err, "error creating points table")
	}

	if _, err = transaction.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "commentSections" (
		    "id" VARCHAR(255) PRIMARY KEY,
			"rowID" VARCHAR(255) NOT NULL REFERENCES "kanbanRows"("id") ON DELETE CASCADE,
		    "canComment" BOOLEAN NOT NULL DEFAULT TRUE
		);
	`); err != nil {
		return errors.Wrap(err, "error creating commentSections table")
	}

	if _, err = transaction.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "comments" (
			"id" VARCHAR(255) PRIMARY KEY,
		    "commentSectionID" VARCHAR(255) NOT NULL REFERENCES "commentSections"(id) ON DELETE CASCADE,
		    "userID" VARCHAR(255) NOT NULL REFERENCES "users"("id"),
		    "text" VARCHAR(255) NOT NULL,
		    "createdAt" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX idx_comments_user ON "comments"("userID");
	`); err != nil {
		return errors.Wrap(err, "error creating comments table")
	}

	return
}

func Connect(credentials types.DBCredentials) (err error) {
	config, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		credentials.User,
		credentials.Password,
		credentials.Host,
		credentials.Port,
		credentials.Database))

	if err != nil {
		return errors.Wrap(err, "error parsing config")
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 24 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute
	config.ConnConfig.ConnectTimeout = 5 * time.Second

	config.ConnConfig.RuntimeParams = map[string]string{
		"application_name": "squid",
		"search_path":      "public",
		"timezone":         "UTC",
	}

	pool, err = pgxpool.NewWithConfig(context.Background(), config)

	if err != nil {
		return errors.Wrap(err, "error on db setup")
	}

	if err = Status(); err != nil {
		return errors.Wrap(err, "error on checking db status")
	}

	return setup()
}

func Close() error {
	if pool != nil {
		pool.Close()
	}
	return nil
}

func Status() (err error) {
	err = pool.Ping(context.Background())
	return
}

func CreateUser(ctx context.Context, id *string, passwordHash *string, user *types.RegisterUser) (types.User, error) {
	if id == nil || passwordHash == nil || user == nil {
		return types.User{}, errors.New("All arguments must be not nil")
	}

	query := `
        INSERT INTO "users" ("id", "username", "firstName", "lastName", "dateOfBirth", "email", "passwordHash")
        VALUES ($1, $2, $3, $4, $5, $6, $7) 
        RETURNING *
    `

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

func GetUser(ctx context.Context, id *string, email *string) (types.User, error) {
	if id == nil && email == nil {
		return types.User{}, errors.New("At least one arguments must not be nil")
	}

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

	if errors.Is(err, pgx.ErrNoRows) {
		return types.User{}, err
	} else if err != nil {
		return types.User{}, errors.Wrap(err, "error on collecting row")
	}

	return user, nil
}

func DeleteUser(ctx context.Context, id *string) error {
	return simpleDelete(ctx, id, "users", "id")
}

func CreateRefreshToken(ctx context.Context, id *string, userID *string, expiresAt *time.Time) (types.RefreshToken, error) {
	if id == nil || userID == nil || expiresAt == nil {
		return types.RefreshToken{}, errors.New("All arguments must be not nil")
	}

	query := `
        INSERT INTO "refreshTokens" ("id", "userID", "expiresAt")
        VALUES ($1, $2, $3) 
        RETURNING *
    `

	row, err := pool.Query(ctx, query, *id, *userID, *expiresAt)
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

func DeleteRefreshToken(ctx context.Context, userID *string) error {
	return simpleDelete(ctx, userID, "refreshTokens", "userID")
}

func GetRefreshToken(ctx context.Context, id *string) (types.RefreshToken, error) {
	if id == nil {
		return types.RefreshToken{}, errors.New("All arguments must be not nil")
	}

	query := `SELECT * FROM "refreshTokens" WHERE "id" = $1`

	row, err := pool.Query(ctx, query, *id)
	if err != nil {
		return types.RefreshToken{}, errors.Wrap(err, "error executing query")
	}
	defer row.Close()

	refreshToken, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[types.RefreshToken])

	if errors.Is(err, pgx.ErrNoRows) {
		return types.RefreshToken{}, err
	} else if err != nil {
		return types.RefreshToken{}, errors.Wrap(err, "error on collecting row")
	}

	return refreshToken, nil
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
