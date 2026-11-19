package controller

import (
	"context"
	"net/http"

	"github.com/txix-open/isp-kit/http/apierrors"

	"isp-admin-service/domain"

	"github.com/txix-open/isp-kit/errors"
)

type userService interface {
	GetProfileById(ctx context.Context, userId int64) (*domain.AdminUserShort, error)
	GetUsers(ctx context.Context, req domain.UsersPageRequest) (*domain.UsersResponse, error)
	GetAllUsers(ctx context.Context) (*domain.UsersResponse, error)
	CreateUser(ctx context.Context, req domain.CreateUserRequest, adminId int64) (*domain.User, error)
	UpdateUser(ctx context.Context, req domain.UpdateUserRequest, adminId int64) (*domain.User, error)
	DeleteUsers(ctx context.Context, ids []int64, adminId int64) (int, error)
	Block(ctx context.Context, adminId int64, userId int) error
	GetById(ctx context.Context, userId int) (*domain.User, error)
	ChangePassword(ctx context.Context, adminId int64, oldPassword string, newPassword string) error
}

type User struct {
	userService userService
}

func NewUser(userService userService) User {
	return User{
		userService: userService,
	}
}

// GetProfile
// @Tags user
// @Summary Получить профиль
// @Description Получить данные профиля
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Success 200 {object} domain.AdminUserShort
// @Failure 401 {object} apierrors.Error "Невалидный токен"
// @Failure 404 {object} apierrors.Error "Пользователя не существует"
// @Failure 500 {object} apierrors.Error
// @Router /user/get_profile [POST]
func (u User) GetProfile(ctx context.Context, r *http.Request) (*domain.AdminUserShort, error) {
	adminId, err := getAdminId(r)
	if err != nil {
		return nil, err
	}

	profile, err := u.userService.GetProfileById(ctx, adminId)
	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		return nil, apierrors.New(http.StatusUnauthorized, domain.ErrUnauthenticatedCode, err.Error(), err)
	case errors.Is(err, domain.ErrNotFound):
		return nil, apierrors.New(http.StatusNotFound, domain.ErrNotFoundCode, err.Error(), err)
	case err != nil:
		return nil, errors.WithMessage(err, "get profile")
	default:
		return profile, nil
	}
}

// GetUsers
// @Tags user
// @Summary Список пользователей
// @Description Получить список пользователей
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.UsersPageRequest true "Тело запроса"
// @Success 200 {object} domain.UsersResponse
// @Failure 400 {object} apierrors.Error
// @Failure 500 {object} apierrors.Error
// @Router /user/get_users [POST]
func (u User) GetUsers(ctx context.Context, req domain.UsersPageRequest) (*domain.UsersResponse, error) {
	if req.Order == nil {
		req.Order = &domain.OrderParams{
			Field: domain.DefaultOrderField,
			Type:  domain.DefaultOrderType,
		}
	}

	users, err := u.userService.GetUsers(ctx, req)
	if err != nil {
		return nil, errors.WithMessage(err, "get users")
	}

	return users, nil
}

// GetAllUsers
// @Tags user
// @Summary Список пользователей
// @Description Получить список пользователей
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Success 200 {object} domain.UsersResponse
// @Failure 400 {object} apierrors.Error
// @Failure 500 {object} apierrors.Error
// @Router /user/get_all_users [POST]
func (u User) GetAllUsers(ctx context.Context) (*domain.UsersResponse, error) {
	users, err := u.userService.GetAllUsers(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get all users")
	}

	return users, nil
}

// CreateUser
// @Tags user
// @Summary Создать пользователя
// @Description Создать пользователя
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.CreateUserRequest true "Тело запроса"
// @Success 200 {object} domain.User
// @Failure 400 {object} apierrors.Error "Невалидное тело запроса"
// @Failure 409 {object} apierrors.Error "Пользователь с указанным email уже существует"
// @Failure 500 {object} apierrors.Error
// @Router /user/create_user [POST]
func (u User) CreateUser(ctx context.Context, req domain.CreateUserRequest, r *http.Request) (*domain.User, error) {
	adminId, err := getAdminId(r)
	if err != nil {
		return nil, err
	}

	user, err := u.userService.CreateUser(ctx, req, adminId)
	switch {
	case errors.Is(err, domain.ErrAlreadyExists):
		return nil, apierrors.New(http.StatusConflict, domain.ErrAlreadyExistsCode, err.Error(), err)
	case err != nil:
		return nil, errors.WithMessage(err, "create user")
	default:
		return user, nil
	}
}

