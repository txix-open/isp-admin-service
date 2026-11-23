package domain

import (
	"time"
)

const (
	SessionStatusAllowed    = "ALLOWED"
	SessionStatusAllowedRus = "Активная"

	SessionStatusExpired    = "EXPIRED"
	SessionStatusExpiredRus = "Истекшая"

	SessionStatusRevoked    = "REVOKED"
	SessionStatusRevokedRus = "Отмененная"
)

type Session struct {
	Id        int       // Идентификатор сессии
	UserId    int       // Идентификатор пользоватеккля
	Status    string    // Статус сессии
	ExpiredAt time.Time // Дата протухания сессии
	CreatedAt time.Time // Дата создания сессии
}

type SessionPageRequest struct {
	LimitOffsetParams

	Order *OrderParams  // Параметры сортировки
	Query *SessionQuery // Параметры выборки
}

type SessionQuery struct {
	Id        *int              // Фильтр по идентификатору сессии
	UserId    []int             // Фильтр по идентификаторам пользователей
	Status    []string          // Фильтр по статусам
	CreatedAt *DateFromToParams // Фильтр по дате создания
	ExpiredAt *DateFromToParams // Фильтр по дате протухания
}

type SessionResponse struct {
	TotalCount int       // Количество сессий
	Items      []Session // Список сессий
}

type RevokeRequest struct {
	Id int `validate:"required"` // Идентификатор сессии
}

func (s Session) StatusRus() string {
	switch s.Status {
	case SessionStatusAllowed:
		return SessionStatusAllowedRus
	case SessionStatusExpired:
		return SessionStatusExpiredRus
	case SessionStatusRevoked:
		return SessionStatusRevokedRus
	}
	return ""
}
