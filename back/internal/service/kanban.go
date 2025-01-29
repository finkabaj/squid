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

func CreateProject(userID *string, project *types.CreateProject) (types.Project, error) {
	if userID == nil || project == nil {
		return types.Project{}, utils.NewBadRequestError(errors.New("userID or project is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	for _, adminID := range project.AdminIDs {
		if adminID == *userID || utils.Have(func(_ int, memberID string) bool {
			return memberID == adminID
		}, project.MembersIDs) {
			return types.Project{}, utils.NewBadRequestError(errors.New("userID should be unique for each category"))
		}
		_, err := repository.GetUser(ctx, &adminID, nil)
		if errors.Is(err, pgx.ErrNoRows) {
			return types.Project{}, utils.NewBadRequestError(errors.New(fmt.Sprintf("no user with id: %s found", adminID)))
		} else if err != nil {
			return types.Project{}, utils.NewInternalError(err)
		}
	}

	for _, memberID := range project.MembersIDs {
		if memberID == *userID {
			return types.Project{}, utils.NewBadRequestError(errors.New("userID should be unique for each category"))
		}

		_, err := repository.GetUser(ctx, &memberID, nil)
		if errors.Is(err, pgx.ErrNoRows) {
			return types.Project{}, utils.NewBadRequestError(errors.New(fmt.Sprintf("no user with id: %s found", memberID)))
		} else if err != nil {
			return types.Project{}, utils.NewInternalError(err)
		}
	}

	id := uuid.New().String()

	newProject, err := repository.CreateProject(ctx, &id, userID, project)
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

	updatedProject, err := repository.UpdateProject(ctx, id, &project, updateProject)

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

	err = repository.DeleteProject(ctx, &user.ID, projectID)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return types.Project{}, utils.NewInternalError(err)
	}

	return project, nil
}

func CreateColumn(user *types.User, createColumn *types.CreateKanbanColumn) (types.KanbanColumn, types.Project, error) {
	if user == nil || createColumn == nil {
		return types.KanbanColumn{}, types.Project{}, utils.NewBadRequestError(errors.New("user or createColumn is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project, err := repository.GetProject(ctx, &createColumn.ProjectID)

	if errors.Is(err, pgx.ErrNoRows) {
		return types.KanbanColumn{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("project with id: %s not found", createColumn.ProjectID)))
	} else if err != nil {
		return types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != user.ID && !utils.Have(func(_ int, adminID string) bool { return adminID == user.ID }, project.AdminIDs) {
		return types.KanbanColumn{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only creator and admin can create new column"))
	}

	id := uuid.New().String()

	newColumn, err := repository.CreateKanbanColumn(ctx, &id, &project.ID, createColumn)

	if err != nil {
		return types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	return newColumn, project, nil
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

func UpdateColumn(columnID *string, user *types.User, updateColumn *types.UpdateKanbanColumn) (types.KanbanColumn, types.Project, error) {
	if columnID == nil || user == nil || updateColumn == nil {
		return types.KanbanColumn{}, types.Project{}, utils.NewBadRequestError(errors.New("columnID or user or updateColumn is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	project, err := repository.GetProject(ctx, &updateColumn.ProjectID)

	if errors.Is(err, pgx.ErrNoRows) {
		return types.KanbanColumn{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("project with id: %s not found", updateColumn.ProjectID)))
	} else if err != nil {
		return types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != user.ID && !utils.Have(func(_ int, adminID string) bool { return adminID == user.ID }, project.AdminIDs) {
		return types.KanbanColumn{}, types.Project{}, utils.NewUnauthorizedError(errors.New("only creator and admin can update column"))
	}

	column, err := repository.GetKanbanColumn(ctx, columnID)

	if errors.Is(err, pgx.ErrNoRows) {
		return types.KanbanColumn{}, types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("column with id: %s not found", *columnID)))
	} else if err != nil {
		return types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	updatedColumn, err := repository.UpdateKanbanColumn(ctx, columnID, updateColumn, column)

	if err != nil {
		return types.KanbanColumn{}, types.Project{}, utils.NewInternalError(err)
	}

	return updatedColumn, project, nil
}

func DeleteColumn(columnID *string, user *types.User) (types.Project, error) {
	if columnID == nil || user == nil {
		return types.Project{}, utils.NewBadRequestError(errors.New("columnID or user is nil"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	column, err := repository.GetKanbanColumn(ctx, columnID)

	if errors.Is(err, pgx.ErrNoRows) {
		return types.Project{}, utils.NewNotFoundError(errors.New(fmt.Sprintf("column with id: %s not found", *columnID)))
	} else if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	project, err := repository.GetProject(ctx, &column.ProjectID)

	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	if project.CreatorID != user.ID && !utils.Have(func(_ int, adminID string) bool { return adminID == user.ID }, project.AdminIDs) {
		return types.Project{}, utils.NewUnauthorizedError(errors.New("only creator can delete column"))
	}

	err = repository.DeleteKanbanColumn(ctx, columnID)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return types.Project{}, utils.NewInternalError(err)
	}

	return project, nil
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
