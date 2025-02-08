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
			return types.Project{}, errors.WithStack(err)
		}

		if len(project.AdminIDs) > 0 {
			adminRows := make([][]interface{}, len(project.AdminIDs))

			for i, userID := range project.AdminIDs {
				adminRows[i] = []interface{}{
					newProject.ID, userID,
				}
			}

			if err = bulkInsert(ctx, tx, "projectAdmins", []string{"projectID", "userID"}, adminRows); err != nil {
				return types.Project{}, errors.WithStack(err)
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
				return types.Project{}, errors.WithStack(err)
			}

			newProject.MembersIDs = project.MembersIDs
		}

		return newProject, nil
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
			return types.Project{}, errors.WithStack(err)
		}

		project.AdminIDs = utils.Map(func(i int, admin types.ProjectAdmin) string {
			return admin.UserID
		}, admins)

		members, err := queryReturning[types.ProjectMember](ctx, `SELECT * FROM "projectMembers" WHERE "projectID" = $1`, id)

		if err != nil {
			return types.Project{}, errors.WithStack(err)
		}

		project.MembersIDs = utils.Map(func(i int, member types.ProjectMember) string {
			return member.UserID
		}, members)

		return project, nil
	})
}

func UpdateProject(ctx context.Context, id *string, updateProject *types.UpdateProject) (types.Project, error) {
	if id == nil || updateProject == nil {
		return types.Project{}, errors.New("All arguments must not be nil")
	}

	return withTx(ctx, func(tx pgx.Tx) (types.Project, error) {
		row := tx.QueryRow(ctx, `
            UPDATE "projects"
            SET "name"=COALESCE($1, "name"), "description"=COALESCE($2, "description")
            WHERE "id"=$3
            RETURNING *
            `, updateProject.Name, updateProject.Description, id)

		var project types.Project
		err := row.Scan(&project.ID, &project.CreatorID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return types.Project{}, errors.WithStack(err)
		}

		if updateProject.AdminIDs != nil {
			_, err = queryOneReturningTx[any](ctx, tx, `DELETE FROM "projectAdmins" WHERE "projectID"=$1`, id)

			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return types.Project{}, errors.WithStack(err)
			}

			if len(*updateProject.AdminIDs) > 0 {
				adminRows := make([][]interface{}, len(*updateProject.AdminIDs))

				for i, userID := range *updateProject.AdminIDs {
					adminRows[i] = []interface{}{
						project.ID, userID,
					}
				}

				if err = bulkInsert(ctx, tx, "projectAdmins", []string{"projectID", "userID"}, adminRows); err != nil {
					return types.Project{}, errors.WithStack(err)
				}

				project.AdminIDs = *updateProject.AdminIDs
			}
		}

		if updateProject.MembersIDs != nil {
			_, err = queryOneReturningTx[any](ctx, tx, `DELETE FROM "projectMembers" WHERE "projectID"=$1`, id)

			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return types.Project{}, errors.WithStack(err)
			}

			if len(*updateProject.MembersIDs) > 0 {
				memberRows := make([][]interface{}, len(*updateProject.MembersIDs))

				for i, userID := range *updateProject.MembersIDs {
					memberRows[i] = []interface{}{
						project.ID, userID,
					}
				}

				if err = bulkInsert(ctx, tx, "projectMembers", []string{"projectID", "userID"}, memberRows); err != nil {
					return types.Project{}, errors.WithStack(err)
				}

				project.MembersIDs = *updateProject.MembersIDs
			}
		}

		return project, nil
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

	return errors.WithStack(err)
}

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
			return types.KanbanColumn{}, errors.WithStack(err)
		}

		if labelID != nil {
			label, err := queryOneReturning[types.KanbanColumnLabel](ctx, `SELECT * FROM kanbanColumnLabels WHERE id = $1`, labelID)

			if err != nil {
				return types.KanbanColumn{}, errors.WithStack(err)
			}

			newColumn.Label = &label
		}

		return newColumn, errors.WithStack(err)
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
			return types.KanbanColumn{}, errors.WithStack(err)
		} else if err != nil {
			return types.KanbanColumn{}, errors.Wrap(err, "error getting column")
		}

		if labelID != nil {
			label, err := queryOneReturning[types.KanbanColumnLabel](ctx, `SELECT * FROM kanbanColumnLabels WHERE id = $1`, labelID)

			if err != nil {
				return types.KanbanColumn{}, errors.WithStack(err)
			}

			column.Label = &label
		}

		return column, nil
	})
}

