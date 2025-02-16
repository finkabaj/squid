package repository

import (
	"context"
	"fmt"

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

	_, err := withTx(ctx, func(tx pgx.Tx) (any, error) {
		var columnOrder int
		var projectID string
		err := tx.QueryRow(ctx, `
            SELECT "order", "projectID" FROM "kanbanColumns" 
            WHERE id = $1
        `, id).Scan(&columnOrder, &projectID)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		_, err = tx.Exec(ctx, `
            DELETE FROM "kanbanColumns" WHERE id = $1
        `, id)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(err)
		}

		_, err = tx.Exec(ctx, `
            UPDATE "kanbanColumns" 
            SET "order" = "order" - 1
            WHERE "projectID" = $1 AND "order" > $2
        `, projectID, columnOrder)

		return nil, errors.WithStack(err)
	})

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

func ShiftOrder(ctx context.Context, tableName string, idName string, id *string, fromOrder int) error {
	if id == nil {
		return errors.New("id must not be nil")
	}

	query := fmt.Sprintf(`
        UPDATE "%s"
        SET "order" = "order" + 1
        WHERE "%s" = $1 AND "order" >= $2
    `, tableName, idName)

	_, err := queryOneReturning[any](ctx, query, id, fromOrder)

	return errors.WithStack(err)
}

func ShiftOrdersInRange(ctx context.Context, tableName string, idName string, id *string, startOrder, endOrder, shift int) error {
	if id == nil {
		return errors.New("id must not be nil")
	}

	query := fmt.Sprintf(`
        UPDATE "%s"
        SET "order" = "order" + $1
        WHERE "%s" = $2
        AND "order" >= $3
        AND "order" <= $4 
    `, tableName, idName)

	_, err := queryOneReturning[any](ctx, query, shift, id, startOrder, endOrder)

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

func CreateKanbanRow(ctx context.Context, id *string, commentSectionID *string, userID *string, createRow *types.CreateKanbanRow) (types.KanbanRow, error) {
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

		row = tx.QueryRow(ctx, `
            INSERT INTO "commentSections"
            ("id", "rowID", "canComment")
            VALUES ($1, $2, $3)
            RETURNING *
        `, commentSectionID, id, true)

		var commentSection types.CommentSection
		if err := row.Scan(&commentSection.ID, &commentSection.RowID, &commentSection.CanComment); err != nil {
			return types.KanbanRow{}, errors.WithStack(err)
		}

		kanbanRow.Comments = &commentSection

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
			rowLabel, err := queryOneReturning[types.KanbanRowLabel](ctx, `SELECT * FROM "kanbanRowLabels" WHERE "id" = $1`, labelID)

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
				label, err := queryOneReturning[types.KanbanRowLabel](ctx, `SELECT * FROM "kanbanRowLabels" WHERE id = $1`, labelID)
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

			historyPoints, err := queryReturning[types.HistoryPoint](ctx, `SELECT * FROM "historyPoints" WHERE "rowID" = $1`, row.ID)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return []types.KanbanRow{}, errors.WithStack(err)
			}

			row.History = &historyPoints

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

	_, err := withTx(ctx, func(tx pgx.Tx) (any, error) {
		var rowOrder int
		var columnID string
		err := tx.QueryRow(ctx, `
            SELECT "order", "columnID" FROM "kanbanRows" 
            WHERE id = $1
        `, rowID).Scan(&rowOrder, &columnID)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		_, err = tx.Exec(ctx, `
            DELETE FROM "kanbanRows" WHERE id = $1
        `, rowID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(err)
		}

		_, err = tx.Exec(ctx, `
            UPDATE "kanbanRows" 
            SET "order" = "order" - 1
            WHERE "columnID" = $1 AND "order" > $2
        `, columnID, rowOrder)

		return nil, errors.WithStack(err)
	})

	return errors.WithStack(err)
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
			label, err := queryOneReturning[types.KanbanRowLabel](ctx, `SELECT * FROM "kanbanRowLabels" WHERE id = $1`, labelID)
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

		historyPoints, err := queryReturning[types.HistoryPoint](ctx, `SELECT * FROM "historyPoints" WHERE "rowID" = $1`, row.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return types.KanbanRow{}, errors.WithStack(err)
		}

		row.History = &historyPoints

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

func CreateKanbanRowLabel(ctx context.Context, id *string, createRowLabel *types.CreateKanbanRowLabel) (types.KanbanRowLabel, error) {
	if id == nil || createRowLabel == nil {
		return types.KanbanRowLabel{}, errors.New("createRowLabel and id must not be nil")
	}

	return queryOneReturning[types.KanbanRowLabel](ctx, `
        INSERT INTO "kanbanRowLabels" 
        ("projectID", "id", "name", "color") 
        VALUES ($1, $2, $3, $4)
        RETURNING *
    `, createRowLabel.ProjectID, id, createRowLabel.Name, createRowLabel.Color)
}

func GetKanbanRowLabel(ctx context.Context, id *string) (types.KanbanRowLabel, error) {
	if id == nil {
		return types.KanbanRowLabel{}, errors.New("id is nil")
	}

	return queryOneReturning[types.KanbanRowLabel](ctx, `SELECT * FROM "kanbanRowLabels" WHERE "id"=$1`, id)
}

func DeleteKanbanRowLabel(ctx context.Context, id *string) error {
	if id == nil {
		return errors.New("id must not be nil")
	}

	_, err := queryOneReturning[any](ctx, `
        DELETE FROM "kanbanRowLabels" WHERE "id" = $1
    `, id)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return errors.WithStack(err)
	}

	return nil
}

