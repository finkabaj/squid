package service

import (
	"context"
	"fmt"
	"time"

	"github.com/finkabaj/squid/back/internal/repository"
	"github.com/finkabaj/squid/back/internal/types"
	"github.com/finkabaj/squid/back/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

func CreateProject(user *types.User, project *types.CreateProject) (types.Project, error) {
	if user == nil || project == nil {
		return types.Project{}, utils.NewBadRequestError(errors.New("user or project is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	for i, adminEmail := range project.AdminEmails {
		if adminEmail == user.Email || utils.Have(func(_ int, memberEmail string) bool {
			return memberEmail == adminEmail
		}, project.MemberEmails) {
			return types.Project{}, utils.NewBadRequestError(errors.New("user email should be unique for each category"))
		}

		if utils.Have(func(j int, aEmail string) bool { return i != j && aEmail == adminEmail }, project.AdminEmails) {
			return types.Project{}, utils.NewBadRequestError(errors.New("user email should be unique for each category"))
		}
		admin, err := repository.GetUser(ctx, nil, &adminEmail)
		project.AdminIDs = append(project.AdminIDs, admin.ID)
		if errors.Is(err, pgx.ErrNoRows) {
			return types.Project{}, utils.NewBadRequestError(errors.New(fmt.Sprintf("no user with email: %s found", adminEmail)))
		} else if err != nil {
			return types.Project{}, utils.NewInternalError(err)
		}
	}

	for i, memberEmail := range project.MemberEmails {
		if memberEmail == user.Email {
			return types.Project{}, utils.NewBadRequestError(errors.New("user email should be unique for each category"))
		}

		if utils.Have(func(j int, mEmail string) bool { return i != j && mEmail == memberEmail }, project.MemberEmails) {
			return types.Project{}, utils.NewBadRequestError(errors.New("user email should be unique for each category"))
		}

		member, err := repository.GetUser(ctx, nil, &memberEmail)
		project.MemberIDs = append(project.MemberIDs, member.ID)
		if errors.Is(err, pgx.ErrNoRows) {
			return types.Project{}, utils.NewBadRequestError(errors.New(fmt.Sprintf("no user with email: %s found", memberEmail)))
		} else if err != nil {
			return types.Project{}, utils.NewInternalError(err)
		}
	}

	id := uuid.New().String()

	newProject, err := repository.CreateProject(ctx, &id, &user.ID, project)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	return newProject, nil
}

func GetProject(userID *string, projectID *string) (types.Project, error) {
	if userID == nil || projectID == nil {
		return types.Project{}, utils.NewBadRequestError(errors.New("userID or projectID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project, err := repository.GetProject(ctx, projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("project with id: %s not found", *projectID)))
	} else if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}
	columns, err := repository.GetColumns(ctx, projectID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}
	for i, column := range columns {
		rows, err := repository.GetRows(ctx, &column.ID)
		if err != nil {
			return types.Project{}, utils.NewInternalError(err)
		}
		for j, row := range rows {
			checklist, err := repository.GetChecklistByRowID(ctx, &row.ID)
			if errors.Is(err, pgx.ErrNoRows) {
				row.Checklist = nil
			} else if err != nil {
				return types.Project{}, utils.NewInternalError(err)
			} else {
				row.Checklist = &checklist
			}
			commentSection, err := repository.GetCommentSectionByRowID(ctx, &row.ID)
			if err != nil {
				return types.Project{}, utils.NewInternalError(err)
			}
			row.CommentSection = &commentSection
			comments, err := repository.GetComments(ctx, &commentSection.ID)
			if err != nil {
				return types.Project{}, utils.NewInternalError(err)
			}
			row.CommentSection.Comments = &comments
			rows[j] = row
		}
		columns[i].Rows = &rows
	}

	project.Columns = &columns

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) &&
		!utils.Have(func(i int, memberID string) bool { return memberID == *userID }, project.MembersIDs) {
		return types.Project{}, utils.NewUnauthorizedError(errors.New("you cannot fetch a project in which you are not participating"))
	}

	return project, nil
}

func GetProjects(userID *string) ([]types.Project, error) {
	if userID == nil {
		return []types.Project{}, utils.NewBadRequestError(errors.New("userID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projects, err := repository.GetProjectsByUserID(ctx, userID)

	if err != nil {
		return []types.Project{}, utils.NewInternalError(err)
	}

	return projects, nil
}

func UpdateProject(id *string, user *types.User, updateProject *types.UpdateProject) (types.Project, error) {
	if id == nil || user == nil || updateProject == nil {
		return types.Project{}, utils.NewBadRequestError(errors.New("all parameters must not be nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if updateProject.Name == nil && updateProject.AdminIDs == nil && updateProject.MembersIDs == nil && updateProject.Description == nil {
		return types.Project{}, utils.NewBadRequestError(errors.New("at least one field must be updated"))
	}

	project, err := repository.GetProject(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("project with id: %s not found", *id)))
	} else if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	if user.ID != project.CreatorID && !utils.Have(func(i int, adminID string) bool { return adminID == user.ID }, project.AdminIDs) {
		return types.Project{}, utils.NewUnauthorizedError(errors.New("only creator and admins allowed to update project"))
	}

	if updateProject.AdminIDs == nil && updateProject.MembersIDs == nil && updateProject.Description == nil && updateProject.Name == nil {
		return types.Project{}, utils.NewBadRequestError(errors.New("at least one field must be updated"))
	}

	if updateProject.AdminIDs != nil {
		for _, adminID := range *updateProject.AdminIDs {
			if adminID == project.CreatorID || (updateProject.MembersIDs != nil && utils.Have(func(_ int, memberID string) bool {
				return memberID == adminID
			}, *updateProject.MembersIDs)) {
				return types.Project{}, utils.NewBadRequestError(errors.New("userID should be unique for each category"))
			}
			_, err := repository.GetUser(ctx, &adminID, nil)
			if errors.Is(err, pgx.ErrNoRows) {
				return types.Project{}, utils.NewBadRequestError(errors.New(fmt.Sprintf("no user with id: %s found", adminID)))
			} else if err != nil {
				return types.Project{}, utils.NewInternalError(err)
			}
		}
	}

	if updateProject.MembersIDs != nil {
		for _, memberID := range *updateProject.MembersIDs {
			if memberID == project.CreatorID {
				return types.Project{}, utils.NewBadRequestError(errors.New("userID should be unique for each category"))
			}

			_, err := repository.GetUser(ctx, &memberID, nil)
			if errors.Is(err, pgx.ErrNoRows) {
				return types.Project{}, utils.NewBadRequestError(errors.New(fmt.Sprintf("no user with id: %s found", memberID)))
			} else if err != nil {
				return types.Project{}, utils.NewInternalError(err)
			}
		}
	}

	updatedProject, err := repository.UpdateProject(ctx, id, updateProject)

	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	return updatedProject, nil
}

func DeleteProject(user *types.User, projectID *string) (types.Project, error) {
	if user == nil || projectID == nil {
		return types.Project{}, utils.NewBadRequestError(errors.New("user or projectID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project, err := repository.GetProject(ctx, projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("project with id: %s not found", *projectID)))
	} else if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != user.ID {
		return types.Project{}, utils.NewUnauthorizedError(errors.New("only creator can delete project"))
	}

	err = repository.DeleteProject(ctx, projectID)

	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	return project, nil
}

func CreateColumn(user *types.User, createColumn *types.CreateKanbanColumn) (types.KanbanColumn, []types.KanbanColumn, types.Project, error) {
	if user == nil || createColumn == nil {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewBadRequestError(errors.New("user or createColumn is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project, err := repository.GetProject(ctx, &createColumn.ProjectID)

	if errors.Is(err, pgx.ErrNoRows) {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("project with id: %s not found", createColumn.ProjectID)))
	} else if err != nil {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != user.ID && !utils.Have(func(_ int, adminID string) bool { return adminID == user.ID }, project.AdminIDs) {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only creator and admin can create new column"))
	}

	if createColumn.LabelID != nil {
		_, err := repository.GetKanbanColumnLabel(ctx, createColumn.LabelID)

		if errors.Is(err, pgx.ErrNoRows) {
			return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("label with id: %s not found", *createColumn.LabelID)))
		} else if err != nil {
			return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
		}
	}

	columns, err := repository.GetColumns(ctx, &createColumn.ProjectID)
	if err != nil {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	if createColumn.Order < 1 {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewBadRequestError(errors.New("order must be at least 1"))
	}

	if len(columns) == 0 {
		if createColumn.Order != 1 {
			return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewBadRequestError(errors.New("first column must have order 1"))
		}
	} else {
		maxOrder := columns[len(columns)-1].Order
		if createColumn.Order > maxOrder+1 {
			return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewBadRequestError(errors.New("cannot create column with gaps in order"))
		}
	}

	if len(columns) > 0 && createColumn.Order <= len(columns) {
		err = repository.ShiftOrder(ctx, "kanbanColumns", "projectID", &createColumn.ProjectID, createColumn.Order)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
		}
	}

	id := uuid.New().String()

	newColumn, err := repository.CreateKanbanColumn(ctx, &id, &project.ID, createColumn)

	if err != nil {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	updatedColumns, err := repository.GetColumns(ctx, &createColumn.ProjectID)
	if err != nil {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	return newColumn, updatedColumns, project, nil
}

func GetColumn(columnID *string, userID *string) (types.KanbanColumn, error) {
	if columnID == nil || userID == nil {
		return types.KanbanColumn{}, utils.NewBadRequestError(errors.New("columnID or userID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	column, err := repository.GetKanbanColumn(ctx, columnID)

	if errors.Is(err, pgx.ErrNoRows) {
		return types.KanbanColumn{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("column with id: %s not found", *columnID)))
	} else if err != nil {
		return types.KanbanColumn{}, utils.NewInternalError(err)
	}

	project, err := repository.GetProject(ctx, &column.ProjectID)

	if err != nil {
		return types.KanbanColumn{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(_ int, adminID string) bool { return adminID == *userID }, project.AdminIDs) &&
		!utils.Have(func(_ int, memberID string) bool { return memberID == *userID }, project.MembersIDs) {
		return types.KanbanColumn{}, utils.NewUnauthorizedError(errors.New("user not participating in that project"))
	}

	return column, nil
}

func UpdateColumn(columnID *string, user *types.User, updateColumn *types.UpdateKanbanColumn) (types.KanbanColumn, []types.KanbanColumn, types.Project, error) {
	if columnID == nil || user == nil || updateColumn == nil {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewBadRequestError(errors.New("columnID or user or updateColumn is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if updateColumn.DeleteLabel == nil && updateColumn.LabelID == nil && updateColumn.Name == nil && updateColumn.Order == nil {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewBadRequestError(errors.New("nothing to update"))
	}

	project, err := repository.GetProject(ctx, &updateColumn.ProjectID)

	if errors.Is(err, pgx.ErrNoRows) {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("project with id: %s not found", updateColumn.ProjectID)))
	} else if err != nil {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != user.ID && !utils.Have(func(_ int, adminID string) bool { return adminID == user.ID }, project.AdminIDs) {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only creator and admin can update column"))
	}

	column, err := repository.GetKanbanColumn(ctx, columnID)

	if errors.Is(err, pgx.ErrNoRows) {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("column with id: %s not found", *columnID)))
	} else if err != nil {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	if column.ProjectID != project.ID {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewBadRequestError(errors.New("column does not belong to project"))
	}

	if updateColumn.LabelID != nil {
		_, err := repository.GetKanbanColumnLabel(ctx, updateColumn.LabelID)

		if errors.Is(err, pgx.ErrNoRows) {
			return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("label with id: %s not found", *updateColumn.LabelID)))
		} else if err != nil {
			return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
		}
	}

	if updateColumn.Order != nil {
		columns, err := repository.GetColumns(ctx, &updateColumn.ProjectID)
		if err != nil {
			return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
		}

		if *updateColumn.Order < 1 {
			return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewBadRequestError(errors.New("order must be at least 1"))
		}

		if *updateColumn.Order > len(columns) {
			return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewBadRequestError(errors.New("order cannot exceed number of columns"))
		}

		currentOrder := column.Order
		newOrder := *updateColumn.Order

		if currentOrder != newOrder {
			if newOrder > currentOrder {
				err = repository.ShiftOrdersInRange(ctx, "kanbanColumns", "projectID", &updateColumn.ProjectID, currentOrder+1, newOrder, -1)
			} else {
				err = repository.ShiftOrdersInRange(ctx, "kanbanColumns", "projectID", &updateColumn.ProjectID, newOrder, currentOrder-1, 1)
			}
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
			}
		}
	}

	updatedColumn, err := repository.UpdateKanbanColumn(ctx, updateColumn, &column)

	if err != nil {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	updatedColumns, err := repository.GetColumns(ctx, &updateColumn.ProjectID)
	if err != nil {
		return types.KanbanColumn{}, []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	return updatedColumn, updatedColumns, project, nil
}

func DeleteColumn(columnID *string, user *types.User) ([]types.KanbanColumn, types.Project, error) {
	if columnID == nil || user == nil {
		return []types.KanbanColumn{}, types.Project{}, utils.NewBadRequestError(errors.New("columnID or user is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	column, err := repository.GetKanbanColumn(ctx, columnID)

	if errors.Is(err, pgx.ErrNoRows) {
		return []types.KanbanColumn{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("column with id: %s not found", *columnID)))
	} else if err != nil {
		return []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	project, err := repository.GetProject(ctx, &column.ProjectID)

	if err != nil {
		return []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != user.ID && !utils.Have(func(_ int, adminID string) bool { return adminID == user.ID }, project.AdminIDs) {
		return []types.KanbanColumn{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only creator can delete column"))
	}

	err = repository.DeleteKanbanColumn(ctx, columnID)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	columns, err := repository.GetColumns(ctx, &project.ID)
	if err != nil {
		return []types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	return columns, project, nil
}

func GetColumns(projectID *string, userID *string) ([]types.KanbanColumn, error) {
	if projectID == nil || userID == nil {
		return []types.KanbanColumn{}, utils.NewBadRequestError(errors.New("projectID or userID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project, err := repository.GetProject(ctx, projectID)

	if errors.Is(err, pgx.ErrNoRows) {
		return []types.KanbanColumn{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("project with id: %s not found", *projectID)))
	} else if err != nil {
		return []types.KanbanColumn{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) &&
		!utils.Have(func(i int, memberID string) bool { return memberID == *userID }, project.MembersIDs) {
		return []types.KanbanColumn{}, utils.NewUnauthorizedError(errors.New("you cannot fetch a project in which you are not participating"))
	}

	columns, err := repository.GetColumns(ctx, projectID)

	if err != nil {
		return []types.KanbanColumn{}, utils.NewInternalError(err)
	}

	return columns, nil
}

func CreateColumnLabel(userID *string, createColumnLabel *types.CreateKanbanColumnLabel) (types.KanbanColumnLabel, types.Project, error) {
	if userID == nil || createColumnLabel == nil {
		return types.KanbanColumnLabel{}, types.Project{}, utils.NewBadRequestError(errors.New("userID or createColumnLabel is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project, err := repository.GetProject(ctx, &createColumnLabel.ProjectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.KanbanColumnLabel{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("project with id: %s not found", createColumnLabel.ProjectID)))
	} else if err != nil {
		return types.KanbanColumnLabel{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return types.KanbanColumnLabel{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can create column label"))
	}

	id := uuid.New().String()

	label, err := repository.CreateKanbanColumnLabel(ctx, &id, createColumnLabel, nil)
	if err != nil {
		return types.KanbanColumnLabel{}, types.Project{}, utils.NewInternalError(err)
	}

	return label, project, nil
}

func DeleteColumnLabel(userID *string, labelID *string) (types.Project, error) {
	if userID == nil || labelID == nil {
		return types.Project{}, utils.NewBadRequestError(errors.New("userID or labelID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	label, err := repository.GetKanbanColumnLabel(ctx, labelID)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("column with id: %s not found", *labelID)))
	} else if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	project, err := repository.GetProject(ctx, &label.ProjectID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can delete column label"))
	}

	err = repository.DeleteKanbanColumnLabel(ctx, labelID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	return project, nil
}

func UpdateColumnLabel(userID *string, labelID *string, updateLabel *types.UpdateKanbanColumnLabel) (types.KanbanColumnLabel, types.Project, error) {
	if userID == nil || labelID == nil || updateLabel == nil {
		return types.KanbanColumnLabel{}, types.Project{}, utils.NewBadRequestError(errors.New("userID or labelID or updateLabel is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if updateLabel.Name == nil && updateLabel.Color == nil {
		return types.KanbanColumnLabel{}, types.Project{}, utils.NewBadRequestError(errors.New("at least one field must be updated"))
	}

	project, err := repository.GetProject(ctx, &updateLabel.ProjectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.KanbanColumnLabel{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("project with id: %s not found", updateLabel.ProjectID)))
	} else if err != nil {
		return types.KanbanColumnLabel{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return types.KanbanColumnLabel{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can update column label"))
	}

	updatedLabel, err := repository.UpdateKanbanColumnLabel(ctx, labelID, updateLabel)
	if err != nil {
		return types.KanbanColumnLabel{}, types.Project{}, utils.NewInternalError(err)
	}

	return updatedLabel, project, nil
}

func GetColumnLabels(userID *string, projectID *string) ([]types.KanbanColumnLabel, error) {
	if projectID == nil || userID == nil {
		return nil, utils.NewBadRequestError(errors.New("projectID or userID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project, err := repository.GetProject(ctx, projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, utils.NewNotFoundError(errors.New(fmt.Sprintf("project with id: %s not found", *projectID)))
	} else if err != nil {
		return nil, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) &&
		!utils.Have(func(i int, memberID string) bool { return memberID == *userID }, project.MembersIDs) {
		return nil, utils.NewUnauthorizedError(errors.New("you cannot fetch labels in a project in which you are not participating"))
	}

	labels, err := repository.GetKanbanColumnLabels(ctx, projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return []types.KanbanColumnLabel{}, nil
	} else if err != nil {
		return nil, utils.NewInternalError(err)
	}

	return labels, nil
}

func CreateRow(userID *string, createRow *types.CreateKanbanRow) (types.KanbanRow, []types.KanbanRow, types.Project, error) {
	if createRow == nil || userID == nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(errors.New("userID or createRow is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	column, err := repository.GetKanbanColumn(ctx, &createRow.ColumnID)

	if errors.Is(err, pgx.ErrNoRows) {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("column with id: %s not found", createRow.ColumnID)))
	} else if err != nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	if createRow.Priority != nil && *createRow.Priority != types.LowPriority && *createRow.Priority != types.MediumPriority && *createRow.Priority != types.HighPriority {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(errors.New("invalid priority"))
	}

	project, err := repository.GetProject(ctx, &column.ProjectID)

	if err != nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(_ int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can create row"))
	}

	assigneeMap := make(map[string]bool, len(createRow.AssignedUsersIDs))
	for _, assigneeID := range createRow.AssignedUsersIDs {
		if assigneeMap[assigneeID] {
			return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(
				errors.New(fmt.Sprintf("duplicate assignee: %s", assigneeID)))
		}
		assigneeMap[assigneeID] = true

		if !utils.Have(func(_ int, id string) bool { return id == assigneeID },
			append(append(project.MembersIDs, project.AdminIDs...), project.CreatorID)) {
			return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(
				errors.New(fmt.Sprintf("user %s is not a project member", assigneeID)))
		}
	}

	if createRow.LabelID != nil {
		_, err := repository.GetKanbanRowLabel(ctx, createRow.LabelID)

		if errors.Is(err, pgx.ErrNoRows) {
			return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("label with id: %s not found", *createRow.LabelID)))
		} else if err != nil {
			return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
		}
	}

	rows, err := repository.GetRows(ctx, &createRow.ColumnID)
	if err != nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	if createRow.Order < 1 {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(errors.New("order must be at least 1"))
	}

	if len(rows) == 0 {
		if createRow.Order != 1 {
			return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(errors.New("first row must have order 1"))
		}
	} else {
		if createRow.Order > len(rows)+1 {
			return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(errors.New("cannot create row with gaps in order"))
		}
	}

	if len(rows) > 0 && createRow.Order <= len(rows) {
		err = repository.ShiftOrder(ctx, "kanbanRows", "columnID", &createRow.ColumnID, createRow.Order)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
		}
	}

	id := uuid.New().String()
	commentSectionID := uuid.New().String()
	row, err := repository.CreateKanbanRow(ctx, &id, &commentSectionID, userID, createRow)
	if err != nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	err = repository.CreateHistoryPoint(ctx, &types.HistoryPoint{
		ID:     uuid.New().String(),
		RowID:  row.ID,
		UserID: *userID,
		Text:   "created",
	})
	if err != nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	historyPoints, err := repository.GetHistoryPoints(ctx, &row.ID)
	if err != nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	row.History = &historyPoints

	updatedRows, err := repository.GetRows(ctx, &createRow.ColumnID)
	if err != nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	return row, updatedRows, project, nil
}

func UpdateRow(userID *string, rowID *string, updateRow *types.UpdateKanbanRow) (types.KanbanRow, []types.KanbanRow, types.Project, error) {
	if updateRow == nil || userID == nil || rowID == nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(errors.New("userID or rowID or updateRow is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if updateRow.Name == nil && updateRow.LabelID == nil && updateRow.Order == nil &&
		updateRow.DeleteLabel == nil && updateRow.Priority == nil && updateRow.AssignedUsersIDs == nil &&
		updateRow.DueDate == nil && updateRow.Description == nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(errors.New("nothing to update"))
	}

	if updateRow.Priority != nil && *updateRow.Priority != types.LowPriority && *updateRow.Priority != types.MediumPriority && *updateRow.Priority != types.HighPriority {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(errors.New("invalid priority"))
	}

	row, err := repository.GetRow(ctx, rowID)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("row with id: %s not found", *rowID)))
	} else if err != nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	if updateRow.ColumnID != row.ColumnID {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(errors.New("columnID cannot be changed"))
	}

	column, err := repository.GetKanbanColumn(ctx, &updateRow.ColumnID)
	if err != nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	if column.ProjectID != updateRow.ProjectID {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(errors.New("projectID cannot be changed"))
	}

	project, err := repository.GetProject(ctx, &updateRow.ProjectID)
	if err != nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(_ int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can update row"))
	}

	if updateRow.LabelID != nil {
		_, err := repository.GetKanbanRowLabel(ctx, updateRow.LabelID)

		if errors.Is(err, pgx.ErrNoRows) {
			return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("label with id: %s not found", *updateRow.LabelID)))
		} else if err != nil {
			return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
		}
	}

	if updateRow.AssignedUsersIDs != nil {
		assigneeMap := make(map[string]bool, len(*updateRow.AssignedUsersIDs))
		for _, assigneeID := range *updateRow.AssignedUsersIDs {
			if assigneeMap[assigneeID] {
				return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(
					errors.New(fmt.Sprintf("duplicate assignee: %s", assigneeID)))
			}
			assigneeMap[assigneeID] = true

			if !utils.Have(func(_ int, id string) bool { return id == assigneeID },
				append(append(project.MembersIDs, project.AdminIDs...), project.CreatorID)) {
				return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(
					errors.New(fmt.Sprintf("user %s is not a project member", assigneeID)))
			}
		}
	}

	if updateRow.Order != nil {
		rows, err := repository.GetRows(ctx, &updateRow.ColumnID)
		if err != nil {
			return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
		}

		if *updateRow.Order < 1 {
			return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(errors.New("order must be at least 1"))
		}

		if *updateRow.Order > len(rows) {
			return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(errors.New("order cannot exceed number of columns"))
		}

		currentOrder := column.Order
		newOrder := *updateRow.Order

		if currentOrder != newOrder {
			if newOrder > currentOrder {
				err = repository.ShiftOrdersInRange(ctx, "kanbanRows", "columnID", &updateRow.ColumnID, currentOrder+1, newOrder, -1)
			} else {
				err = repository.ShiftOrdersInRange(ctx, "kanbanRows", "columnID", &updateRow.ColumnID, newOrder, currentOrder-1, 1)
			}
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
			}
		}
	}

	updatedRow, err := repository.UpdateKanbanRow(ctx, updateRow, &row)
	if err != nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	err = repository.CreateHistoryPoint(ctx, &types.HistoryPoint{
		ID:     uuid.NewString(),
		RowID:  updatedRow.ID,
		UserID: *userID,
		Text:   "updated",
	})
	if err != nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	historyPoints, err := repository.GetHistoryPoints(ctx, &updatedRow.ID)
	if err != nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	updatedRow.History = &historyPoints

	updatedRows, err := repository.GetRows(ctx, &updateRow.ColumnID)
	if err != nil {
		return types.KanbanRow{}, []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	return updatedRow, updatedRows, project, nil
}

func DeleteRow(userID *string, rowID *string) ([]types.KanbanRow, types.Project, error) {
	if userID == nil || rowID == nil {
		return []types.KanbanRow{}, types.Project{}, utils.NewBadRequestError(errors.New("userID or rowID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row, err := repository.GetRow(ctx, rowID)
	if errors.Is(err, pgx.ErrNoRows) {
		return []types.KanbanRow{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("row with id: %s not found", *rowID)))
	} else if err != nil {
		return []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	column, err := repository.GetKanbanColumn(ctx, &row.ColumnID)
	if err != nil {
		return []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	project, err := repository.GetProject(ctx, &column.ProjectID)
	if err != nil {
		return []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(_ int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return []types.KanbanRow{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can delete row"))
	}

	err = repository.DeleteRow(ctx, rowID)
	if err != nil {
		return []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	rows, err := repository.GetRows(ctx, &row.ColumnID)
	if err != nil {
		return []types.KanbanRow{}, types.Project{}, utils.NewInternalError(err)
	}

	return rows, project, nil
}

func GetRows(userID *string, columnID *string) ([]types.KanbanRow, error) {
	if userID == nil || columnID == nil {
		return []types.KanbanRow{}, utils.NewBadRequestError(errors.New("userID or columnID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	column, err := repository.GetKanbanColumn(ctx, columnID)
	if errors.Is(err, pgx.ErrNoRows) {
		return []types.KanbanRow{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("column with id: %s not found", *columnID)))
	} else if err != nil {
		return []types.KanbanRow{}, utils.NewInternalError(err)
	}

	project, err := repository.GetProject(ctx, &column.ProjectID)
	if err != nil {
		return []types.KanbanRow{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(_ int, adminID string) bool { return adminID == *userID }, project.AdminIDs) &&
		!utils.Have(func(_ int, memberId string) bool { return memberId == *userID }, project.MembersIDs) {
		return []types.KanbanRow{}, utils.NewUnauthorizedError(errors.New("only participants can get rows"))
	}

	rows, err := repository.GetRows(ctx, columnID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return []types.KanbanRow{}, utils.NewInternalError(err)
	}

	return rows, nil
}

func CreateRowLabel(userID *string, createRowLabel *types.CreateKanbanRowLabel) (types.KanbanRowLabel, types.Project, error) {
	if userID == nil || createRowLabel == nil {
		return types.KanbanRowLabel{}, types.Project{}, utils.NewBadRequestError(errors.New("userID or createRowLabel is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project, err := repository.GetProject(ctx, &createRowLabel.ProjectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.KanbanRowLabel{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("project with id: %s not found", createRowLabel.ProjectID)))
	} else if err != nil {
		return types.KanbanRowLabel{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return types.KanbanRowLabel{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can create row label"))
	}

	id := uuid.New().String()

	label, err := repository.CreateKanbanRowLabel(ctx, &id, createRowLabel)
	if err != nil {
		return types.KanbanRowLabel{}, types.Project{}, utils.NewInternalError(err)
	}

	return label, project, nil
}

func DeleteRowLabel(userID *string, labelID *string) (types.Project, error) {
	if userID == nil || labelID == nil {
		return types.Project{}, utils.NewBadRequestError(errors.New("userID or labelID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	label, err := repository.GetKanbanRowLabel(ctx, labelID)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("row with id: %s not found", *labelID)))
	} else if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	project, err := repository.GetProject(ctx, &label.ProjectID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can delete row label"))
	}

	err = repository.DeleteKanbanRowLabel(ctx, labelID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	return project, nil
}

func UpdateRowLabel(userID *string, labelID *string, updateLabel *types.UpdateKanbanRowLabel) (types.KanbanRowLabel, types.Project, error) {
	if userID == nil || labelID == nil || updateLabel == nil {
		return types.KanbanRowLabel{}, types.Project{}, utils.NewBadRequestError(errors.New("userID or labelID or updateLabel is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if updateLabel.Name == nil && updateLabel.Color == nil {
		return types.KanbanRowLabel{}, types.Project{}, utils.NewBadRequestError(errors.New("at least one field must be updated"))
	}

	project, err := repository.GetProject(ctx, &updateLabel.ProjectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.KanbanRowLabel{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("project with id: %s not found", updateLabel.ProjectID)))
	} else if err != nil {
		return types.KanbanRowLabel{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return types.KanbanRowLabel{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can update row label"))
	}

	updatedLabel, err := repository.UpdateKanbanRowLabel(ctx, labelID, updateLabel)
	if err != nil {
		return types.KanbanRowLabel{}, types.Project{}, utils.NewInternalError(err)
	}

	return updatedLabel, project, nil
}

func GetRowLabels(userID *string, projectID *string) ([]types.KanbanRowLabel, error) {
	if projectID == nil || userID == nil {
		return nil, utils.NewBadRequestError(errors.New("projectID or userID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project, err := repository.GetProject(ctx, projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, utils.NewNotFoundError(errors.New(fmt.Sprintf("project with id: %s not found", *projectID)))
	} else if err != nil {
		return nil, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) &&
		!utils.Have(func(i int, memberID string) bool { return memberID == *userID }, project.MembersIDs) {
		return nil, utils.NewUnauthorizedError(errors.New("you cannot fetch labels in a project in which you are not participating"))
	}

	labels, err := repository.GetKanbanRowLabels(ctx, projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return []types.KanbanRowLabel{}, nil
	} else if err != nil {
		return nil, utils.NewInternalError(err)
	}

	return labels, nil
}

func CreateChecklist(userID *string, rowID *string) (types.Checklist, types.Project, error) {
	if userID == nil || rowID == nil {
		return types.Checklist{}, types.Project{}, utils.NewBadRequestError(errors.New("userID or rowID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row, err := repository.GetRow(ctx, rowID)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Checklist{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("row with id: %s not found", *rowID)))
	} else if err != nil {
		return types.Checklist{}, types.Project{}, utils.NewInternalError(err)
	}

	if exists, err := repository.ChecklistExists(ctx, rowID); err != nil {
		return types.Checklist{}, types.Project{}, utils.NewInternalError(err)
	} else if exists {
		return types.Checklist{}, types.Project{}, utils.NewBadRequestError(errors.New("checklist already exists"))
	}

	column, err := repository.GetKanbanColumn(ctx, &row.ColumnID)
	if err != nil {
		return types.Checklist{}, types.Project{}, utils.NewInternalError(err)
	}

	project, err := repository.GetProject(ctx, &column.ProjectID)
	if err != nil {
		return types.Checklist{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return types.Checklist{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can create checklist"))
	}

	checklist, err := repository.CreateChecklist(ctx, &types.Checklist{
		ID:    uuid.New().String(),
		RowID: *rowID,
	})

	if err != nil {
		return types.Checklist{}, types.Project{}, utils.NewInternalError(err)
	}

	return checklist, project, nil
}

func DeleteChecklist(userID *string, checklistID *string) (types.Project, error) {
	if userID == nil || checklistID == nil {
		return types.Project{}, utils.NewBadRequestError(errors.New("userID or checklistID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checklist, err := repository.GetChecklist(ctx, checklistID)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("checklist with id: %s not found", *checklistID)))
	} else if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	row, err := repository.GetRow(ctx, &checklist.RowID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	column, err := repository.GetKanbanColumn(ctx, &row.ColumnID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	project, err := repository.GetProject(ctx, &column.ProjectID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can delete checklist"))
	}

	err = repository.DeleteChecklist(ctx, checklistID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	return project, nil
}

func CreatePoint(userID *string, createPoint *types.CreatePoint) (types.Point, types.Project, error) {
	if userID == nil || createPoint == nil {
		return types.Point{}, types.Project{}, utils.NewBadRequestError(errors.New("userID or createPoint is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checklist, err := repository.GetChecklist(ctx, &createPoint.ChecklistID)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Point{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("checklist with id: %s not found", createPoint.ChecklistID)))
	} else if err != nil {
		return types.Point{}, types.Project{}, utils.NewInternalError(err)
	}

	row, err := repository.GetRow(ctx, &checklist.RowID)
	if err != nil {
		return types.Point{}, types.Project{}, utils.NewInternalError(err)
	}

	column, err := repository.GetKanbanColumn(ctx, &row.ColumnID)
	if err != nil {
		return types.Point{}, types.Project{}, utils.NewInternalError(err)
	}

	project, err := repository.GetProject(ctx, &column.ProjectID)
	if err != nil {
		return types.Point{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return types.Point{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can create point"))
	}

	id := uuid.New().String()

	point, err := repository.CreatePoint(ctx, &id, createPoint)
	if err != nil {
		return types.Point{}, types.Project{}, utils.NewInternalError(err)
	}

	return point, project, nil
}

func UpdatePoint(userID *string, id *string, updatePoint *types.UpdatePoint) (types.Point, types.Project, error) {
	if userID == nil || id == nil || updatePoint == nil {
		return types.Point{}, types.Project{}, utils.NewBadRequestError(errors.New("userID or id or updatePoint is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := repository.GetPoint(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Point{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("point with id: %s not found", *id)))
	} else if err != nil {
		return types.Point{}, types.Project{}, utils.NewInternalError(err)
	}

	project, err := repository.GetProject(ctx, &updatePoint.ProjectID)
	if err != nil {
		return types.Point{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return types.Point{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can update point"))
	}

	point, err := repository.UpdatePoint(ctx, id, updatePoint, false)
	if err != nil {
		return types.Point{}, types.Project{}, utils.NewInternalError(err)
	}

	return point, project, nil
}

func UpdatePointStatus(userID *string, id *string) (types.Point, types.Project, error) {
	if userID == nil || id == nil {
		return types.Point{}, types.Project{}, utils.NewBadRequestError(errors.New("userID or id is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	point, err := repository.GetPoint(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Point{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("point with id: %s not found", *id)))
	} else if err != nil {
		return types.Point{}, types.Project{}, utils.NewInternalError(err)
	}

	checklist, err := repository.GetChecklist(ctx, &point.ChecklistID)
	if err != nil {
		return types.Point{}, types.Project{}, utils.NewInternalError(err)
	}
	row, err := repository.GetRow(ctx, &checklist.RowID)
	if err != nil {
		return types.Point{}, types.Project{}, utils.NewInternalError(err)
	}
	column, err := repository.GetKanbanColumn(ctx, &row.ColumnID)
	if err != nil {
		return types.Point{}, types.Project{}, utils.NewInternalError(err)
	}
	project, err := repository.GetProject(ctx, &column.ProjectID)
	if err != nil {
		return types.Point{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) &&
		!utils.Have(func(i int, memberID string) bool { return memberID == *userID }, project.MembersIDs) {
		return types.Point{}, types.Project{}, utils.NewUnauthorizedError(errors.New("you not participate in this project"))
	}

	var pointInfo types.UpdatePoint
	completed := !point.Completed
	pointInfo.Completed = &completed
	if point.Completed {
		pointInfo.CompletedAt = nil
		pointInfo.CompletedBy = nil
	} else {
		now := time.Now()
		pointInfo.CompletedAt = &now
		pointInfo.CompletedBy = userID
	}

	updatedPoint, err := repository.UpdatePoint(ctx, id, &pointInfo, true)
	if err != nil {
		return types.Point{}, types.Project{}, utils.NewInternalError(err)
	}

	return updatedPoint, project, nil
}

func DeletePoint(userID *string, id *string) (types.Project, error) {
	if userID == nil || id == nil {
		return types.Project{}, utils.NewBadRequestError(errors.New("userID or id is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	point, err := repository.GetPoint(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("point with id: %s not found", *id)))
	} else if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	checklist, err := repository.GetChecklist(ctx, &point.ChecklistID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}
	row, err := repository.GetRow(ctx, &checklist.RowID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}
	column, err := repository.GetKanbanColumn(ctx, &row.ColumnID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}
	project, err := repository.GetProject(ctx, &column.ProjectID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can delete point"))
	}

	err = repository.DeletePoint(ctx, id)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	return project, nil
}

func GetPoints(userID *string, checklistID *string) ([]types.Point, error) {
	if userID == nil || checklistID == nil {
		return []types.Point{}, utils.NewBadRequestError(errors.New("userID or checklistID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checklist, err := repository.GetChecklist(ctx, checklistID)
	if errors.Is(err, pgx.ErrNoRows) {
		return []types.Point{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("checklist with id: %s not found", *checklistID)))
	} else if err != nil {
		return []types.Point{}, utils.NewInternalError(err)
	}
	row, err := repository.GetRow(ctx, &checklist.RowID)
	if err != nil {
		return []types.Point{}, utils.NewInternalError(err)
	}
	column, err := repository.GetKanbanColumn(ctx, &row.ColumnID)
	if err != nil {
		return []types.Point{}, utils.NewInternalError(err)
	}
	project, err := repository.GetProject(ctx, &column.ProjectID)
	if err != nil {
		return []types.Point{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) &&
		!utils.Have(func(i int, memberID string) bool { return memberID == *userID }, project.MembersIDs) {
		return []types.Point{}, utils.NewUnauthorizedError(errors.New("you not participate in this project"))
	}

	points, err := repository.GetPoints(ctx, checklistID)
	if err != nil {
		return []types.Point{}, utils.NewInternalError(err)
	}

	return points, nil
}

func UpdateCanComment(userID *string, commentSectionID *string) (bool, types.Project, error) {
	if userID == nil || commentSectionID == nil {
		return false, types.Project{}, utils.NewBadRequestError(errors.New("userID or commentSectionID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	commentSection, err := repository.GetCommentSection(ctx, commentSectionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("comment section with id: %s not found", *commentSectionID)))
	} else if err != nil {
		return false, types.Project{}, utils.NewInternalError(err)
	}
	row, err := repository.GetRow(ctx, &commentSection.RowID)
	if err != nil {
		return false, types.Project{}, utils.NewInternalError(err)
	}
	column, err := repository.GetKanbanColumn(ctx, &row.ColumnID)
	if err != nil {
		return false, types.Project{}, utils.NewInternalError(err)
	}
	project, err := repository.GetProject(ctx, &column.ProjectID)
	if err != nil {
		return false, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) {
		return false, types.Project{}, utils.NewUnauthorizedError(errors.New("only admins can update comment section"))
	}

	commentSection, err = repository.ChangeCanComment(ctx, commentSectionID)
	if err != nil {
		return false, types.Project{}, utils.NewInternalError(err)
	}

	return commentSection.CanComment, project, nil
}

func GetComments(userID *string, commentSectionID *string) ([]types.Comment, error) {
	if userID == nil || commentSectionID == nil {
		return []types.Comment{}, utils.NewBadRequestError(errors.New("userID or commentSectionID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	commentSection, err := repository.GetCommentSection(ctx, commentSectionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return []types.Comment{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("comment section with id: %s not found", *commentSectionID)))
	} else if err != nil {
		return []types.Comment{}, utils.NewInternalError(err)
	}
	row, err := repository.GetRow(ctx, &commentSection.RowID)
	if err != nil {
		return []types.Comment{}, utils.NewInternalError(err)
	}
	column, err := repository.GetKanbanColumn(ctx, &row.ColumnID)
	if err != nil {
		return []types.Comment{}, utils.NewInternalError(err)
	}
	project, err := repository.GetProject(ctx, &column.ProjectID)
	if err != nil {
		return []types.Comment{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) &&
		!utils.Have(func(i int, memberID string) bool { return memberID == *userID }, project.MembersIDs) {
		return []types.Comment{}, utils.NewUnauthorizedError(errors.New("you not participate in this project"))
	}

	comments, err := repository.GetComments(ctx, commentSectionID)
	if err != nil {
		return []types.Comment{}, utils.NewInternalError(err)
	}

	return comments, nil
}

func CreateComment(userID *string, createComment *types.CreateComment) (types.Comment, types.Project, error) {
	if userID == nil || createComment == nil {
		return types.Comment{}, types.Project{}, utils.NewBadRequestError(errors.New("userID or createComment is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	commentSection, err := repository.GetCommentSection(ctx, &createComment.CommentSectionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Comment{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("comment section with id: %s not found", createComment.CommentSectionID)))
	} else if err != nil {
		return types.Comment{}, types.Project{}, utils.NewInternalError(err)
	}

	if !commentSection.CanComment {
		return types.Comment{}, types.Project{}, utils.NewBadRequestError(errors.New("this comment section can't be commented"))
	}

	row, err := repository.GetRow(ctx, &commentSection.RowID)
	if err != nil {
		return types.Comment{}, types.Project{}, utils.NewInternalError(err)
	}
	column, err := repository.GetKanbanColumn(ctx, &row.ColumnID)
	if err != nil {
		return types.Comment{}, types.Project{}, utils.NewInternalError(err)
	}
	project, err := repository.GetProject(ctx, &column.ProjectID)
	if err != nil {
		return types.Comment{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) &&
		!utils.Have(func(i int, memberID string) bool { return memberID == *userID }, project.MembersIDs) {
		return types.Comment{}, types.Project{}, utils.NewUnauthorizedError(errors.New("you not participate in this project"))
	}

	id := uuid.New().String()
	comment, err := repository.CreateComment(ctx, userID, &id, createComment)
	if err != nil {
		return types.Comment{}, types.Project{}, utils.NewInternalError(err)
	}

	return comment, project, nil
}

func DeleteComment(userID *string, commentID *string) (types.Project, error) {
	if userID == nil || commentID == nil {
		return types.Project{}, utils.NewBadRequestError(errors.New("userID or commentID is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	comment, err := repository.GetComment(ctx, commentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("comment with id: %s not found", *commentID)))
	} else if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}
	commentSection, err := repository.GetCommentSection(ctx, &comment.CommentSectionID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}
	row, err := repository.GetRow(ctx, &commentSection.RowID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}
	column, err := repository.GetKanbanColumn(ctx, &row.ColumnID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}
	project, err := repository.GetProject(ctx, &column.ProjectID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != *userID &&
		!utils.Have(func(i int, adminID string) bool { return adminID == *userID }, project.AdminIDs) &&
		comment.UserID != *userID {
		return types.Project{}, utils.NewUnauthorizedError(errors.New("you cannot delete this comment"))
	}

	err = repository.DeleteComment(ctx, commentID)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	return project, nil
}