func UpdateKanbanColumn(ctx context.Context, updateColumn *types.UpdateKanbanColumn, column *types.KanbanColumn) (types.KanbanColumn, error) {
	if column == nil || updateColumn == nil {
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

		row := tx.QueryRow(ctx, `UPDATE "kanbanColumns" SET "name"=COALESCE($1, "name"), "order"=COALESCE($2, "order"), "labelID"=$3 WHERE "id"=$4 RETURNING *`,
			updateColumn.Name,
			updateColumn.Order,
			newLabelID,
			column.ID)

		var updatedColumn types.KanbanColumn
		var labelID *string
		err := row.Scan(&updatedColumn.ID, &updatedColumn.ProjectID, &updatedColumn.Name, &updatedColumn.Order, &labelID)
		if errors.Is(err, pgx.ErrNoRows) {
			return types.KanbanColumn{}, errors.WithStack(err)
		} else if err != nil {
			return types.KanbanColumn{}, errors.Wrap(err, "error getting column")
		}

		if labelID != nil {
			label, err := queryOneReturning[types.KanbanColumnLabel](ctx, `SELECT * FROM kanbanColumnLabels WHERE id = $1`, labelID)

			if err != nil {
				return types.KanbanColumn{}, errors.WithStack(err)
			}

			updatedColumn.Label = &label
		}

		return updatedColumn, nil
	})
}

func DeleteKanbanColumn(ctx context.Context, id *string) error {
	if id == nil {
		return errors.New("Id must not be nil")
	}

	_, err := queryOneReturning[any](ctx, `DELETE FROM "kanbanColumns" WHERE id = $1`, id)

	return errors.WithStack(err)
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
			return nil, errors.WithStack(err)
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
				return nil, errors.WithStack(err)
			}

			admins, err := queryReturning[types.ProjectAdmin](ctx,
				`SELECT * FROM "projectAdmins" WHERE "projectID" = $1`, project.ID)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			project.AdminIDs = utils.Map(func(i int, admin types.ProjectAdmin) string {
				return admin.UserID
			}, admins)

			members, err := queryReturning[types.ProjectMember](ctx,
				`SELECT * FROM "projectMembers" WHERE "projectID" = $1`, project.ID)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			project.MembersIDs = utils.Map(func(i int, member types.ProjectMember) string {
				return member.UserID
			}, members)

			projects = append(projects, project)
		}

		return projects, errors.WithStack(rows.Err())
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
			return nil, errors.WithStack(err)
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
				return nil, errors.WithStack(err)
			}

			if labelID != nil {
				label, err := queryOneReturningTx[types.KanbanColumnLabel](ctx, tx, `SELECT * FROM "kanbanColumnLabels" WHERE "id" = $1`, labelID)
				if err != nil {
					return nil, errors.WithStack(err)
				}
				column.Label = &label
			}

			columns = append(columns, column)
		}

		return columns, errors.WithStack(rows.Err())
	})
}

func ShiftColumnOrder(ctx context.Context, projectID *string, fromOrder int) error {
	if projectID == nil {
		return errors.New("projectID must not be nil")
	}

	_, err := queryOneReturning[any](ctx, `
        UPDATE "kanbanColumns"
        SET "order" = "order" + 1
        WHERE "projectID" = $1 AND "order" >= $2 
    `, projectID, fromOrder)

	return errors.WithStack(err)
}

func ShiftColumnOrdersInRange(ctx context.Context, projectID *string, startOrder, endOrder, shift int) error {
	if projectID == nil {
		return errors.New("projectID must not be nil")
	}

	_, err := queryOneReturning[any](ctx, `
        UPDATE "kanbanColumns" 
        SET "order" = "order" + $1
        WHERE "projectID" = $2 
        AND "order" >= $3 
        AND "order" <= $4
    `, shift, projectID, startOrder, endOrder)

	return errors.WithStack(err)
}

