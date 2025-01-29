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
		row := tx.QueryRow(ctx, `SELECT * FROM "projects" WHERE id = $1`, id)

		var project types.Project
		err := row.Scan(&project.ID, &project.CreatorID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return types.Project{}, err
		} else if err != nil {
			return types.Project{}, errors.Wrap(err, "error getting project")
		}

		admins, err := queryReturning[types.ProjectAdmin](ctx, `SELECT * FROM "projectAdmins" WHERE "projectID" = $1`, id)

		if err != nil {
			return types.Project{}, err
		}

		project.AdminIDs = utils.Map(func(i int, admin types.ProjectAdmin) string {
			return admin.UserID
		}, admins)

		members, err := queryReturning[types.ProjectMember](ctx, `SELECT * FROM "projectMembers" WHERE "projectID" = $1`, id)

		if err != nil {
			return types.Project{}, err
		}

		project.MembersIDs = utils.Map(func(i int, member types.ProjectMember) string {
			return member.UserID
		}, members)

		return project, err
	})
}

func UpdateProject(ctx context.Context, id *string, project *types.Project, updateProject *types.UpdateProject) (types.Project, error) {
	if id == nil || updateProject == nil {
		return types.Project{}, errors.New("All arguments must not be nil")
	}

	return withTx(ctx, func(tx pgx.Tx) (types.Project, error) {
		row := tx.QueryRow(ctx, `
            UPDATE "projects"
            SET "name"=$1, "description"=$2
            WHERE "id"=$3
            RETURNING *
            `, utils.UpdateSelector(updateProject.Name, &project.Name), utils.UpdateSelector(updateProject.Description, &project.Description), id)

		var project types.Project
		err := row.Scan(&project.ID, &project.CreatorID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return types.Project{}, err
		}

		if updateProject.AdminIDs != nil {
			_, err = queryOneReturningTx[any](ctx, tx, `DELETE FROM "projectAdmins" WHERE "projectID"=$1`, id)

			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return types.Project{}, err
			}

			if len(*updateProject.AdminIDs) > 0 {
				adminRows := make([][]interface{}, len(*updateProject.AdminIDs))

				for i, userID := range *updateProject.AdminIDs {
					adminRows[i] = []interface{}{
						project.ID, userID,
					}
				}

				if err = bulkInsert(ctx, tx, "projectAdmins", []string{"projectID", "userID"}, adminRows); err != nil {
					return types.Project{}, err
				}

				project.AdminIDs = *updateProject.AdminIDs
			}
		}

		if updateProject.MembersIDs != nil {
			_, err = queryOneReturningTx[any](ctx, tx, `DELETE FROM "projectMembers" WHERE "projectID"=$1`, id)

			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return types.Project{}, err
			}

			if len(*updateProject.MembersIDs) > 0 {
				memberRows := make([][]interface{}, len(*updateProject.MembersIDs))

				for i, userID := range *updateProject.MembersIDs {
					memberRows[i] = []interface{}{
						project.ID, userID,
					}
				}

				if err = bulkInsert(ctx, tx, "projectMembers", []string{"projectID", "userID"}, memberRows); err != nil {
					return types.Project{}, err
				}

				project.MembersIDs = *updateProject.MembersIDs
			}
		}

		return project, err
	})
}

func DeleteProject(ctx context.Context, userID *string, projectID *string) error {
	if userID == nil || projectID == nil {
		return errors.New("Id must not be nil")
	}

	_, err := queryReturning[any](ctx, `
    		DELETE FROM "projectMembers" WHERE "projectID" = $1;
    		DELETE FROM "projectAdmins" WHERE "projectID" = $1;
    		DELETE FROM "projects" WHERE "id" = $1;
		`, projectID)

	return err
}

// TODO: reorder other columns on creation
func CreateKanbanColumn(ctx context.Context, id *string, projectID *string, createColumn *types.CreateKanbanColumn) (types.KanbanColumn, error) {
	if projectID == nil || createColumn == nil {
		return types.KanbanColumn{}, errors.New("All arguments must not be nil")
	}

	return withTx(ctx, func(tx pgx.Tx) (types.KanbanColumn, error) {
		row := tx.QueryRow(ctx, `
        	INSERT INTO "kanbanColumns" ("id", "projectID", "name", "order", "labelID")
        	VALUES ($1, $2, $3, $4, $5)
        	RETURNING *
    	`, id, createColumn.ProjectID, createColumn.Name, createColumn.Order, createColumn.LabelID)

		var newColumn types.KanbanColumn
		var labelID *string
		err := row.Scan(&newColumn.ID, &newColumn.ProjectID, &newColumn.Name, &newColumn.Order, &labelID)
		if err != nil {
			return types.KanbanColumn{}, err
		}

		if labelID != nil {
			label, err := queryOneReturning[types.KanbanColumnLabel](ctx, `SELECT * FROM kanbanColumnLabels WHERE id = $1`, labelID)

			if err != nil {
				return types.KanbanColumn{}, err
			}

			newColumn.Label = &label
		}

		return newColumn, err
	})
}

func GetKanbanColumn(ctx context.Context, id *string) (types.KanbanColumn, error) {
	if id == nil {
		return types.KanbanColumn{}, errors.New("Id must not be nil")
	}

	return withTx(ctx, func(tx pgx.Tx) (types.KanbanColumn, error) {
		row := tx.QueryRow(ctx, `SELECT * FROM "kanbanColumns" WHERE id = $1`, id)

		var column types.KanbanColumn
		var labelID *string
		err := row.Scan(&column.ID, &column.ProjectID, &column.Name, &column.Order, &labelID)
		if errors.Is(err, pgx.ErrNoRows) {
			return types.KanbanColumn{}, err
		} else if err != nil {
			return types.KanbanColumn{}, errors.Wrap(err, "error getting column")
		}

		if labelID != nil {
			label, err := queryOneReturning[types.KanbanColumnLabel](ctx, `SELECT * FROM kanbanColumnLabels WHERE id = $1`, labelID)

			if err != nil {
				return types.KanbanColumn{}, err
			}

			column.Label = &label
		}

		return column, nil
	})
}

