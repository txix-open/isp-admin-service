package controller

import (
	"context"
	"net/http"
	"strconv"

	"isp-admin-service/auth"
	"isp-admin-service/domain"

	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/http/apierrors"
	"github.com/txix-open/isp-kit/log"
)

type authService interface {
	Login(ctx context.Context, request domain.LoginRequest) (*domain.LoginResponse, error)
	LoginWithOauth(ctx context.Context, request domain.LoginOauthRequest) (*domain.LoginResponse, error)
	Logout(ctx context.Context, adminId int64, request *domain.LogoutRequest) error
}

type Auth struct {
	authService authService
	logger      log.Logger
}

func NewAuth(authService authService, logger log.Logger) Auth {
	return Auth{
		authService: authService,
		logger:      logger,
	}
}

// Logout
// @Tags auth
// @Summary Выход из авторизованной сессии
// @Description Выход из авторизованной сессии администрирования
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Success 200
// @Failure 400 {object} domain.GrpcError "Невалидный токен"
// @Failure 500 {object} domain.GrpcError
// @Router /auth/logout [POST]
func (a Auth) Logout(ctx context.Context, r *http.Request) error {
	adminId, err := getAdminId(r)
	if err != nil {
		return err
	}

	err = a.authService.Logout(ctx, adminId, nil)
	if err != nil {
		return errors.WithMessage(err, "logout")
	}

	return nil
}

// LogoutWithReason
// @Tags auth
// @Summary Выход по бездействию из авторизованной сессии
// @Description Выход по бездействию из авторизованной сессии администрирования
// @Accept json
// @Produce json
// @Param body body domain.LogoutRequest true "Тело запроса"
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Success 200
// @Failure 400 {object} domain.GrpcError "Невалидный токен"
// @Failure 500 {object} domain.GrpcError
// @Router /auth/logout_with_reason [POST]
func (a Auth) LogoutWithReason(ctx context.Context, logoutRequest domain.LogoutRequest, r *http.Request) error {
	adminId, err := getAdminId(r)
	if err != nil {
		return err
	}

	err = a.authService.Logout(ctx, adminId, &logoutRequest)
	if err != nil {
		return errors.WithMessage(err, "logout")
	}

	return nil
}

// Login
// @Tags auth
// @Summary Авторизация по логину и паролю
// @Description Авторизация с получением токена администратора
// @Accept json
// @Produce json
// @Param body body domain.LoginRequest true "Тело запроса"
// @Success 200 {object} domain.LoginResponse
// @Failure 400 {object} domain.GrpcError
// @Failure 401 {object} domain.GrpcError "Данные для авторизации не верны"
// @Failure 500 {object} domain.GrpcError
// @Router /auth/login [POST]
func (a Auth) Login(ctx context.Context, authRequest domain.LoginRequest) (*domain.LoginResponse, error) {
	auth, err := a.authService.Login(ctx, authRequest)

	switch {
	case errors.Is(err, domain.ErrOauthAuthorization):
		return nil, apierrors.New(http.StatusBadRequest, domain.ErrOauthAuthorizationCode, err.Error(), err)
	case errors.Is(err, domain.ErrUnauthenticated):
		a.logger.Error(ctx, err.Error())
		return nil, apierrors.New(http.StatusUnauthorized, domain.ErrUnauthenticatedCode, err.Error(), err)
	case errors.Is(err, domain.ErrTooManyLoginRequests):
		return nil, apierrors.New(http.StatusTooManyRequests, domain.ErrTooManyLoginRequestsCode, err.Error(), err)
	case err != nil:
		return nil, errors.WithMessage(err, "login")
	default:
		return auth, nil
	}
}

// LoginWithOauth
// @Tags auth
// @Summary Авторизация по авторизационному коду от OAuth
// @Description Авторизация с получением токена администратора
// @Accept json
// @Produce json
// @Param body body domain.LoginOauthRequest true "Тело запроса"
// @Success 200 {object} domain.LoginResponse
// @Failure 401 {object} domain.GrpcError "Некорректный код для авторизации или неизвестное название клиента"
// @Failure 412 {object} domain.GrpcError "Авторизация OAuth не настроена на сервере"
// @Failure 500 {object} domain.GrpcError
// @Router /auth/login_with_oauth [POST]
func (a Auth) LoginWithOauth(ctx context.Context, request domain.LoginOauthRequest) (*domain.LoginResponse, error) {
	auth, err := a.authService.LoginWithOauth(ctx, request)

	switch {
	case errors.Is(err, domain.ErrOauthIsMissed):
		return nil, apierrors.New(http.StatusBadRequest, domain.ErrOauthIsMissedCode, err.Error(), err)
	case errors.Is(err, domain.ErrUnauthenticated):
		return nil, apierrors.New(http.StatusUnauthorized, domain.ErrUnauthenticatedCode, err.Error(), err)
	case err != nil:
		return nil, errors.WithMessage(err, "login with oauth")
	default:
		return auth, nil
	}
}

func getAdminId(r *http.Request) (int64, error) {
	token, err := auth.StringFromHeader(domain.AdminAuthIdHeader, r.Header)
	if err != nil {
		return 0, apierrors.NewBusinessError(http.StatusBadRequest, err.Error(), err)
	}

	adminId, err := strconv.Atoi(token)
	if err != nil {
		return 0, apierrors.NewBusinessError(http.StatusBadRequest, err.Error(), err)
	}

	return int64(adminId), nil
}