func CreateKanbanColumnLabel(ctx context.Context, id *string, createColumnLabel *types.CreateKanbanColumnLabel, specialTag *types.SpecialTag) (types.KanbanColumnLabel, error) {
	if id == nil || createColumnLabel == nil {
		return types.KanbanColumnLabel{}, errors.New("createColumnLabel and id must not be nil")
	}

	return queryOneReturning[types.KanbanColumnLabel](ctx, `
        INSERT INTO "kanbanColumnLabels" 
        ("id", "name", "projectID", "color", "specialTag") 
        VALUES ($1, $2, $3, $4, $5)
        RETURNING *
    `, id, createColumnLabel.Name, createColumnLabel.ProjectID, createColumnLabel.Color, specialTag)
}

func GetKanbanColumnLabel(ctx context.Context, id *string) (types.KanbanColumnLabel, error) {
	if id == nil {
		return types.KanbanColumnLabel{}, errors.New("id is nil")
	}

	return queryOneReturning[types.KanbanColumnLabel](ctx, `SELECT * FROM "kanbanColumnLabels" WHERE "id"=$1`, id)
}

func DeleteKanbanColumnLabel(ctx context.Context, id *string) error {
	if id == nil {
		return errors.New("createColumnLabel must not be nil")
	}

	_, err := queryOneReturning[types.KanbanColumnLabel](ctx, `
        DELETE FROM "kanbanColumnLabels" WHERE "id" = $1
    `, id)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return errors.WithStack(err)
	}

	return nil
}

func UpdateKanbanColumnLabel(ctx context.Context, id *string, updateColumnLabel *types.UpdateKanbanColumnLabel) (types.KanbanColumnLabel, error) {
	if id == nil || updateColumnLabel == nil {
		return types.KanbanColumnLabel{}, errors.New("createColumnLabel and id  must not be nil")
	}

	return queryOneReturning[types.KanbanColumnLabel](ctx, `
        UPDATE "kanbanColumnLabels" 
        SET "name" = coalesce($1, "name"), "color" = coalesce($2, "color")
        WHERE "id"=$3
        RETURNING *
    `, updateColumnLabel.Name, updateColumnLabel.Color, id)
}

func GetKanbanColumnLabels(ctx context.Context, projectID *string) ([]types.KanbanColumnLabel, error) {
	if projectID == nil {
		return nil, errors.New("projectID must not be nil")
	}

	return queryReturning[types.KanbanColumnLabel](ctx, `
        SELECT * FROM "kanbanColumnLabels" WHERE "projectID" = $1
    `, projectID)
}

func CreateKanbanRow(ctx context.Context, id *string, userID *string, createRow *types.CreateKanbanRow) (types.KanbanRow, error) {
	if id == nil || userID == nil || createRow == nil {
		return types.KanbanRow{}, errors.New("id, userID and createRow must not be nil")
	}

	return withTx(ctx, func(tx pgx.Tx) (types.KanbanRow, error) {
		row := tx.QueryRow(ctx, `
            INSERT INTO "kanbanRows" 
            ("id", "columnID", "name", "description", "order", "creatorID", "priority", "labelID", "dueDate") 
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
            RETURNING *
        `, id, createRow.ColumnID, createRow.Name, createRow.Description, createRow.Order, userID, createRow.Priority, createRow.LabelID, createRow.DueDate)

		var kanbanRow types.KanbanRow
		var labelID *string

		if err := row.Scan(&kanbanRow.ID,
			&kanbanRow.ColumnID, &kanbanRow.Name,
			&kanbanRow.Description, &kanbanRow.Order,
			&kanbanRow.CreatorID, &kanbanRow.Priority,
			&labelID, &kanbanRow.CreatedAt,
			&kanbanRow.UpdatedAt, &kanbanRow.DueDate); err != nil {
			return types.KanbanRow{}, errors.WithStack(err)
		}

		if len(createRow.AssignedUsersIDs) > 0 {
			assignee := make([][]interface{}, len(createRow.AssignedUsersIDs))

			for i, userID := range createRow.AssignedUsersIDs {
				assignee[i] = []interface{}{id, userID}
			}

			if err := bulkInsert(ctx, tx, "kanbanRowAssignees", []string{"rowID", "userID"}, assignee); err != nil {
				return types.KanbanRow{}, errors.WithStack(err)
			}

			kanbanRow.AssignedUsersIDs = createRow.AssignedUsersIDs
		}

		if labelID != nil {
			rowLabel, err := queryOneReturningTx[types.KanbanRowLabel](ctx, tx, `SELECT * FROM "kanbanRowLabels" WHERE "id" = $1`, labelID)

			if err != nil {
				return types.KanbanRow{}, errors.WithStack(err)
			}

			kanbanRow.Label = &rowLabel
		}

		return kanbanRow, nil
	})
}

