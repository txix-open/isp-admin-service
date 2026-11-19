package domain

type AuditDownloadRequest struct {
	LimitOffsetParams

	Order *OrderParams
	Query *AuditQuery
}
type SessionDownloadRequest struct {
	LimitOffsetParams

	Order *OrderParams
	Query *SessionQuery
}
type UsersDownloadRequest struct {
	LimitOffsetParams

	Order *OrderParams
	Query *UserQuery
}
