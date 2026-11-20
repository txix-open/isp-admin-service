package controller

import (
	"context"
	"net/http"

	"isp-admin-service/domain"

	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/http/apierrors"
)

type AuditService interface {
	All(ctx context.Context, req domain.AuditPageRequest) (*domain.AuditResponse, error)
	Events(ctx context.Context) ([]domain.AuditEvent, error)
	SetEvents(ctx context.Context, req []domain.SetAuditEvent) error
}

type Audit struct {
	service AuditService
}

func NewAudit(service AuditService) Audit {
	return Audit{
		service: service,
	}
}

// All
// @Tags log
// @Summary Получение списка логов
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body domain.AuditPageRequest true "Тело запроса (пагинация, сортировка, фильтры)"
// @Success 200 {object} domain.AuditResponse
// @Failure 400 {object} apierrors.Error "Невалидное тело запроса"
// @Failure 500 {object} apierrors.Error
// @Router /log/all [POST]
func (c Audit) All(ctx context.Context, req domain.AuditPageRequest) (*domain.AuditResponse, error) {
	if req.Order == nil {
		req.Order = &domain.OrderParams{
			Field: domain.DefaultOrderField,
			Type:  domain.DefaultOrderType,
		}
	}

	return c.service.All(ctx, req)
}

// Events
// @Tags log
// @Summary Получение списка событий аудита
// @Description Возвращает полный список доступных событий аудита
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Success 200 {array} domain.AuditEvent
// @Failure 500 {object} apierrors.Error
// @Router /log/events [POST]
func (c Audit) Events(ctx context.Context) ([]domain.AuditEvent, error) {
	return c.service.Events(ctx)
}

// SetEvents
// @Tags log
// @Summary Установить события аудита
// @Description Устанавливает (обновляет) набор активных событий аудита.
// @Accept json
// @Produce json
// @Param X-AUTH-ADMIN header string true "Токен администратора"
// @Param body body []domain.SetAuditEvent true "Список событий аудита для установки"
// @Success 200 {object} map[string]interface{} "События успешно установлены"
// @Failure 400 {object} apierrors.Error "Неизвестное событие аудита"
// @Failure 500 {object} apierrors.Error
// @Router /log/set_events [POST]
func (c Audit) SetEvents(ctx context.Context, req []domain.SetAuditEvent) error {
	err := c.service.SetEvents(ctx, req)
	switch {
	case errors.As(err, &domain.UnknownAuditEventError{}):
		return apierrors.New(http.StatusBadRequest, domain.UnknownAuditEventErrorCode, err.Error(), err)
	default:
		return err
	}
}