func UpdateKanbanRow(ctx context.Context, updateRow *types.UpdateKanbanRow, row *types.KanbanRow) (types.KanbanRow, error) {
	if row == nil || updateRow == nil {
		return types.KanbanRow{}, errors.New("row and updateRow must not be nil")
	}

	var newLabelID *string
	if updateRow.LabelID != nil {
		newLabelID = updateRow.LabelID
	} else if updateRow.DeleteLabel != nil {
		newLabelID = nil
	} else if row.Label != nil {
		newLabelID = &row.Label.ID
	}

	return withTx(ctx, func(tx pgx.Tx) (types.KanbanRow, error) {
		qRow := tx.QueryRow(ctx,
			`UPDATE "kanbanRows"
                SET "name" = coalesce($1, "name"),
                    "description" = coalesce($2, "description"),
                    "order" = coalesce($3, "order"),
                    "priority" = coalesce($4, "priority"),
                    "labelID" = coalesce($5, "labelID"),
                    "dueDate" = coalesce($6, "dueDate")
                WHERE "id" = $7
                RETURNING *
            `, updateRow.Name, updateRow.Description, updateRow.Order, updateRow.Priority, newLabelID, updateRow.DueDate, row.ID)

		var updatedRow types.KanbanRow
		var labelID *string
		err := qRow.Scan(&updatedRow.ID, &updatedRow.ColumnID,
			&updatedRow.Name, &updatedRow.Description,
			&updatedRow.Order, &updatedRow.CreatorID, &updatedRow.Priority,
			&labelID, &updatedRow.CreatedAt,
			&updatedRow.UpdatedAt, &updatedRow.DueDate,
		)

		if err != nil {
			return types.KanbanRow{}, errors.WithStack(err)
		}

		if updateRow.AssignedUsersIDs != nil {
			_, err = tx.Exec(ctx, `DELETE FROM "kanbanRowAssignees" WHERE "rowID" = $1`, row.ID)
			if err != nil {
				return types.KanbanRow{}, errors.WithStack(err)
			}

			if len(*updateRow.AssignedUsersIDs) > 0 {
				for _, userID := range *updateRow.AssignedUsersIDs {
					_, err = tx.Exec(ctx,
						`INSERT INTO "kanbanRowAssignees" ("rowID", "userID") VALUES ($1, $2)`,
						row.ID, userID)
					if err != nil {
						return types.KanbanRow{}, errors.WithStack(err)
					}
				}
				updatedRow.AssignedUsersIDs = *updateRow.AssignedUsersIDs
			}
		}

		if labelID != nil {
			label, err := queryOneReturningTx[types.KanbanRowLabel](ctx, tx,
				`SELECT * FROM "kanbanRowLabels" WHERE "id" = $1`, labelID)
			if err != nil {
				return types.KanbanRow{}, errors.WithStack(err)
			}
			updatedRow.Label = &label
		}

		return updatedRow, nil
	})
}

