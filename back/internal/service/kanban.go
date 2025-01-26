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
		!utils.Have(func(i int, memberID string) bool { return memberID == *userID }, project.AdminIDs) {
		return types.Project{}, utils.NewUnauthorizedError(errors.New("you cannot fetch a project in which you are not participating"))
	}

	return project, nil
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