// TODO: reorder other columns on update
func UpdateKanbanColumn(ctx context.Context, id *string, updateColumn *types.UpdateKanbanColumn, column types.KanbanColumn) (types.KanbanColumn, error) {
	if id == nil || updateColumn == nil {
		return types.KanbanColumn{}, errors.New("all arguments must not be nil")
	}

	return withTx(ctx, func(tx pgx.Tx) (types.KanbanColumn, error) {
		var newLabelID *string
		if updateColumn.LabelID != nil {
			newLabelID = updateColumn.LabelID
		} else if updateColumn.DeleteLabel != nil {
			newLabelID = nil
		} else if column.Label != nil {
			newLabelID = &column.Label.ID
		}

		row := tx.QueryRow(ctx, `UPDATE "kanbanColumns" SET "name"=$1, "order"=$2, "labelID"=$3 WHERE "id"=$4 RETURNING *`,
			utils.UpdateSelector(updateColumn.Name, &column.Name),
			utils.UpdateSelector(updateColumn.Order, &column.Order),
			newLabelID,
			id)

		var updatedColumn types.KanbanColumn
		var labelID *string
		err := row.Scan(&updatedColumn.ID, &updatedColumn.ProjectID, &updatedColumn.Name, &updatedColumn.Order, &labelID)
		if errors.Is(err, pgx.ErrNoRows) {
			return types.KanbanColumn{}, err
		} else if err != nil {
			return types.KanbanColumn{}, errors.Wrap(err, "error getting column")
		}

		if labelID != nil {
			label, err := queryOneReturning[types.KanbanColumnLabel](ctx, `SELECT * FROM kanbanColumnLabels WHERE id = $1`, labelID)

			if err != nil {
				return types.KanbanColumn{}, err
			}

			updatedColumn.Label = &label
		}

		return updatedColumn, nil
	})
}

// TODO: reorder other columns on delete
func DeleteKanbanColumn(ctx context.Context, id *string) error {
	if id == nil {
		return errors.New("Id must not be nil")
	}

	_, err := queryOneReturning[any](ctx, `DELETE FROM "kanbanColumns" WHERE id = $1`, id)

	return err
}

func GetProjectsByUserID(ctx context.Context, userID *string) ([]types.Project, error) {
	if userID == nil {
		return []types.Project{}, errors.New("UserId must not be nil")
	}

	return withTx(ctx, func(tx pgx.Tx) ([]types.Project, error) {
		rows, err := tx.Query(ctx, `
			SELECT DISTINCT p.*
			FROM "projects" p
			WHERE p."creatorID" = $1
			    OR EXISTS (
			        SELECT 1 FROM "projectAdmins" pa 
			        WHERE pa."projectID" = p."id" AND pa."userID" = $1
			    )
			    OR EXISTS (
			        SELECT 1 FROM "projectMembers" pm 
			        WHERE pm."projectID" = p."id" AND pm."userID" = $1
			    )
			ORDER BY p."createdAt" DESC;
		`, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var projects []types.Project
		for rows.Next() {
			var project types.Project
			err := rows.Scan(
				&project.ID,
				&project.CreatorID,
				&project.Name,
				&project.Description,
				&project.CreatedAt,
				&project.UpdatedAt,
			)
			if err != nil {
				return nil, err
			}

			admins, err := queryReturning[types.ProjectAdmin](ctx,
				`SELECT * FROM "projectAdmins" WHERE "projectID" = $1`, project.ID)
			if err != nil {
				return nil, err
			}
			project.AdminIDs = utils.Map(func(i int, admin types.ProjectAdmin) string {
				return admin.UserID
			}, admins)

			members, err := queryReturning[types.ProjectMember](ctx,
				`SELECT * FROM "projectMembers" WHERE "projectID" = $1`, project.ID)
			if err != nil {
				return nil, err
			}
			project.MembersIDs = utils.Map(func(i int, member types.ProjectMember) string {
				return member.UserID
			}, members)

			projects = append(projects, project)
		}

		return projects, rows.Err()
	})
}

func GetColumns(ctx context.Context, projectID *string) ([]types.KanbanColumn, error) {
	if projectID == nil {
		return []types.KanbanColumn{}, errors.New("projectID must not be nil")
	}

	return withTx(ctx, func(tx pgx.Tx) ([]types.KanbanColumn, error) {
		rows, err := tx.Query(ctx, `
                SELECT * FROM "kanbanColumns" WHERE "projectID"=$1 ORDER BY "order" ASC
            `, projectID)

		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var columns []types.KanbanColumn
		for rows.Next() {
			var column types.KanbanColumn
			var labelID *string
			err := rows.Scan(
				&column.ID,
				&column.ProjectID,
				&column.Name,
				&column.Order,
				&labelID,
			)
			if err != nil {
				return nil, err
			}

			if labelID != nil {
				label, err := queryOneReturningTx[types.KanbanColumnLabel](ctx, tx, `SELECT * FROM "kanbanColumnLabels" WHERE "id" = $1`, labelID)
				if err != nil {
					return nil, err
				}
				column.Label = &label
			}

			columns = append(columns, column)
		}

		return columns, nil
	})
}