func GetRows(ctx context.Context, columnID *string) ([]types.KanbanRow, error) {
	if columnID == nil {
		return []types.KanbanRow{}, errors.New("columnID must not be nil")
	}

	return withTx(ctx, func(tx pgx.Tx) ([]types.KanbanRow, error) {
		qRows, err := tx.Query(ctx, `
            SELECT "id", "columnID", "name", "description", "order", 
                   "creatorID", "priority", "labelID", "createdAt", 
                   "updatedAt", "dueDate"
            FROM "kanbanRows"
            WHERE "columnID"=$1 
            ORDER BY "order" ASC
        `, columnID)
		if err != nil {
			return []types.KanbanRow{}, errors.WithStack(err)
		}
		defer qRows.Close()

		var rows []types.KanbanRow
		for qRows.Next() {
			var row types.KanbanRow
			var labelID *string

			err = qRows.Scan(&row.ID, &row.ColumnID, &row.Name, &row.Description, &row.Order, &row.CreatorID, &row.Priority, &labelID, &row.CreatedAt, &row.UpdatedAt, &row.DueDate)
			if err != nil {
				return []types.KanbanRow{}, errors.WithStack(err)
			}

			if labelID != nil {
				label, err := queryOneReturningTx[types.KanbanRowLabel](ctx, tx, `SELECT * FROM kanbanRowLabels WHERE id = $1`, labelID)
				if err != nil {
					return []types.KanbanRow{}, errors.WithStack(err)
				}

				row.Label = &label
			}

			assignees, err := queryReturning[types.KanbanRowAssignedUser](ctx,
				`SELECT * FROM "kanbanRowAssignees" WHERE "rowID"=$1`, row.ID)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			row.AssignedUsersIDs = utils.Map(func(_ int, user types.KanbanRowAssignedUser) string {
				return user.UserID
			}, assignees)

			rows = append(rows, row)
		}

		if err = qRows.Err(); err != nil {
			return []types.KanbanRow{}, errors.WithStack(err)
		}

		return rows, nil
	})
}

func DeleteRow(ctx context.Context, rowID *string) error {
	if rowID == nil {
		return errors.New("rowID must not be nil")
	}

	_, err := queryOneReturning[any](ctx, `
        DELETE FROM "kanbanRows" WHERE "id"=$1
        `, rowID)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return errors.WithStack(err)
	}

	return nil
}

func GetRow(ctx context.Context, rowID *string) (types.KanbanRow, error) {
	if rowID == nil {
		return types.KanbanRow{}, errors.New("rowID must not be nil")
	}

	return withTx(ctx, func(tx pgx.Tx) (types.KanbanRow, error) {
		qRow := tx.QueryRow(ctx, `SELECT * FROM "kanbanRows" WHERE id = $1`, rowID)

		var row types.KanbanRow
		var labelID *string

		err := qRow.Scan(&row.ID, &row.ColumnID, &row.Name, &row.Description, &row.Order, &row.CreatorID, &row.Priority, &labelID, &row.CreatedAt, &row.UpdatedAt, &row.DueDate)
		if err != nil {
			return types.KanbanRow{}, errors.WithStack(err)
		}

		if labelID != nil {
			label, err := queryOneReturning[types.KanbanRowLabel](ctx, `SELECT * FROM kanbanRowLabels WHERE id = $1`, labelID)
			if err != nil {
				return types.KanbanRow{}, errors.WithStack(err)
			}

			row.Label = &label
		}

		assignees, err := queryReturningTx[types.KanbanRowAssignedUser](ctx, tx,
			`SELECT * FROM "kanbanRowAssignees" WHERE "rowID"=$1`, row.ID)
		if err != nil {
			return types.KanbanRow{}, errors.WithStack(err)
		}
		row.AssignedUsersIDs = utils.Map(func(_ int, user types.KanbanRowAssignedUser) string {
			return user.UserID
		}, assignees)

		return row, nil
	})
}

func ShiftRowOrder(ctx context.Context, columnID *string, fromOrder int) error {
	if columnID == nil {
		return errors.New("columnID must not be nil")
	}

	_, err := queryOneReturning[any](ctx, `
        UPDATE "kanbanRows"
        SET "order" = "order" + 1
        WHERE "columnID" = $1 AND "order" >= $2 
    `, columnID, fromOrder)

	return errors.WithStack(err)
}
