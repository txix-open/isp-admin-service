package controller

import (
	"context"
	"net/http"

	"isp-admin-service/domain"

	"github.com/txix-open/isp-kit/errors"
)

type DownloadService interface {
	DownloadUsers(ctx context.Context, req domain.UsersDownloadRequest, w http.ResponseWriter) error
	DownloadSessions(ctx context.Context, req domain.SessionDownloadRequest, w http.ResponseWriter) error
	DownloadAudit(ctx context.Context, req domain.AuditDownloadRequest, w http.ResponseWriter) error
	DownloadRoles(ctx context.Context, w http.ResponseWriter) error
}

type Download struct {
	service DownloadService
}

func NewDownload(service DownloadService) Download {
	return Download{
		service: service,
	}
}

// Users
// @Tags Download
// @Summary Скачать список пользователей в csv
// @Accept json
// @Produce json
// @Param body body domain.UsersDownloadRequest true "Тело запроса"
// @Success 200
// @Failure 500 {object} apierrors.Error "Внутренняя ошибка сервиса ```{errorCode: -1}```"
// @Router /user/download [POST]
func (c Download) Users(ctx context.Context, req domain.UsersDownloadRequest, w http.ResponseWriter) error {
	err := c.service.DownloadUsers(ctx, req, w)
	if err != nil {
		return errors.WithMessage(err, "download users")
	}

	return nil
}

// Sessions
// @Tags Download
// @Summary Скачать список сессий в csv
// @Accept json
// @Produce json
// @Param body body domain.SessionDownloadRequest true "Тело запроса"
// @Success 200
// @Failure 500 {object} apierrors.Error "Внутренняя ошибка сервиса ```{errorCode: -1}```"
// @Router /session/download [POST]
func (c Download) Sessions(ctx context.Context, req domain.SessionDownloadRequest, w http.ResponseWriter) error {
	err := c.service.DownloadSessions(ctx, req, w)
	if err != nil {
		return errors.WithMessage(err, "download sessions")
	}

	return nil
}

// Audit
// @Tags Download
// @Summary Скачать журнал аудита в csv
// @Accept json
// @Produce json
// @Param body body domain.AuditDownloadRequest true "Тело запроса"
// @Success 200
// @Failure 500 {object} apierrors.Error "Внутренняя ошибка сервиса ```{errorCode: -1}```"
// @Router /log/download [POST]
func (c Download) Audit(ctx context.Context, req domain.AuditDownloadRequest, w http.ResponseWriter) error {
	err := c.service.DownloadAudit(ctx, req, w)
	if err != nil {
		return errors.WithMessage(err, "download audit")
	}

	return nil
}

// Roles
// @Tags Download
// @Summary Скачать матрицу ролей в csv
// @Accept json
// @Produce json
// @Success 200
// @Failure 500 {object} apierrors.Error "Внутренняя ошибка сервиса ```{errorCode: -1}```"
// @Router /role/download [POST]
func (c Download) Roles(ctx context.Context, w http.ResponseWriter) error {
	err := c.service.DownloadRoles(ctx, w)
	if err != nil {
		return errors.WithMessage(err, "download roles")
	}

	return nil
}
