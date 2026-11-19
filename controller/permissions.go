package controller

import (
	"context"

	"isp-admin-service/domain"
)

type permissionsService interface {
	All(ctx context.Context) []domain.Permission
}

type Permissions struct {
	permissionsService permissionsService
}

func NewPermissions(permissionsService permissionsService) Permissions {
	return Permissions{
		permissionsService: permissionsService,
	}
}

// GetPermissions
// @Tags user
// @Summary Получить все разрешения
// @Description Получить все разрешения
// @Accept json
// @Produce json
// @Success 200 {array} domain.Permission
// @Failure 500 {object} apierrors.Error
// @Router /user/get_permissions [POST]
func (u Permissions) GetPermissions(ctx context.Context) []domain.Permission {
	return u.permissionsService.All(ctx)
}