// UpdateUser
// @Tags user
// @Summary Обновить пользователя
// @Description Обновить данные существующего пользователя
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.UpdateUserRequest true "Тело запроса"
// @Success 200 {object} domain.User
// @Failure 400 {object} apierrors.Error "Невалидное тело запроса"
// @Failure 404 {object} apierrors.Error "Пользователь с указанным id не существует"
// @Failure 409 {object} apierrors.Error "Пользователь с указанным email уже существует"
// @Failure 500 {object} apierrors.Error
// @Router /user/update_user [POST]
func (u User) UpdateUser(ctx context.Context, req domain.UpdateUserRequest, r *http.Request) (*domain.User, error) {
	adminId, err := getAdminId(r)
	if err != nil {
		return nil, err
	}

	result, err := u.userService.UpdateUser(ctx, req, adminId)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return nil, apierrors.New(http.StatusNotFound, domain.ErrNotFoundCode, err.Error(), err)
	case errors.Is(err, domain.ErrInvalid):
		return nil, apierrors.New(http.StatusBadRequest, domain.ErrInvalidCode, err.Error(), err)
	case errors.Is(err, domain.ErrAlreadyExists):
		return nil, apierrors.New(http.StatusConflict, domain.ErrAlreadyExistsCode, err.Error(), err)
	case err != nil:
		return nil, errors.WithMessage(err, "update user")
	default:
		return result, nil
	}
}

// DeleteUser
// @Tags user
// @Summary Удалить пользователя
// @Description Удалить существующего пользователя
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.IdentitiesRequest true "Тело запроса"
// @Success 200 {object} domain.DeleteResponse
// @Failure 400 {object} apierrors.Error "Невалидное тело запроса"
// @Failure 500 {object} apierrors.Error
// @Router /user/delete_user [POST]
func (u User) DeleteUser(ctx context.Context, identities domain.IdentitiesRequest, r *http.Request) (*domain.DeleteResponse, error) {
	adminId, err := getAdminId(r)
	if err != nil {
		return nil, err
	}

	deletedCount, err := u.userService.DeleteUsers(ctx, identities.Ids, adminId)
	if err != nil {
		return nil, errors.WithMessage(err, "delete")
	}

	return &domain.DeleteResponse{Deleted: deletedCount}, nil
}

// Block
// @Tags user
// @Summary Метод блокировки/разблокировки пользователя
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.IdRequest true "Тело запроса"
// @Success 200
// @Failure 400 {object} apierrors.Error "Невалидное тело запроса"
// @Failure 500 {object} apierrors.Error
// @Router /user/block_user [POST]
func (u User) Block(ctx context.Context, identities domain.IdRequest, r *http.Request) error {
	adminId, err := getAdminId(r)
	if err != nil {
		return err
	}

	err = u.userService.Block(ctx, adminId, identities.UserId)
	if err != nil {
		return errors.WithMessage(err, "block")
	}
	return nil
}

// GetById
// @Tags user
// @Summary Метод получения данных по пользователю
// @Accept json
// @Produce json
// @Param body body domain.IdRequest true "Тело запроса"
// @Success 200 {object} domain.User
// @Failure 400 {object} apierrors.Error "Невалидное тело запроса"
// @Failure 500 {object} apierrors.Error
// @Router /user/get_by_id [POST]
func (u User) GetById(ctx context.Context, identities domain.IdRequest) (*domain.User, error) {
	return u.userService.GetById(ctx, identities.UserId)
}

// ChangePassword
// @Tags user
// @Summary Метод изменения пароля пользователя
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.ChangePasswordRequest true "Тело запроса"
// @Success 200
// @Failure 400 {object} apierrors.Error "Невалидное тело запроса или неверный старый пароль"
// @Failure 500 {object} apierrors.Error "Внутренняя ошибка"
// @Router /user/change_password [POST]
func (u User) ChangePassword(ctx context.Context, req domain.ChangePasswordRequest, r *http.Request) error {
	adminId, err := getAdminId(r)
	if err != nil {
		return err
	}

	err = u.userService.ChangePassword(ctx, adminId, req.OldPassword, req.NewPassword)
	switch {
	case errors.Is(err, domain.ErrInvalidPassword):
		return apierrors.New(http.StatusBadRequest, domain.ErrCodeInvalidPassword, err.Error(), err)
	case err != nil:
		return apierrors.NewInternalServiceError(err)
	default:
		return nil
	}
}