func UpdateKanbanRowLabel(ctx context.Context, id *string, updateRowLabel *types.UpdateKanbanRowLabel) (types.KanbanRowLabel, error) {
	if id == nil || updateRowLabel == nil {
		return types.KanbanRowLabel{}, errors.New("updateRowLabel and id  must not be nil")
	}

	return queryOneReturning[types.KanbanRowLabel](ctx, `
        UPDATE "kanbanRowLabels" 
        SET "name" = coalesce($1, "name"), "color" = coalesce($2, "color")
        WHERE "id"=$3
        RETURNING *
    `, updateRowLabel.Name, updateRowLabel.Color, id)
}

func GetKanbanRowLabels(ctx context.Context, projectID *string) ([]types.KanbanRowLabel, error) {
	if projectID == nil {
		return nil, errors.New("projectID must not be nil")
	}

	return queryReturning[types.KanbanRowLabel](ctx, `
        SELECT * FROM "kanbanRowLabels" WHERE "projectID" = $1
    `, projectID)
}

func CreateHistoryPoint(ctx context.Context, createHistoryPoint *types.HistoryPoint) error {
	if createHistoryPoint == nil {
		return errors.New("createHistoryPoint must not be nil")
	}

	_, err := queryOneReturning[types.HistoryPoint](ctx, `
        INSERT INTO "historyPoints" 
        ("id", "rowID", "userID", "text") 
        VALUES ($1, $2, $3, $4)
        RETURNING *
    `, createHistoryPoint.ID, createHistoryPoint.RowID, createHistoryPoint.UserID, createHistoryPoint.Text)

	return err
}

func GetHistoryPoints(ctx context.Context, rowID *string) ([]types.HistoryPoint, error) {
	if rowID == nil {
		return nil, errors.New("rowID must not be nil")
	}

	return queryReturning[types.HistoryPoint](ctx, `SELECT * FROM "historyPoints" WHERE "rowID" = $1`, rowID)
}

func CreateChecklist(ctx context.Context, createChecklist *types.Checklist) (types.Checklist, error) {
	if createChecklist == nil {
		return types.Checklist{}, errors.New("createChecklist must not be nil")
	}

	if _, err := pool.Exec(ctx, `
        INSERT INTO "checklists" 
        ("id", "rowID")
        VALUES ($1, $2)
    `, createChecklist.ID, createChecklist.RowID); err != nil {
		return types.Checklist{}, errors.WithStack(err)
	}

	return *createChecklist, nil
}

func GetChecklist(ctx context.Context, checklistID *string) (types.Checklist, error) {
	if checklistID == nil {
		return types.Checklist{}, errors.New("checklistID must not be nil")
	}

	row := pool.QueryRow(ctx, `SELECT * FROM "checklists" WHERE "id" = $1`, checklistID)

	var checklist types.Checklist
	err := row.Scan(&checklist.ID, &checklist.RowID)
	if err != nil {
		return types.Checklist{}, errors.WithStack(err)
	}

	points, err := queryReturning[types.Point](ctx, `SELECT * FROM "points" WHERE "checklistID" = $1 ORDER BY "completed" ASC`, checklistID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return types.Checklist{}, errors.WithStack(err)
	}

	checklist.Points = &points

	return checklist, nil
}

func DeleteChecklist(ctx context.Context, id *string) error {
	if id == nil {
		return errors.New("id must not be nil")
	}

	_, err := queryOneReturning[any](ctx, `
        DELETE FROM "checklists" WHERE "id" = $1
    `, id)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return errors.WithStack(err)
	}

	return nil
}

func ChecklistExists(ctx context.Context, rowID *string) (bool, error) {
	if rowID == nil {
		return false, errors.New("rowID must not be nil")
	}

	row := pool.QueryRow(ctx, `
        SELECT 1 FROM "checklists" WHERE "rowID" = $1;
    `, rowID)

	var check any
	err := row.Scan(&check)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return false, errors.WithStack(err)
	}

	return check != nil, nil
}

func CreatePoint(ctx context.Context, id *string, createPoint *types.CreatePoint) (types.Point, error) {
	if createPoint == nil || id == nil {
		return types.Point{}, errors.New("createPoint and id must not be nil")
	}

	point, err := queryOneReturning[types.Point](ctx, `
        INSERT INTO "points"
        ("id", "checklistID", "name", "description", "completed")
        VALUES ($1, $2, $3, $4, false)
        RETURNING *
    `, id, createPoint.ChecklistID, createPoint.Name, createPoint.Description)

	return point, errors.WithStack(err)
}

