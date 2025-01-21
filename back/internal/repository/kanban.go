package repository

import (
	"context"

	"github.com/finkabaj/squid/back/internal/types"
	"github.com/finkabaj/squid/back/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

func CreateProject(ctx context.Context, id *string, creatorID *string, project *types.CreateProject) (types.Project, error) {
	if id == nil || project == nil {
		return types.Project{}, errors.New("All arguments must not be nil")
	}

	return withTx(ctx, func(tx pgx.Tx) (types.Project, error) {
		row := tx.QueryRow(ctx, `
        	INSERT INTO "projects" ("id", "creatorID", "name", "description")
        	VALUES ($1, $2, $3, $4)
        	RETURNING *
    	`, id, creatorID, project.Name, project.Description)

		var newProject types.Project
		err := row.Scan(&newProject.ID, &newProject.CreatorID, &newProject.Name, &newProject.Description, &newProject.CreatedAt, &newProject.UpdatedAt)
		if err != nil {
			return types.Project{}, err
		}

		if len(project.AdminIDs) > 0 {
			adminRows := make([][]interface{}, len(project.AdminIDs))

			for i, userID := range project.AdminIDs {
				adminRows[i] = []interface{}{
					newProject.ID, userID,
				}
			}

			if err = bulkInsert(ctx, tx, "projectAdmins", []string{"projectID", "userID"}, adminRows); err != nil {
				return types.Project{}, err
			}

			newProject.AdminIDs = project.AdminIDs
		}

		if len(project.MembersIDs) > 0 {
			memberRows := make([][]interface{}, len(project.MembersIDs))

			for i, userID := range project.MembersIDs {
				memberRows[i] = []interface{}{
					newProject.ID, userID,
				}
			}

			if err = bulkInsert(ctx, tx, "projectMembers", []string{"projectID", "userID"}, memberRows); err != nil {
				return types.Project{}, err
			}

			newProject.MembersIDs = project.MembersIDs
		}

		return newProject, err
	})
}

func GetProject(ctx context.Context, id *string) (types.Project, error) {
	if id == nil {
		return types.Project{}, errors.New("Id must not be nil")
	}

	return withTx(ctx, func(tx pgx.Tx) (types.Project, error) {
		project, err := queryOneReturning[types.Project](ctx, `SELECT * FROM "projects" WHERE id = $1`, id)

		if err != nil {
			return types.Project{}, err
		}

		admins, err := queryReturning[types.ProjectAdmin](ctx, `SELECT * FROM "projectAdmins" WHERE projectID = $1`, id)

		if err != nil {
			return types.Project{}, err
		}

		project.AdminIDs = utils.Map(func(i int, admin types.ProjectAdmin) string {
			return admin.UserID
		}, admins)

		members, err := queryReturning[types.ProjectMember](ctx, `SELECT * FROM "projectMembers" WHERE projectID = $1`, id)

		if err != nil {
			return types.Project{}, err
		}

		project.MembersIDs = utils.Map(func(i int, member types.ProjectMember) string {
			return member.UserID
		}, members)

		return project, err
	})
}
