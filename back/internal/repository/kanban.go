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
		newProject, err := insertReturningTx[types.Project](ctx, tx, `
        INSERT INTO "projects" ("id", "creatorID", "name", "description")
        VALUES ($1, $2, $3, $4)
        RETURNING *
    `, *id, *creatorID, project.Name, project.Description)

		if err != nil {
			return types.Project{}, err
		}

		adminRows := make([][]interface{}, len(project.AdminIDs))

		for _, userID := range project.AdminIDs {
			adminRows = append(adminRows, []interface{}{
				newProject.ID, userID,
			})
		}

		if err = bulkInsert(ctx, tx, "projectAdmins", []string{"projectID", "userID"}, adminRows); err != nil {
			return types.Project{}, err
		}

		memberRows := make([][]interface{}, len(project.MembersIDs))

		for _, userID := range project.MembersIDs {
			memberRows = append(memberRows, []interface{}{
				newProject.ID, userID,
			})
		}

		if err = bulkInsert(ctx, tx, "projectMembers", []string{"projectID", "userID"}, memberRows); err != nil {
			return types.Project{}, err
		}

		return newProject, err
	})
}

func GetProject(ctx context.Context, id *string) (types.Project, error) {
	if id == nil {
		return types.Project{}, errors.New("Id must not be nil")
	}

	return withTx(ctx, func(tx pgx.Tx) (types.Project, error) {
		project, err := selectOneReturning[types.Project](ctx, `SELECT * FROM "projects" WHERE id = $1`, *id)

		if err != nil {
			return types.Project{}, err
		}

		admins, err := selectReturning[types.ProjectAdmin](ctx, `SELECT * FROM "projectAdmins" WHERE projectID = $1`, *id)

		if err != nil {
			return types.Project{}, err
		}

		project.AdminIDs = utils.Map(func(i int, admin types.ProjectAdmin) string {
			return admin.UserID
		}, admins)

		members, err := selectReturning[types.ProjectMember](ctx, `SELECT * FROM "projectMembers" WHERE projectID = $1`, *id)

		if err != nil {
			return types.Project{}, err
		}

		project.MembersIDs = utils.Map(func(i int, member types.ProjectMember) string {
			return member.UserID
		}, members)

		return project, err
	})
}