func UpdatePoint(ctx context.Context, id *string, updatePoint *types.UpdatePoint, updateStatus bool) (types.Point, error) {
	if updatePoint == nil || id == nil {
		return types.Point{}, errors.New("updatePoint and id must not be nil")
	}

	if updateStatus {
		point, err := queryOneReturning[types.Point](ctx, `
            UPDATE "points"
            SET "completed" = $1, "completedAt" = $2, "completedBy" = $3
            WHERE "id" = $4
            RETURNING *
        `, updatePoint.Completed, updatePoint.CompletedAt, updatePoint.CompletedBy, id)

		return point, errors.WithStack(err)
	} else {
		point, err := queryOneReturning[types.Point](ctx, `
        UPDATE "points"
        SET name=coalesce($1, name),
            description=coalesce($2, description)
        WHERE "id"=$3
        RETURNING *
    `, updatePoint.Name, updatePoint.Description, id)

		return point, errors.WithStack(err)
	}
}

func DeletePoint(ctx context.Context, id *string) error {
	if id == nil {
		return errors.New("id must not be nil")
	}

	_, err := queryOneReturning[any](ctx, `
        DELETE FROM "points" WHERE "id" = $1
    `, id)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return errors.WithStack(err)
	}

	return nil
}

func GetPoints(ctx context.Context, checklistID *string) ([]types.Point, error) {
	if checklistID == nil {
		return []types.Point{}, errors.New("checklistID must not be nil")
	}

	points, err := queryReturning[types.Point](ctx, `SELECT * FROM "points" WHERE "checklistID" = $1 ORDER BY "completed" ASC`, checklistID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return []types.Point{}, errors.WithStack(err)
	}

	return points, nil
}

func GetPoint(ctx context.Context, id *string) (types.Point, error) {
	if id == nil {
		return types.Point{}, errors.New("id must not be nil")
	}

	point, err := queryOneReturning[types.Point](ctx, `
        SELECT * FROM "points" WHERE "id" = $1
    `, id)

	return point, errors.WithStack(err)
}

func ChangeCanComment(ctx context.Context, id *string) (types.CommentSection, error) {
	if id == nil {
		return types.CommentSection{}, errors.New("id must not be nil")
	}

	row := pool.QueryRow(ctx, `
        UPDATE "commentSections"
        SET "canComment" = NOT "canComment"
        WHERE "id" = $1
        RETURNING *
    `, id)

	var commentSection types.CommentSection
	if err := row.Scan(&commentSection.ID, &commentSection.RowID, &commentSection.CanComment); err != nil {
		return types.CommentSection{}, errors.WithStack(err)
	}

	return commentSection, nil
}

func CreateComment(ctx context.Context, userID, id *string, createComment *types.CreateComment) (types.Comment, error) {
	if createComment == nil || id == nil {
		return types.Comment{}, errors.New("createComment and id must not be nil")
	}

	comment, err := queryOneReturning[types.Comment](ctx, `
        INSERT INTO "comments"
        ("id", "commentSectionID", "userID", "text")
        VALUES ($1, $2, $3, $4)
        RETURNING *
    `, id, createComment.CommentSectionID, userID, createComment.Text)

	return comment, errors.WithStack(err)
}

func DeleteComment(ctx context.Context, id *string) error {
	if id == nil {
		return errors.New("id must not be nil")
	}

	_, err := queryOneReturning[any](ctx, `
        DELETE FROM "comments" WHERE "id" = $1
    `, id)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return errors.WithStack(err)
	}

	return nil
}

func GetComments(ctx context.Context, commentSectionID *string) ([]types.Comment, error) {
	if commentSectionID == nil {
		return []types.Comment{}, errors.New("commentSectionID must not be nil")
	}

	comments, err := queryReturning[types.Comment](ctx, `SELECT * FROM "comments" WHERE "commentSectionID" = $1`, commentSectionID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return []types.Comment{}, errors.WithStack(err)
	}

	return comments, nil
}

func GetCommentSection(ctx context.Context, id *string) (types.CommentSection, error) {
	if id == nil {
		return types.CommentSection{}, errors.New("id must not be nil")
	}

	row := pool.QueryRow(ctx, `
        SELECT * FROM "commentSections" WHERE "id" = $1
    `, id)

	var CommentSection types.CommentSection
	if err := row.Scan(&CommentSection.ID, &CommentSection.RowID, &CommentSection.CanComment); err != nil {
		return types.CommentSection{}, errors.WithStack(err)
	}

	return CommentSection, nil
}

func GetComment(ctx context.Context, id *string) (types.Comment, error) {
	if id == nil {
		return types.Comment{}, errors.New("id must not be nil")
	}

	comment, err := queryOneReturning[types.Comment](ctx, `
        SELECT * FROM "comments" WHERE "id" = $1
    `, id)

	return comment, errors.WithStack(err)
}
