package controller

import (
	"context"
	"msp-admin-service/domain"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UiSettingsService interface {
	SudirSettings(ctx context.Context, req domain.ClientNameRequest) (*domain.UiSudirSetting, error)
}

type UiSettings struct {
	service UiSettingsService
}

func NewUiSettings(service UiSettingsService) UiSettings {
	return UiSettings{
		service: service,
	}
}

// UiSudirSettings
// @Tags ui_settings
// @Summary Получение настроек для ui для авторизации через СУДИР
// @Description Получение настроек для ui для авторизации через СУДИР
// @Accept json
// @Produce json
// @Param body body domain.ClientNameRequest true "Тело запроса"
// @Success 200 {object} domain.UiSudirSetting
// @Failure 400 {object} domain.GrpcError
// @Failure 401 {object} domain.GrpcError "Неизвестное название клиента"
// @Failure 500 {object} domain.GrpcError
// @Router /ui_settings/sudir [POST]
func (c UiSettings) SudirSettings(ctx context.Context, req domain.ClientNameRequest) (*domain.UiSudirSetting, error) {
	resp, err := c.service.SudirSettings(ctx, req)
	switch {
	case errors.Is(err, domain.ErrUnknownClient):
		return nil, status.Error(codes.Unauthenticated, "invalid client name")
	case err != nil:
		return nil, err
	default:
		return resp, nil
	}
}
