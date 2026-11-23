package controller

import (
	"context"
	"net/http"

	"isp-admin-service/domain"

	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/http/apierrors"
)

type roleService interface {
	All(ctx context.Context, req domain.RoleAllRequest) ([]domain.Role, error)
	Create(ctx context.Context, req domain.CreateRoleRequest, adminId int64) (*domain.Role, error)
	Update(ctx context.Context, req domain.UpdateRoleRequest, adminId int64) (*domain.Role, error)
	Delete(ctx context.Context, req domain.DeleteRoleRequest, adminId int64) error
}

type Role struct {
	roleService roleService
}

func NewRole(roleService roleService) Role {
	return Role{
		roleService: roleService,
	}
}

// All
//
//	@Tags			role
//	@Summary		Список ролей
//	@Description	Получить список ролей
//	@Accept			json
//	@Produce		json
//	@Param			X-AUTH-ADMIN	header		string	true	"Токен администратора"
//	@Success		200				{array}		domain.Role
//	@Failure		400				{object}	apierrors.Error
//	@Failure		500				{object}	apierrors.Error
//	@Router			/role/all [POST]
func (u Role) All(ctx context.Context, req domain.RoleAllRequest) ([]domain.Role, error) {
	users, err := u.roleService.All(ctx, req)
	if err != nil {
		return nil, errors.WithMessage(err, "get roles")
	}

	return users, nil
}

// CreateRole
//
//	@Tags			role
//	@Summary		Создать роль
//	@Description	Создать роль
//	@Accept			json
//	@Produce		json
//	@Param			X-AUTH-ADMIN	header		string						true	"Токен администратора"
//	@Param			body			body		domain.CreateRoleRequest	true	"Тело запроса"
//	@Success		200				{object}	domain.Role
//	@Failure		400				{object}	apierrors.Error	"Невалидное тело запроса"
//	@Failure		409				{object}	apierrors.Error	"Роль с указанным именем уже существует"
//	@Failure		500				{object}	apierrors.Error
//	@Router			/role/create [POST]
func (u Role) CreateRole(ctx context.Context, req domain.CreateRoleRequest, r *http.Request) (*domain.Role, error) {
	adminId, err := getAdminId(r)
	if err != nil {
		return nil, err
	}

	role, err := u.roleService.Create(ctx, req, adminId)
	switch {
	case errors.Is(err, domain.ErrAlreadyExists):
		return nil, apierrors.New(http.StatusConflict, domain.ErrAlreadyExistsCode, err.Error(), err)
	case err != nil:
		return nil, errors.WithMessage(err, "create role")
	default:
		return role, nil
	}
}

// UpdateRole
//
//	@Tags			role
//	@Summary		Обновить роль
//	@Description	Обновить данные существующую роль
//	@Accept			json
//	@Produce		json
//	@Param			X-AUTH-ADMIN	header		string						true	"Токен администратора"
//	@Param			body			body		domain.UpdateRoleRequest	true	"Тело запроса"
//	@Success		200				{object}	domain.Role
//	@Failure		404				{object}	apierrors.Error	"Роль с указанным id не существует"
//	@Failure		409				{object}	apierrors.Error	"Роль с указанным именем уже существует"
//	@Failure		500				{object}	apierrors.Error
//	@Router			/role/update [POST]
func (u Role) UpdateRole(ctx context.Context, req domain.UpdateRoleRequest, r *http.Request) (*domain.Role, error) {
	adminId, err := getAdminId(r)
	if err != nil {
		return nil, err
	}

	result, err := u.roleService.Update(ctx, req, adminId)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return nil, apierrors.New(http.StatusNotFound, domain.ErrNotFoundCode, err.Error(), err)
	case errors.Is(err, domain.ErrAlreadyExists):
		return nil, apierrors.New(http.StatusConflict, domain.ErrAlreadyExistsCode, err.Error(), err)
	case err != nil:
		return nil, errors.WithMessage(err, "update role")
	default:
		return result, nil
	}
}

// DeleteRole
//
//	@Tags			role
//	@Summary		Удалить роль
//	@Description	Удалить существующую роль
//	@Accept			json
//	@Produce		json
//	@Param			X-AUTH-ADMIN	header	string						true	"Токен администратора"
//	@Param			body			body	domain.DeleteRoleRequest	true	"Тело запроса"
//	@Success		200
//	@Failure		400	{object}	apierrors.Error	"Невалидное тело запроса"
//	@Failure		500	{object}	apierrors.Error
//	@Router			/role/delete [POST]
func (u Role) DeleteRole(ctx context.Context, req domain.DeleteRoleRequest, r *http.Request) error {
	adminId, err := getAdminId(r)
	if err != nil {
		return err
	}

	err = u.roleService.Delete(ctx, req, adminId)
	if err != nil {
		return errors.WithMessage(err, "delete")
	}

	return nil
}
