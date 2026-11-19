package entity

import "time"

const (
	SessionStatusAllowed    = "ALLOWED"
	SessionStatusAllowedRus = "Активная"

	SessionStatusExpired    = "EXPIRED"
	SessionStatusExpiredRus = "Истекшая"

	SessionStatusRevoked    = "REVOKED"
	SessionStatusRevokedRus = "Отмененная"
)

type AuditResponse struct {
	TotalCount int
	Items      []Audit
}

type Session struct {
	Id        int
	UserId    int
	Status    string
	ExpiredAt time.Time
	CreatedAt time.Time
}

type SessionResponse struct {
	TotalCount int
	Items      []Session
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

type UsersResponse struct {
	Items []User
}

type Permission struct {
	Key  string
	Name string
}

func AuditColumnNames() []string {
	return []string{"Пользователь", "Время события", "Действие"}
}

func SessionColumnNames() []string {
	return []string{"Пользователь", "Идентификатор", "Дата начала сессии", "Дата окончания сессии", "Статус"}
}

func UserColumnNames() []string {
	return []string{"id", "Пользователь", "Описание", "Email", "Роль", "Последняя сессия"}
}

func RoleColumnNames() []string {
	return []string{"Действие", "Ключ в системе"}
}
