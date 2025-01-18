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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := uuid.New().String()

	newProject, err := repository.CreateProject(ctx, &id, userID, project)
	if err != nil {
		return types.Project{}, utils.NewInternalError(err)
	}

	return newProject, errors.New("not implemented")
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
