package controller

import (
	"context"
	"isp-admin-service/domain"
	"net/http"

	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/http/apierrors"
)

type UiSettingsService interface {
	OauthSettings(ctx context.Context, req domain.ClientNameRequest) (*domain.UiOauthSetting, error)
}

type UiSettings struct {
	service UiSettingsService
}

func NewUiSettings(service UiSettingsService) UiSettings {
	return UiSettings{
		service: service,
	}
}

// UiOauthSettings
// @Tags ui_settings
// @Summary Получение настроек для ui для авторизации через OAuth
// @Description Получение настроек для ui для авторизации через OAuth
// @Accept json
// @Produce json
// @Param body body domain.ClientNameRequest true "Тело запроса"
// @Success 200 {object} domain.UiOauthSetting
// @Failure 400 {object} domain.GrpcError
// @Failure 401 {object} domain.GrpcError "Неизвестное название клиента"
// @Failure 500 {object} domain.GrpcError
// @Router /ui_settings/oauth [POST]
func (c UiSettings) OauthSettings(ctx context.Context, req domain.ClientNameRequest) (*domain.UiOauthSetting, error) {
	resp, err := c.service.OauthSettings(ctx, req)
	switch {
	case errors.Is(err, domain.ErrUnknownClient):
		return nil, apierrors.New(http.StatusUnauthorized, domain.ErrUnknownClientCode, err.Error(), err)
	case err != nil:
		return nil, err
	default:
		return resp, nil
	}
}
