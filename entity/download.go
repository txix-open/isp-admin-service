package entity

const (
	SessionStatusAllowed    = "ALLOWED"
	SessionStatusAllowedRus = "Активная"

	SessionStatusExpired    = "EXPIRED"
	SessionStatusExpiredRus = "Истекшая"

	SessionStatusRevoked    = "REVOKED"
	SessionStatusRevokedRus = "Отмененная"
)

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
