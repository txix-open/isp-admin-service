package domain

type AuditDownloadRequest struct {
	LimitOffestParams

	Order *OrderParams
	Query *AuditQuery
}
type SessionDownloadRequest struct {
	LimitOffestParams

	Order *OrderParams
	Query *SessionQuery
}
type UsersDownloadRequest struct {
	LimitOffestParams

	Order *OrderParams
	Query *UserQuery
}
